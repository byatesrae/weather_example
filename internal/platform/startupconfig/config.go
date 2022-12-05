package startupconfig

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Parser combines flag & environment variable parsing with programmatic "Usage"
// documentation detailing both approaches (as opposed to just flag usage).
type Parser struct {
	// FlagNameToEnvVarName is used to generate an environment variable name from
	// a flag name.
	//
	// If this is not set the default behaviour is to return the value of flagName
	// but as uppercase and with characters "-", "/" & "." replaced with underscores.
	FlagNameToEnvVarName func(flagName string) string
}

// Parse the flags in the flag.FlagSet. This should be used instead of calling the
// target flag.FlagSet's Parse method.
//
// Parse behaves identical to the flag.FlagSet's Parse method except that if a flag
// is not set then Parse will attempt to set that flag's value by looking up the
// value of an environment variable. The environment variable is looked up using
// the output of Parser.FlagNameToEnvVarName.
func (p *Parser) Parse(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse command-line arguments: %w", err)
	}

	flagsSet := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

	var visitAllErr error
	fs.VisitAll(func(f *flag.Flag) {
		if visitAllErr != nil {
			return
		}

		if flagsSet[f.Name] {
			return
		}

		envVarName := p.flagNameToEnvVarName(f.Name)
		if envVarName == "" {
			return
		}

		envVarValue, ok := os.LookupEnv(envVarName)
		if !ok {
			return
		}

		if err := fs.Set(f.Name, envVarValue); err != nil {
			visitAllErr = fmt.Errorf("set flag %q value from environment variable %q: %w", f.Name, envVarName, err)
			return
		}
	})

	if visitAllErr != nil {
		return fmt.Errorf("parse environment variables: %w", visitAllErr)
	}

	return nil
}

// Usage returns a function that should be set on fs's Usage field before calling
// Parser.Parse with fs as an argument.
//
// The behaviour is similar to flag.FlagSet's default Usage, however there is additional
// detail around the environment variables that will be looked up.
func (p *Parser) Usage(fs *flag.FlagSet) func() {
	return func() {
		// See tests for example outputs.

		if fs.Name() == "" {
			fmt.Fprintf(fs.Output(), "Usage:\n")
		} else {
			fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		}

		var errs []error
		fs.VisitAll(func(f *flag.Flag) {
			if err := p.fPrintUsage(fs.Output(), f); err != nil {
				errs = append(errs, err)
			}
		})

		if len(errs) > 0 {
			fmt.Fprintln(fs.Output())

			for _, err := range errs {
				fmt.Fprintln(fs.Output(), err)
			}
		}
	}
}

// FlagError will output a message with err as detail for the flag named flagName
// and then return an error. This is in similar spirit to flag.FlagSet.Usage, where
// usage is also output except in this case only usage for the one flag is output.
//
// This is ideally used to output validation errors after successfully calling Parser.Parse
// with fs as an argument.
func (p *Parser) FlagError(fs *flag.FlagSet, flagName string, err error) error {
	f := fs.Lookup(flagName)
	if f == nil {
		return fmt.Errorf("flag %q does not exist", flagName)
	}

	msg := fmt.Sprintf("error with flag -%s: %v", f.Name, err)
	fmt.Fprintln(fs.Output(), msg)
	p.fPrintUsage(fs.Output(), f)

	return fmt.Errorf(msg)
}

// fPrintUsage output's one flag's usage.
func (p *Parser) fPrintUsage(w io.Writer, f *flag.Flag) error {
	var b strings.Builder

	// output name
	fmt.Fprintf(&b, "  -%s", f.Name)

	flagType, usage := flag.UnquoteUsage(f)

	// output type
	if len(flagType) > 0 {
		b.WriteString(" ")
		b.WriteString(flagType)
	}

	// output usage
	b.WriteString("\n    \t")
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	// output environment variable usage
	envVarName := p.flagNameToEnvVarName(f.Name)
	if envVarName != "" {
		b.WriteString("\n    \tIf not specified, the value will be read from environment variable \"")
		b.WriteString(envVarName)
		b.WriteString("\".")
	}

	flagValueType := getFlagValueType(f)
	isZero, isZeroErr := isZeroValue(f, f.DefValue)

	// output default value
	if flagValueType == flagValueTypeBool || !isZero {
		if flagValueType == flagValueTypeString {
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}

	fmt.Fprint(w, b.String(), "\n")

	return isZeroErr
}

// flagNameToEnvVarName calls p.FlagNameToEnvVarName if set or, if not set, executes
// the default behaviour.
func (p *Parser) flagNameToEnvVarName(flagName string) string {
	if p.FlagNameToEnvVarName != nil {
		return p.FlagNameToEnvVarName(flagName)
	}

	envName := strings.NewReplacer("-", "_", ".", "_", "/", "_").Replace(flagName)

	return strings.ToUpper(envName)
}

// flagValueType is the value type of a flag.Flag.Value.
type flagValueType string

const (
	flagValueTypeUnknown flagValueType = ""
	flagValueTypeBool    flagValueType = "bool"
	flagValueTypeString  flagValueType = "string"
)

func getFlagValueType(f *flag.Flag) flagValueType {
	name, _ := flag.UnquoteUsage(&flag.Flag{
		Usage: strings.ReplaceAll(f.Usage, "`", ""), // See flag.UnquoteUsage implementation.
		Value: f.Value,
	})

	switch name {
	case "":
		return flagValueTypeBool
	case "string":
		return flagValueTypeString
	default:
		return flagValueTypeUnknown
	}
}

// isZeroValue determines whether s represents the zero value for f.Value.
func isZeroValue(f *flag.Flag, s string) (bool, error) {
	fValueType := reflect.TypeOf(f.Value)

	var zeroValue reflect.Value
	if fValueType.Kind() == reflect.Pointer {
		zeroValue = reflect.New(fValueType.Elem())
	} else {
		zeroValue = reflect.Zero(fValueType)
	}

	flagZeroValue, ok := zeroValue.Interface().(flag.Value)
	if !ok {
		// The Value type may itself be an interface type.
		return false, fmt.Errorf("could not determine if flag %q had a zero value", f.Name)
	}

	return s == flagZeroValue.String(), nil
}
