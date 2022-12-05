package startupconfig

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Parser struct {
	Fs                   *flag.FlagSet
	FlagNameToEnvVarName func(flagName string) string
}

func (p *Parser) Parse(args []string) error {
	if err := p.Fs.Parse(args); err != nil {
		return fmt.Errorf("parse command-line arguments: %w", err)
	}

	flagsSet := map[string]bool{}
	p.Fs.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

	var visitAllErr error
	p.Fs.VisitAll(func(f *flag.Flag) {
		if visitAllErr != nil {
			return
		}

		if flagsSet[f.Name] {
			return
		}

		envVarName := p.flagNameToEnvVarName(f.Name)

		envVarValue, ok := os.LookupEnv(envVarName)
		if !ok {
			return
		}

		if err := p.Fs.Set(f.Name, envVarValue); err != nil {
			visitAllErr = fmt.Errorf("set flag %q value from environment variable %q: %w", f.Name, envVarName, err)
			return
		}
	})

	if visitAllErr != nil {
		return fmt.Errorf("parse environment variables: %w", visitAllErr)
	}

	return nil
}

func (p *Parser) Usage() {
	// See tests for example outputs.

	if p.Fs.Name() == "" {
		fmt.Fprintf(p.Fs.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(p.Fs.Output(), "Usage of %s:\n", p.Fs.Name())
	}

	var errs []error
	p.Fs.VisitAll(func(f *flag.Flag) {
		if err := p.singleUsage(f); err != nil {
			errs = append(errs, err)
		}
	})

	if len(errs) > 0 {
		fmt.Fprintln(p.Fs.Output())

		for _, err := range errs {
			fmt.Fprintln(p.Fs.Output(), err)
		}
	}
}

func (p *Parser) FlagError(flagName string, err error) error {
	f := p.Fs.Lookup(flagName)
	if f == nil {
		return fmt.Errorf("flag %q does not exist", flagName)
	}

	msg := fmt.Sprintf("error with flag -%s: %v", f.Name, err)
	fmt.Fprintln(p.Fs.Output(), msg)
	p.singleUsage(f)

	return fmt.Errorf(msg)
}

func (p *Parser) singleUsage(f *flag.Flag) error {
	var b strings.Builder
	fmt.Fprintf(&b, "  -%s", f.Name)
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString(" ")
		b.WriteString(name)
	}
	b.WriteString("\n    \t")
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	envVarName := p.flagNameToEnvVarName(f.Name)
	if envVarName != "" {
		b.WriteString("\n    \tIf not specified, the value will be read from environment variable \"")
		b.WriteString(envVarName)
		b.WriteString("\".")
	}

	flagType := getFlagType(f)
	isZero, isZeroErr := isZeroValue(f, f.DefValue)

	if flagType == flagTypeBool || !isZero {
		if flagType == flagTypeString {
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}

	fmt.Fprint(p.Fs.Output(), b.String(), "\n")

	return isZeroErr
}

func (p *Parser) flagNameToEnvVarName(flagName string) string {
	if p.FlagNameToEnvVarName != nil {
		return p.FlagNameToEnvVarName(flagName)
	}

	envName := strings.NewReplacer("-", "_", ".", "_", "/", "_").Replace(flagName)

	return strings.ToUpper(envName)
}

type flagType string

const (
	flagTypeUnknown flagType = ""
	flagTypeBool    flagType = "bool"
	flagTypeString  flagType = "string"
)

func getFlagType(f *flag.Flag) flagType {
	_, unquotedUsage := flag.UnquoteUsage(f)
	typeStr, _ := flag.UnquoteUsage(&flag.Flag{Usage: unquotedUsage, Value: f.Value})

	switch typeStr {
	case "":
		return flagTypeBool
	case "string":
		return flagTypeString
	default:
		return flagTypeUnknown
	}
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(f *flag.Flag, value string) (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, f.Name, e)
		}
	}()
	return value == z.Interface().(flag.Value).String(), nil
}
