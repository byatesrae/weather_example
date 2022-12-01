package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type variable struct {
	name           string
	envName        string
	flagName       string
	usage          string
	setFromFlag    func()
	setFromEnv     func() (bool, error)
	setFromDefault func()
}

type Variables struct {
	flagset   *flag.FlagSet
	variables []variable
}

func New(flagset *flag.FlagSet) *Variables {
	return &Variables{
		flagset: flagset,
	}
}

func (v *Variables) Parse(commandlineArgs []string) error {
	if err := v.flagset.Parse(commandlineArgs); err != nil {
		return fmt.Errorf("config: parse command-line arguments: %w", err)
	}

	for _, variable := range v.variables {
		set := false
		v.flagset.Visit(func(f *flag.Flag) {
			if variable.flagName == f.Name {
				set = true
			}
		})

		if set {
			variable.setFromFlag()

			continue
		}

		if variable.envName != "" {
			strVal, ok := os.LookupEnv(variable.envName)
			if ok {
				err := variable.setFromEnv()

				*target = envVal

				return true, nil

				set, err := variable.setFromEnv()
				if err != nil {
					return fmt.Errorf("parse environment variable \"%s\": %w", variable.envName, err)
				}
			}
		}

		if set {
			continue
		}

		variable.setFromDefault()
	}

	return nil
}

type AddVarOptions struct {
	envName  string
	flagName string
}

func AddVarWithEnvName(envName string) func(v *AddVarOptions) {
	return func(v *AddVarOptions) {
		v.envName = envName
	}
}

func AddVarWithFlagName(flagName string) func(v *AddVarOptions) {
	return func(v *AddVarOptions) {
		v.flagName = flagName
	}
}

func (v *Variables) AddString(name, defaultValue, usage string, optionOverrides ...func(*AddVarOptions)) *string {
	target := new(string)

	v.AddStringVar(target, name, defaultValue, usage, optionOverrides...)

	return target
}

func (v *Variables) AddStringVar(target *string, name, defaultValue, usage string, optionOverrides ...func(*AddVarOptions)) {
	options := AddVarOptions{}
	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	var flagTarget *string

	if options.flagName != "" {
		flagTarget = v.flagset.String(options.flagName, defaultValue, usage)
	}

	v.variables = append(v.variables, variable{
		name:     name,
		envName:  options.envName,
		flagName: options.flagName,
		usage:    usage,
		setFromFlag: func() {
			if flagTarget != nil {
				*target = *flagTarget
			}
		},
		setFromEnv: func() (bool, error) {
			if options.envName == "" {
				return false, nil
			}

			envVal, ok := os.LookupEnv(options.envName)
			if !ok {
				return false, nil
			}

			*target = envVal

			return true, nil
		},
		setFromDefault: func() {
			*target = defaultValue
		},
	})
}

func (v *Variables) AddInt(name string, defaultValue int, usage string, optionOverrides ...func(*AddVarOptions)) *int {
	target := new(int)

	v.AddIntVar(target, name, defaultValue, usage, optionOverrides...)

	return target
}

func (v *Variables) AddIntVar(target *int, name string, defaultValue int, usage string, optionOverrides ...func(*AddVarOptions)) {
	options := AddVarOptions{}
	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	var flagTarget *int

	if options.flagName != "" {
		flagTarget = v.flagset.Int(options.flagName, defaultValue, usage)
	}

	v.variables = append(v.variables, variable{
		name:     name,
		envName:  options.envName,
		flagName: options.flagName,
		usage:    usage,
		setFromFlag: func() {
			if flagTarget != nil {
				*target = *flagTarget
			}
		},
		setFromEnv: func() (bool, error) {
			if options.envName == "" {
				return false, nil
			}

			envVal, ok := os.LookupEnv(options.envName)
			if !ok {
				return false, nil
			}

			val, err := strconv.Atoi(envVal)
			if err != nil {
				return false, fmt.Errorf("config: convert \"%s\" to integer: %w", envVal, err)
			}

			*target = val

			return true, nil
		},
		setFromDefault: func() {
			*target = defaultValue
		},
	})
}

func (v *Variables) AddDuration(name string, defaultValue time.Duration, usage string, optionOverrides ...func(*AddVarOptions)) *time.Duration {
	target := new(time.Duration)

	v.AddDurationVar(target, name, defaultValue, usage, optionOverrides...)

	return target
}

func (v *Variables) AddDurationVar(target *time.Duration, name string, defaultValue time.Duration, usage string, optionOverrides ...func(*AddVarOptions)) {
	options := AddVarOptions{}
	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	var flagTarget *time.Duration

	if options.flagName != "" {
		flagTarget = v.flagset.Duration(options.flagName, defaultValue, usage)
	}

	v.variables = append(v.variables, variable{
		name:     name,
		envName:  options.envName,
		flagName: options.flagName,
		usage:    usage,
		setFromFlag: func() {
			if flagTarget != nil {
				*target = *flagTarget
			}
		},
		setFromEnv: func() (bool, error) {
			if options.envName == "" {
				return false, nil
			}

			envVal, ok := os.LookupEnv(options.envName)
			if !ok {
				return false, nil
			}

			val, err := time.ParseDuration(envVal)
			if err != nil {
				return false, fmt.Errorf("config: convert \"%s\" to duration: %w", envVal, err)
			}

			*target = val

			return true, nil
		},
		setFromDefault: func() {
			*target = defaultValue
		},
	})
}

func (v *Variables) AddBool(name string, defaultValue bool, usage string, optionOverrides ...func(*AddVarOptions)) *bool {
	target := new(bool)

	v.AddBoolVar(target, name, defaultValue, usage, optionOverrides...)

	return target
}

func (v *Variables) AddBoolVar(target *bool, name string, defaultValue bool, usage string, optionOverrides ...func(*AddVarOptions)) {
	options := AddVarOptions{}
	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	var flagTarget *bool

	if options.flagName != "" {
		flagTarget = v.flagset.Bool(options.flagName, defaultValue, usage)
	}

	v.variables = append(v.variables, variable{
		name:     name,
		envName:  options.envName,
		flagName: options.flagName,
		usage:    usage,
		setFromFlag: func() {
			if flagTarget != nil {
				*target = *flagTarget
			}
		},
		setFromEnv: func() (bool, error) {
			if options.envName == "" {
				return false, nil
			}

			envVal, ok := os.LookupEnv(options.envName)
			if !ok {
				return false, nil
			}

			val, err := strconv.ParseBool(envVal)
			if err != nil {
				return false, fmt.Errorf("config: convert \"%s\" to bool: %w", envVal, err)
			}

			*target = val

			return true, nil
		},
		setFromDefault: func() {
			*target = defaultValue
		},
	})
}
