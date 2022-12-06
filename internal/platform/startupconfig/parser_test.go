package startupconfig

import (
	"bytes"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParserParse(t *testing.T) {
	t.Parallel()

	// values is just a bag of values, a common type used in testing the parsing.
	type values struct {
		firstBool     bool
		firstInt      int
		firstInt64    int64
		firstUint     uint
		firstUint64   uint64
		firstString   string
		firstFloat64  float64
		firstDuration time.Duration
		firstText     text
		firstFunc     string
	}

	for _, tc := range []struct {
		name    string
		withEnv map[string]string
		with    *Parser

		// Create a FlagSet, optionally binding its flags to the variable addresses
		// in v, to be compared for equality to "expected" below.
		giveFs         func(v *values) *flag.FlagSet
		giveArgs       []string
		expected       *values
		expectedOutput string
		expectedErr    string
	}{
		{
			name:    "all_types_args",
			withEnv: map[string]string{},
			with:    &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.BoolVar(&v.firstBool, "first-bool", false, "")
				fs.IntVar(&v.firstInt, "first-int", 0, "")
				fs.Int64Var(&v.firstInt64, "first-int64", 0, "")
				fs.UintVar(&v.firstUint, "first-uint", 0, "")
				fs.Uint64Var(&v.firstUint64, "first-uint64", 0, "")
				fs.StringVar(&v.firstString, "first-string", "", "")
				fs.Float64Var(&v.firstFloat64, "first-float64", 0, "")
				fs.DurationVar(&v.firstDuration, "first-duration", 0, "")
				fs.TextVar(&v.firstText, "first-text", &text{}, "")
				fs.Func("first-func", "", func(s string) error { v.firstFunc = s; return nil })

				return fs
			},
			giveArgs: []string{
				"-first-bool=true",
				"-first-int=1111",
				"-first-int64=2222",
				"-first-uint=3333",
				"-first-uint64=4444",
				"-first-string=aaaa",
				"-first-float64=555.5",
				"-first-duration=6666s",
				"-first-text=bbbb",
				"-first-func=cccc",
			},
			expected: &values{
				firstBool:     true,
				firstInt:      1111,
				firstInt64:    2222,
				firstUint:     3333,
				firstUint64:   4444,
				firstString:   "aaaa",
				firstFloat64:  555.5,
				firstDuration: time.Second * 6666,
				firstText:     text{b: []byte("bbbb")},
				firstFunc:     "cccc",
			},
			expectedErr: "",
		},
		{
			name: "all_types_env",
			withEnv: map[string]string{
				"FIRST_BOOL":     "TRUE",
				"FIRST_INT":      "1111",
				"FIRST_INT64":    "2222",
				"FIRST_UINT":     "3333",
				"FIRST_UINT64":   "4444",
				"FIRST_STRING":   "aaaa",
				"FIRST_FLOAT64":  "555.5",
				"FIRST_DURATION": "6666s",
				"FIRST_TEXT":     "bbbb",
				"FIRST_FUNC":     "cccc",
			},
			with: &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.BoolVar(&v.firstBool, "first-bool", false, "")
				fs.IntVar(&v.firstInt, "first-int", 0, "")
				fs.Int64Var(&v.firstInt64, "first-int64", 0, "")
				fs.UintVar(&v.firstUint, "first-uint", 0, "")
				fs.Uint64Var(&v.firstUint64, "first-uint64", 0, "")
				fs.StringVar(&v.firstString, "first-string", "", "")
				fs.Float64Var(&v.firstFloat64, "first-float64", 0, "")
				fs.DurationVar(&v.firstDuration, "first-duration", 0, "")
				fs.TextVar(&v.firstText, "first-text", &text{}, "")
				fs.Func("first-func", "", func(s string) error { v.firstFunc = s; return nil })

				return fs
			},
			giveArgs: []string{},
			expected: &values{
				firstBool:     true,
				firstInt:      1111,
				firstInt64:    2222,
				firstUint:     3333,
				firstUint64:   4444,
				firstString:   "aaaa",
				firstFloat64:  555.5,
				firstDuration: time.Second * 6666,
				firstText:     text{b: []byte("bbbb")},
				firstFunc:     "cccc",
			},
			expectedErr: "",
		},
		{
			name: "all_types_arg_and_env",
			withEnv: map[string]string{
				"FIRST_BOOL":     "TRUE",
				"FIRST_INT":      "7777",
				"FIRST_INT64":    "8888",
				"FIRST_UINT":     "9999",
				"FIRST_UINT64":   "11111",
				"FIRST_STRING":   "AAAA",
				"FIRST_FLOAT64":  "2222.2",
				"FIRST_DURATION": "33333",
				"FIRST_TEXT":     "BBBB",
				"FIRST_FUNC":     "DDDD",
			},
			with: &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.BoolVar(&v.firstBool, "first-bool", false, "")
				fs.IntVar(&v.firstInt, "first-int", 0, "")
				fs.Int64Var(&v.firstInt64, "first-int64", 0, "")
				fs.UintVar(&v.firstUint, "first-uint", 0, "")
				fs.Uint64Var(&v.firstUint64, "first-uint64", 0, "")
				fs.StringVar(&v.firstString, "first-string", "", "")
				fs.Float64Var(&v.firstFloat64, "first-float64", 0, "")
				fs.DurationVar(&v.firstDuration, "first-duration", 0, "")
				fs.TextVar(&v.firstText, "first-text", &text{}, "")
				fs.Func("first-func", "", func(s string) error { v.firstFunc = s; return nil })

				return fs
			},
			giveArgs: []string{
				"-first-uint64=4444",
				"-first-string=aaaa",
				"-first-float64=555.5",
				"-first-duration=6666s",
				"-first-text=bbbb",
				"-first-func=cccc",
			},
			expected: &values{
				firstBool:     true,
				firstInt:      7777,
				firstInt64:    8888,
				firstUint:     9999,
				firstUint64:   4444,
				firstString:   "aaaa",
				firstFloat64:  555.5,
				firstDuration: time.Second * 6666,
				firstText:     text{b: []byte("bbbb")},
				firstFunc:     "cccc",
			},
			expectedErr: "",
		},
		{
			name:    "custom_env_name",
			withEnv: map[string]string{"CUSTOM_ENV": "true"},
			with:    &Parser{FlagNameToEnvVarName: func(flagName string) string { return "CUSTOM_ENV" }},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.BoolVar(&v.firstBool, "first-bool", false, "")

				return fs
			},
			giveArgs: []string{},
			expected: &values{firstBool: true},
		},
		{
			name:    "ignore_env",
			withEnv: map[string]string{"FIRST_BOOL": "true"},
			with:    &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.BoolVar(&v.firstBool, "first-bool", false, "")

				return fs
			},
			giveArgs: []string{},
			expected: &values{firstBool: false},
		},
		{
			name:    "nothing_set",
			withEnv: map[string]string{},
			with:    &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.Bool("first-bool", false, "")

				return fs
			},
			giveArgs:    []string{},
			expected:    &values{},
			expectedErr: "",
		},
		{
			name:    "env_error",
			withEnv: map[string]string{"FIRST_BOOL": "WOOPS", "FIRST_STRING": "aaaa"},
			with:    &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.Bool("first-bool", false, "")
				fs.String("first-string", "", "")

				return fs
			},
			giveArgs:    []string{},
			expected:    &values{},
			expectedErr: "parse environment variables: set flag \"first-bool\" value from environment variable \"FIRST_BOOL\": parse error",
		},
		{
			name:    "arg_error",
			withEnv: map[string]string{},
			with:    &Parser{},
			giveFs: func(v *values) *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)
				fs.Bool("first-bool", false, "")

				return fs
			},
			giveArgs:       []string{"-first-bool=woops"},
			expected:       &values{},
			expectedOutput: "invalid boolean value \"woops\" for -first-bool: parse error\nUsage:\n  -first-bool\n    \tIf not specified, the value will be read from environment variable \"FIRST_BOOL\". (default false)\n",
			expectedErr:    "parse command-line arguments: invalid boolean value \"woops\" for -first-bool: parse error",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			for k, v := range tc.withEnv {
				t.Setenv(k, v)
			}

			fsValues := &values{}

			fs := tc.giveFs(fsValues)

			output := bytes.NewBufferString("")
			fs.SetOutput(output)
			fs.Usage = tc.with.Usage(fs)

			// Do
			err := tc.with.Parse(fs, tc.giveArgs)

			// Assert
			assert.Equal(t, tc.expected, fsValues, "Values")
			assert.Equal(t, tc.expectedOutput, output.String(), "FlagSet Output")

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr, "Parse err")
			} else {
				assert.NoError(t, err, "Parse err")
			}
		})
	}
}

func TestParserUsage(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		with     *Parser
		give     *flag.FlagSet
		expected string
	}{
		{
			name: "all_types_all_usages_all_defaults_all_env",
			with: &Parser{},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", true, "First bool usage.")
				fs.Int("first-int", 111, "First int usage.")
				fs.Int64("first-int64", 222, "First int64 usage.")
				fs.Uint("first-uint", 333, "First uint usage.")
				fs.Uint64("first-uint64", 444, "First uint64 usage.")
				fs.String("first-string", "aaa", "First string usage.")
				fs.Float64("first-float64", 555, "First float64 usage.")
				fs.Duration("first-duration", time.Second*666, "First duration usage.")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("bbb")}, "First text usage.")
				fs.Func("first-func", "First func usage.", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tFirst bool usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_BOOL\". (default true)\n  -first-duration duration\n    \tFirst duration usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_DURATION\". (default 11m6s)\n  -first-float64 float\n    \tFirst float64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_FLOAT64\". (default 555)\n  -first-func value\n    \tFirst func usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_FUNC\".\n  -first-int int\n    \tFirst int usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT\". (default 111)\n  -first-int64 int\n    \tFirst int64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT64\". (default 222)\n  -first-string string\n    \tFirst string usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_STRING\". (default \"aaa\")\n  -first-text value\n    \tFirst text usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_TEXT\". (default bbb)\n  -first-uint uint\n    \tFirst uint usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT\". (default 333)\n  -first-uint64 uint\n    \tFirst uint64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT64\". (default 444)\n",
		},
		{
			name: "all_types_all_usages_all_defaults_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", true, "First bool usage.")
				fs.Int("first-int", 111, "First int usage.")
				fs.Int64("first-int64", 222, "First int64 usage.")
				fs.Uint("first-uint", 333, "First uint usage.")
				fs.Uint64("first-uint64", 444, "First uint64 usage.")
				fs.String("first-string", "aaa", "First string usage.")
				fs.Float64("first-float64", 555, "First float64 usage.")
				fs.Duration("first-duration", time.Second*666, "First duration usage.")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("bbb")}, "First text usage.")
				fs.Func("first-func", "First func usage.", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tFirst bool usage. (default true)\n  -first-duration duration\n    \tFirst duration usage. (default 11m6s)\n  -first-float64 float\n    \tFirst float64 usage. (default 555)\n  -first-func value\n    \tFirst func usage.\n  -first-int int\n    \tFirst int usage. (default 111)\n  -first-int64 int\n    \tFirst int64 usage. (default 222)\n  -first-string string\n    \tFirst string usage. (default \"aaa\")\n  -first-text value\n    \tFirst text usage. (default bbb)\n  -first-uint uint\n    \tFirst uint usage. (default 333)\n  -first-uint64 uint\n    \tFirst uint64 usage. (default 444)\n",
		},
		{
			name: "all_types_all_usages_no_defaults_all_env",
			with: &Parser{},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", false, "First bool usage.")
				fs.Int("first-int", 0, "First int usage.")
				fs.Int64("first-int64", 0, "First int64 usage.")
				fs.Uint("first-uint", 0, "First uint usage.")
				fs.Uint64("first-uint64", 0, "First uint64 usage.")
				fs.String("first-string", "", "First string usage.")
				fs.Float64("first-float64", 0, "First float64 usage.")
				fs.Duration("first-duration", 0, "First duration usage.")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("")}, "First text usage.")
				fs.Func("first-func", "First func usage.", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tFirst bool usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_BOOL\". (default false)\n  -first-duration duration\n    \tFirst duration usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_DURATION\".\n  -first-float64 float\n    \tFirst float64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_FLOAT64\".\n  -first-func value\n    \tFirst func usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_FUNC\".\n  -first-int int\n    \tFirst int usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT\".\n  -first-int64 int\n    \tFirst int64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT64\".\n  -first-string string\n    \tFirst string usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_STRING\".\n  -first-text value\n    \tFirst text usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_TEXT\".\n  -first-uint uint\n    \tFirst uint usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT\".\n  -first-uint64 uint\n    \tFirst uint64 usage.\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT64\".\n",
		},
		{
			name: "all_types_all_usages_no_defaults_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", false, "First bool usage.")
				fs.Int("first-int", 0, "First int usage.")
				fs.Int64("first-int64", 0, "First int64 usage.")
				fs.Uint("first-uint", 0, "First uint usage.")
				fs.Uint64("first-uint64", 0, "First uint64 usage.")
				fs.String("first-string", "", "First string usage.")
				fs.Float64("first-float64", 0, "First float64 usage.")
				fs.Duration("first-duration", 0, "First duration usage.")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("")}, "First text usage.")
				fs.Func("first-func", "First func usage.", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tFirst bool usage. (default false)\n  -first-duration duration\n    \tFirst duration usage.\n  -first-float64 float\n    \tFirst float64 usage.\n  -first-func value\n    \tFirst func usage.\n  -first-int int\n    \tFirst int usage.\n  -first-int64 int\n    \tFirst int64 usage.\n  -first-string string\n    \tFirst string usage.\n  -first-text value\n    \tFirst text usage.\n  -first-uint uint\n    \tFirst uint usage.\n  -first-uint64 uint\n    \tFirst uint64 usage.\n",
		},
		{
			name: "all_types_no_usages_all_defaults_all_env",
			with: &Parser{},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", true, "")
				fs.Int("first-int", 111, "")
				fs.Int64("first-int64", 222, "")
				fs.Uint("first-uint", 333, "")
				fs.Uint64("first-uint64", 444, "")
				fs.String("first-string", "aaa", "")
				fs.Float64("first-float64", 555, "")
				fs.Duration("first-duration", time.Second*666, "")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("bbb")}, "")
				fs.Func("first-func", "", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tIf not specified, the value will be read from environment variable \"FIRST_BOOL\". (default true)\n  -first-duration duration\n    \tIf not specified, the value will be read from environment variable \"FIRST_DURATION\". (default 11m6s)\n  -first-float64 float\n    \tIf not specified, the value will be read from environment variable \"FIRST_FLOAT64\". (default 555)\n  -first-func value\n    \tIf not specified, the value will be read from environment variable \"FIRST_FUNC\".\n  -first-int int\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT\". (default 111)\n  -first-int64 int\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT64\". (default 222)\n  -first-string string\n    \tIf not specified, the value will be read from environment variable \"FIRST_STRING\". (default \"aaa\")\n  -first-text value\n    \tIf not specified, the value will be read from environment variable \"FIRST_TEXT\". (default bbb)\n  -first-uint uint\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT\". (default 333)\n  -first-uint64 uint\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT64\". (default 444)\n",
		},
		{
			name: "all_types_no_usages_all_defaults_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", true, "")
				fs.Int("first-int", 111, "")
				fs.Int64("first-int64", 222, "")
				fs.Uint("first-uint", 333, "")
				fs.Uint64("first-uint64", 444, "")
				fs.String("first-string", "aaa", "")
				fs.Float64("first-float64", 555, "")
				fs.Duration("first-duration", time.Second*666, "")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("bbb")}, "")
				fs.Func("first-func", "", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool (default true)\n  -first-duration duration (default 11m6s)\n  -first-float64 float (default 555)\n  -first-func value\n  -first-int int (default 111)\n  -first-int64 int (default 222)\n  -first-string string (default \"aaa\")\n  -first-text value (default bbb)\n  -first-uint uint (default 333)\n  -first-uint64 uint (default 444)\n",
		},
		{
			name: "all_types_no_usages_no_defaults_all_env",
			with: &Parser{},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", false, "")
				fs.Int("first-int", 0, "")
				fs.Int64("first-int64", 0, "")
				fs.Uint("first-uint", 0, "")
				fs.Uint64("first-uint64", 0, "")
				fs.String("first-string", "", "")
				fs.Float64("first-float64", 0, "")
				fs.Duration("first-duration", 0, "")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("")}, "")
				fs.Func("first-func", "", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool\n    \tIf not specified, the value will be read from environment variable \"FIRST_BOOL\". (default false)\n  -first-duration duration\n    \tIf not specified, the value will be read from environment variable \"FIRST_DURATION\".\n  -first-float64 float\n    \tIf not specified, the value will be read from environment variable \"FIRST_FLOAT64\".\n  -first-func value\n    \tIf not specified, the value will be read from environment variable \"FIRST_FUNC\".\n  -first-int int\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT\".\n  -first-int64 int\n    \tIf not specified, the value will be read from environment variable \"FIRST_INT64\".\n  -first-string string\n    \tIf not specified, the value will be read from environment variable \"FIRST_STRING\".\n  -first-text value\n    \tIf not specified, the value will be read from environment variable \"FIRST_TEXT\".\n  -first-uint uint\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT\".\n  -first-uint64 uint\n    \tIf not specified, the value will be read from environment variable \"FIRST_UINT64\".\n",
		},
		{
			name: "all_types_no_usages_no_defaults_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("test_flagset", flag.ContinueOnError)
				fs.Bool("first-bool", false, "")
				fs.Int("first-int", 0, "")
				fs.Int64("first-int64", 0, "")
				fs.Uint("first-uint", 0, "")
				fs.Uint64("first-uint64", 0, "")
				fs.String("first-string", "", "")
				fs.Float64("first-float64", 0, "")
				fs.Duration("first-duration", 0, "")
				fs.TextVar(&text{}, "first-text", &text{b: []byte("")}, "")
				fs.Func("first-func", "", func(s string) error { return nil })

				return fs
			}(),
			expected: "Usage of test_flagset:\n  -first-bool (default false)\n  -first-duration duration\n  -first-float64 float\n  -first-func value\n  -first-int int\n  -first-int64 int\n  -first-string string\n  -first-text value\n  -first-uint uint\n  -first-uint64 uint\n",
		},
		{
			name: "no_title_no_flags",
			with: &Parser{},
			give: func() *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)

				return fs
			}(),
			expected: "Usage:\n",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := bytes.NewBufferString("")
			tc.give.SetOutput(output)

			tc.with.Usage(tc.give)()

			assert.Equal(t, tc.expected, output.String(), "Output")
		})
	}
}

func TestParserFlagError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		with           *Parser
		giveFs         *flag.FlagSet
		giveFlagName   string
		giveErr        error
		expectedOutput string
		expectedErr    string
	}{
		{
			name: "int_err_with_default_with_usage_with_env",
			with: &Parser{},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", true, "Test bool flag usage.")
				fs.Int("test-int-flag", 123, "Test int flag usage.")
				fs.String("test-string-flag", "abc", "Test string flag usage.")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tTest int flag usage.\n    \tIf not specified, the value will be read from environment variable \"TEST_INT_FLAG\". (default 123)\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_with_default_with_usage_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", true, "Test bool flag usage.")
				fs.Int("test-int-flag", 123, "Test int flag usage.")
				fs.String("test-string-flag", "abc", "Test string flag usage.")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tTest int flag usage. (default 123)\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_with_default_no_usage_with_env",
			with: &Parser{},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", true, "")
				fs.Int("test-int-flag", 123, "")
				fs.String("test-string-flag", "abc", "")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tIf not specified, the value will be read from environment variable \"TEST_INT_FLAG\". (default 123)\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_with_default_no_usage_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", true, "")
				fs.Int("test-int-flag", 123, "")
				fs.String("test-string-flag", "abc", "")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int (default 123)\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_no_default_with_usage_with_env",
			with: &Parser{},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", false, "Test bool flag usage.")
				fs.Int("test-int-flag", 0, "Test int flag usage.")
				fs.String("test-string-flag", "", "Test string flag usage.")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tTest int flag usage.\n    \tIf not specified, the value will be read from environment variable \"TEST_INT_FLAG\".\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_no_default_with_usage_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", false, "Test bool flag usage.")
				fs.Int("test-int-flag", 0, "Test int flag usage.")
				fs.String("test-string-flag", "", "Test string flag usage.")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tTest int flag usage.\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_no_default_no_usage_with_env",
			with: &Parser{},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", false, "")
				fs.Int("test-int-flag", 0, "")
				fs.String("test-string-flag", "", "")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n    \tIf not specified, the value will be read from environment variable \"TEST_INT_FLAG\".\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "int_err_no_default_no_usage_no_env",
			with: &Parser{FlagNameToEnvVarName: func(flagName string) string { return "" }},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("testflags", flag.ContinueOnError)

				fs.Bool("test-bool-flag", false, "")
				fs.Int("test-int-flag", 0, "")
				fs.String("test-string-flag", "", "")

				return fs
			}(),
			giveFlagName:   "test-int-flag",
			giveErr:        fmt.Errorf("Test123"),
			expectedOutput: "error with flag -test-int-flag: Test123\n  -test-int-flag int\n",
			expectedErr:    "error with flag -test-int-flag: Test123",
		},
		{
			name: "flag_does_not_exist",
			with: &Parser{},
			giveFs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("", flag.ContinueOnError)

				return fs
			}(),
			giveFlagName: "ABCD",
			giveErr:      fmt.Errorf("Test123"),
			expectedErr:  "flag \"ABCD\" does not exist",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			output := bytes.NewBufferString("")
			tc.giveFs.SetOutput(output)
			tc.giveFs.Usage = tc.with.Usage(tc.giveFs)

			// Do
			actualErr := tc.with.FlagError(tc.giveFs, tc.giveFlagName, tc.giveErr)

			// Assert
			assert.Equal(t, tc.expectedOutput, output.String(), "FlagSet Output")

			if tc.expectedErr != "" {
				assert.EqualError(t, actualErr, tc.expectedErr, "FlagError err")
			} else {
				assert.NoError(t, actualErr, "FlagError err")
			}
		})
	}
}

type text struct {
	b []byte
}

func (t *text) UnmarshalText(text []byte) error {
	t.b = text

	return nil
}

func (t *text) MarshalText() ([]byte, error) {
	return t.b, nil
}
