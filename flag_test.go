/*
 * Copyright (c) 2021-present Fabien Potencier <fabien@symfony.com>
 *
 * This file is part of Symfony CLI project
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package console

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

var boolFlagTests = []struct {
	name             string
	expectedForFalse string
	expectedForTrue  string
}{
	{"help", "<info>--help</>\t", "<info>--help</>\t<comment>[default: true]</>"},
	{"h", "<info>-h</>\t", "<info>-h</>\t<comment>[default: true]</>"},
}

func resetEnv(env []string) {
	for _, e := range env {
		fields := strings.SplitN(e, "=", 2)
		os.Setenv(fields[0], fields[1])
	}
}

func TestBoolFlagHelpOutput(t *testing.T) {
	for _, test := range boolFlagTests {
		flag := &BoolFlag{Name: test.name}
		output := flag.String()

		if output != test.expectedForFalse {
			t.Errorf("%q does not match %q", output, test.expectedForFalse)
		}

		flag.DefaultValue = true
		output = flag.String()

		if output != test.expectedForTrue {
			t.Errorf("%q does not match %q", output, test.expectedForTrue)
		}
	}
}

var stringFlagTests = []struct {
	name     string
	aliases  []string
	usage    string
	value    string
	expected string
}{
	{"foo", nil, "", "", "<info>--foo=value</>\t"},
	{"f", nil, "", "", "<info>-f=value</>\t"},
	{"f", nil, "The total `foo` desired", "all", "<info>-f=foo</>\tThe total foo desired <comment>[default: \"all\"]</>"},
	{"test", nil, "", "Something", "<info>--test=value</>\t<comment>[default: \"Something\"]</>"},
	{"config", []string{"c"}, "Load configuration from `FILE`", "", "<info>--config=FILE, -c=FILE</>\tLoad configuration from FILE"},
	{"config", []string{"c"}, "Load configuration from `CONFIG`", "config.json", "<info>--config=CONFIG, -c=CONFIG</>\tLoad configuration from CONFIG <comment>[default: \"config.json\"]</>"},
}

func TestStringFlagHelpOutput(t *testing.T) {
	for _, test := range stringFlagTests {
		flag := &StringFlag{Name: test.name, Aliases: test.aliases, Usage: test.usage, DefaultValue: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringFlagDefaultText(t *testing.T) {
	flag := &StringFlag{Name: "foo", Aliases: nil, Usage: "amount of `foo` requested", DefaultValue: "none", DefaultText: "all of it"}
	expected := "<info>--foo=foo</>\tamount of foo requested <comment>[default: all of it]</>"
	output := flag.String()

	if output != expected {
		t.Errorf("%q does not match %q", output, expected)
	}
}

func TestStringFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_FOO", "derp")
	for _, test := range stringFlagTests {
		flag := &StringFlag{Name: test.name, Aliases: test.aliases, DefaultValue: test.value, EnvVars: []string{"APP_FOO"}}
		output := flag.String()

		expectedSuffix := " [$APP_FOO]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_FOO%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var stringSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *StringSlice
	expected string
}{
	{"foo", nil, NewStringSlice(""), "<info>--foo=value</>\t"},
	{"f", nil, NewStringSlice(""), "<info>-f=value</>\t"},
	{"f", nil, NewStringSlice("Lipstick"), "<info>-f=value</>\t<comment>[default: \"Lipstick\"]</>"},
	{"test", nil, NewStringSlice("Something"), "<info>--test=value</>\t<comment>[default: \"Something\"]</>"},
	{"dee", []string{"d"}, NewStringSlice("Inka", "Dinka", "dooo"), "<info>--dee=value, -d=value</>\t<comment>[default: \"Inka\", \"Dinka\", \"dooo\"]</>"},
}

func TestStringSliceFlagHelpOutput(t *testing.T) {
	for _, test := range stringSliceFlagTests {
		flag := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_QWWX", "11,4")
	for _, test := range stringSliceFlagTests {
		flag := &StringSliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value, EnvVars: []string{"APP_QWWX"}}
		output := flag.String()

		expectedSuffix := " [$APP_QWWX]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_QWWX%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var stringMapFlagTests = []struct {
	name     string
	aliases  []string
	value    *StringMap
	expected string
}{
	{"foo", nil, NewStringMap(map[string]string{}), "<info>--foo=key=value</>\t"},
	{"f", nil, NewStringMap(map[string]string{}), "<info>-f=key=value</>\t"},
	{"f", nil, NewStringMap(map[string]string{"foo": "bar"}), "<info>-f=key=value</>\t<comment>[default: \"foo=bar\"]</>"},
	{"test", nil, NewStringMap(map[string]string{"foo": "bar", "fooz": "baz"}), "<info>--test=key=value</>\t<comment>[default: \"foo=bar\", \"fooz=baz\"]</>"},
	{"dee", []string{"d"}, NewStringMap(map[string]string{}), "<info>--dee=key=value, -d=key=value</>\t"},
}

func TestStringMapFlagHelpOutput(t *testing.T) {
	for _, test := range stringMapFlagTests {
		flag := &StringMapFlag{Name: test.name, Aliases: test.aliases, Destination: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestStringMapFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_QWWX", "11,4")
	for _, test := range stringMapFlagTests {
		flag := &StringMapFlag{Name: test.name, Aliases: test.aliases, Destination: test.value, EnvVars: []string{"APP_QWWX"}}
		output := flag.String()

		expectedSuffix := " [$APP_QWWX]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_QWWX%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var intFlagTests = []struct {
	name     string
	expected string
}{
	{"hats", "<info>--hats=value</>\t<comment>[default: 9]</>"},
	{"H", "<info>-H=value</>\t<comment>[default: 9]</>"},
}

func TestIntFlagHelpOutput(t *testing.T) {
	for _, test := range intFlagTests {
		flag := &IntFlag{Name: test.name, DefaultValue: 9}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestIntFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAR", "2")
	for _, test := range intFlagTests {
		flag := &IntFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var int64FlagTests = []struct {
	name     string
	expected string
}{
	{"hats", "<info>--hats=value</>\t<comment>[default: 8589934592]</>"},
	{"H", "<info>-H=value</>\t<comment>[default: 8589934592]</>"},
}

func TestInt64FlagHelpOutput(t *testing.T) {
	for _, test := range int64FlagTests {
		flag := Int64Flag{Name: test.name, DefaultValue: 8589934592}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestInt64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAR", "2")
	for _, test := range int64FlagTests {
		flag := IntFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var uintFlagTests = []struct {
	name     string
	expected string
}{
	{"nerfs", "<info>--nerfs=value</>\t<comment>[default: 41]</>"},
	{"N", "<info>-N=value</>\t<comment>[default: 41]</>"},
}

func TestUintFlagHelpOutput(t *testing.T) {
	for _, test := range uintFlagTests {
		flag := UintFlag{Name: test.name, DefaultValue: 41}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUintFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAR", "2")
	for _, test := range uintFlagTests {
		flag := UintFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var uint64FlagTests = []struct {
	name     string
	expected string
}{
	{"gerfs", "<info>--gerfs=value</>\t<comment>[default: 8589934582]</>"},
	{"G", "<info>-G=value</>\t<comment>[default: 8589934582]</>"},
}

func TestUint64FlagHelpOutput(t *testing.T) {
	for _, test := range uint64FlagTests {
		flag := Uint64Flag{Name: test.name, DefaultValue: 8589934582}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestUint64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAR", "2")
	for _, test := range uint64FlagTests {
		flag := UintFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var durationFlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "<info>--hooting=value</>\t<comment>[default: 1s]</>"},
	{"H", "<info>-H=value</>\t<comment>[default: 1s]</>"},
}

func TestDurationFlagHelpOutput(t *testing.T) {
	for _, test := range durationFlagTests {
		flag := &DurationFlag{Name: test.name, DefaultValue: 1 * time.Second}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestDurationFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAR", "2h3m6s")
	for _, test := range durationFlagTests {
		flag := &DurationFlag{Name: test.name, EnvVars: []string{"APP_BAR"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAR]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAR%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var intSliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *IntSlice
	expected string
}{
	{"heads", nil, NewIntSlice(), "<info>--heads=value</>\t"},
	{"H", nil, NewIntSlice(), "<info>-H=value</>\t"},
	{"H", []string{"heads"}, NewIntSlice(9, 3), "<info>-H=value, --heads=value</>\t<comment>[default: 9, 3]</>"},
}

func TestIntSliceFlagHelpOutput(t *testing.T) {
	for _, test := range intSliceFlagTests {
		flag := &IntSliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestIntSliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_SMURF", "42,3")
	for _, test := range intSliceFlagTests {
		flag := &IntSliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value, EnvVars: []string{"APP_SMURF"}}
		output := flag.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var int64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *Int64Slice
	expected string
}{
	{"heads", nil, NewInt64Slice(), "<info>--heads=value</>\t"},
	{"H", nil, NewInt64Slice(), "<info>-H=value</>\t"},
	{"heads", []string{"H"}, NewInt64Slice(int64(2), int64(17179869184)),
		"<info>--heads=value, -H=value</>\t<comment>[default: 2, 17179869184]</>"},
}

func TestInt64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range int64SliceFlagTests {
		flag := Int64SliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestInt64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_SMURF", "42,17179869184")
	for _, test := range int64SliceFlagTests {
		flag := Int64SliceFlag{Name: test.name, Destination: test.value, EnvVars: []string{"APP_SMURF"}}
		output := flag.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var float64FlagTests = []struct {
	name     string
	expected string
}{
	{"hooting", "<info>--hooting=value</>\t<comment>[default: 0.1]</>"},
	{"H", "<info>-H=value</>\t<comment>[default: 0.1]</>"},
}

func TestFloat64FlagHelpOutput(t *testing.T) {
	for _, test := range float64FlagTests {
		flag := &Float64Flag{Name: test.name, DefaultValue: float64(0.1)}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFloat64FlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_BAZ", "99.4")
	for _, test := range float64FlagTests {
		flag := &Float64Flag{Name: test.name, EnvVars: []string{"APP_BAZ"}}
		output := flag.String()

		expectedSuffix := " [$APP_BAZ]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_BAZ%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

var float64SliceFlagTests = []struct {
	name     string
	aliases  []string
	value    *Float64Slice
	expected string
}{
	{"heads", nil, NewFloat64Slice(), "<info>--heads=value</>\t"},
	{"H", nil, NewFloat64Slice(), "<info>-H=value</>\t"},
	{"heads", []string{"H"}, NewFloat64Slice(float64(0.1234), float64(-10.5)),
		"<info>--heads=value, -H=value</>\t<comment>[default: 0.1234, -10.5]</>"},
}

func TestFloat64SliceFlagHelpOutput(t *testing.T) {
	for _, test := range float64SliceFlagTests {
		flag := Float64SliceFlag{Name: test.name, Aliases: test.aliases, Destination: test.value}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestFloat64SliceFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_SMURF", "0.1234,-10.5")
	for _, test := range float64SliceFlagTests {
		flag := Float64SliceFlag{Name: test.name, Destination: test.value, EnvVars: []string{"APP_SMURF"}}
		output := flag.String()

		expectedSuffix := " [$APP_SMURF]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_SMURF%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%q does not end with"+expectedSuffix, output)
		}
	}
}

var genericFlagTests = []struct {
	name     string
	value    Generic
	expected string
}{
	{"toads", &Parser{"abc", "def"}, "<info>--toads=value</>\ttest flag <comment>[default: abc,def]</>"},
	{"t", &Parser{"abc", "def"}, "<info>-t=value</>\ttest flag <comment>[default: abc,def]</>"},
}

func TestGenericFlagHelpOutput(t *testing.T) {
	for _, test := range genericFlagTests {
		flag := &GenericFlag{Name: test.name, Destination: test.value, Usage: "test flag"}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%q does not match %q", output, test.expected)
		}
	}
}

func TestGenericFlagWithEnvVarHelpOutput(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_ZAP", "3")
	for _, test := range genericFlagTests {
		flag := &GenericFlag{Name: test.name, EnvVars: []string{"APP_ZAP"}}
		output := flag.String()

		expectedSuffix := " [$APP_ZAP]"
		if runtime.GOOS == "windows" {
			expectedSuffix = " [%APP_ZAP%]"
		}
		if !strings.HasSuffix(output, expectedSuffix) {
			t.Errorf("%s does not end with"+expectedSuffix, output)
		}
	}
}

func TestParseMultiString(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&StringFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("serve") != "10" {
				t.Errorf("main name not set")
			}
			if ctx.String("s") != "10" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10"})
}

func TestParseDestinationString(t *testing.T) {
	var dest string
	a := Application{
		Flags: []Flag{
			&StringFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != "10" {
				t.Errorf("expected destination String 10")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--dest", "10"})
}

func TestParseMultiStringFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_COUNT", "20")
	(&Application{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, EnvVars: []string{"APP_COUNT"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("count") != "20" {
				t.Errorf("main name not set")
			}
			if ctx.String("c") != "20" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_COUNT", "20")
	(&Application{
		Flags: []Flag{
			&StringFlag{Name: "count", Aliases: []string{"c"}, EnvVars: []string{"COMPAT_COUNT", "APP_COUNT"}},
		},
		Action: func(ctx *Context) error {
			if ctx.String("count") != "20" {
				t.Errorf("main name not set")
			}
			if ctx.String("c") != "20" {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSlice(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewStringSlice()},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(ctx.StringSlice("serve"), expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaults(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewStringSlice("9", "2")},
		},
		Action: func(ctx *Context) error {
			expected := []string{"10", "20"}
			if !reflect.DeepEqual(ctx.StringSlice("serve"), expected) {
				t.Errorf("main name not set: %v != %v", expected, ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), expected) {
				t.Errorf("short name not set: %v != %v", expected, ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiStringSliceWithDefaultsUnset(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewStringSlice("9", "2")},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("serve"), []string{"9", "2"}) {
				t.Errorf("main name not set: %v", ctx.StringSlice("serve"))
			}
			if !reflect.DeepEqual(ctx.StringSlice("s"), []string{"9", "2"}) {
				t.Errorf("short name not set: %v", ctx.StringSlice("s"))
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnv(t *testing.T) {
	t.SkipNow()
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewStringSlice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvWithDefaults(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewStringSlice("1", "2", "5"), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewStringSlice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiStringSliceFromEnvCascadeWithDefaults(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&StringSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewStringSlice("1", "2", "5"), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.StringSlice("intervals"), []string{"20", "30", "40"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.StringSlice("i"), []string{"20", "30", "40"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt(t *testing.T) {
	a := Application{
		Flags: []Flag{
			&IntFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("serve") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("s") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run", "-s", "10"})
}

func TestParseDestinationInt(t *testing.T) {
	var dest int
	a := Application{
		Flags: []Flag{
			&IntFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != 10 {
				t.Errorf("expected destination Int 10")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--dest", "10"})
}

func TestParseMultiIntFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_TIMEOUT_SECONDS", "10")
	a := Application{
		Flags: []Flag{
			&IntFlag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("timeout") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("t") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiIntFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_TIMEOUT_SECONDS", "10")
	a := Application{
		Flags: []Flag{
			&IntFlag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Int("timeout") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("t") != 10 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiIntSlice(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewIntSlice()},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{10, 20}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{10, 20}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaults(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewIntSlice(9, 2)},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{10, 20}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{10, 20}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "20"})
}

func TestParseMultiIntSliceWithDefaultsUnset(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewIntSlice(9, 2)},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("serve"), []int{9, 2}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.IntSlice("s"), []int{9, 2}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewIntSlice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnvWithDefaults(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewIntSlice(1, 2, 5), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiIntSliceFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,40")

	(&Application{
		Flags: []Flag{
			&IntSliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewIntSlice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.IntSlice("intervals"), []int{20, 30, 40}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.IntSlice("i"), []int{20, 30, 40}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt64Slice(t *testing.T) {
	(&Application{
		Flags: []Flag{
			&Int64SliceFlag{Name: "serve", Aliases: []string{"s"}, Destination: NewInt64Slice()},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("serve"), []int64{10, 17179869184}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("s"), []int64{10, 17179869184}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}).Run([]string{"run", "-s", "10", "-s", "17179869184"})
}

func TestParseMultiInt64SliceFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,17179869184")

	(&Application{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewInt64Slice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("intervals"), []int64{20, 30, 17179869184}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("i"), []int64{20, 30, 17179869184}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiInt64SliceFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "20,30,17179869184")

	(&Application{
		Flags: []Flag{
			&Int64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewInt64Slice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Int64Slice("intervals"), []int64{20, 30, 17179869184}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Int64Slice("i"), []int64{20, 30, 17179869184}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64(t *testing.T) {
	a := Application{
		Flags: []Flag{
			&Float64Flag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("serve") != 10.2 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("s") != 10.2 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run", "-s", "10.2"})
}

func TestParseDestinationFloat64(t *testing.T) {
	var dest float64
	a := Application{
		Flags: []Flag{
			&Float64Flag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != 10.2 {
				t.Errorf("expected destination Float64 10.2")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--dest", "10.2"})
}

func TestParseMultiFloat64FromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	a := Application{
		Flags: []Flag{
			&Float64Flag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("timeout") != 15.5 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("t") != 15.5 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiFloat64FromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_TIMEOUT_SECONDS", "15.5")
	a := Application{
		Flags: []Flag{
			&Float64Flag{Name: "timeout", Aliases: []string{"t"}, EnvVars: []string{"COMPAT_TIMEOUT_SECONDS", "APP_TIMEOUT_SECONDS"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Float64("timeout") != 15.5 {
				t.Errorf("main name not set")
			}
			if ctx.Float64("t") != 15.5 {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiFloat64SliceFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "0.1,-10.5")

	(&Application{
		Flags: []Flag{
			&Float64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewFloat64Slice(), EnvVars: []string{"APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Float64Slice("intervals"), []float64{0.1, -10.5}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Float64Slice("i"), []float64{0.1, -10.5}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiFloat64SliceFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_INTERVALS", "0.1234,-10.5")

	(&Application{
		Flags: []Flag{
			&Float64SliceFlag{Name: "intervals", Aliases: []string{"i"}, Destination: NewFloat64Slice(), EnvVars: []string{"COMPAT_INTERVALS", "APP_INTERVALS"}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Float64Slice("intervals"), []float64{0.1234, -10.5}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Float64Slice("i"), []float64{0.1234, -10.5}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}).Run([]string{"run"})
}

func TestParseMultiBool(t *testing.T) {
	a := Application{
		Flags: []Flag{
			&BoolFlag{Name: "serve", Aliases: []string{"s"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("serve") != true {
				t.Errorf("main name not set")
			}
			if ctx.Bool("s") != true {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--serve"})
}

func TestParseDestinationBool(t *testing.T) {
	var dest bool
	a := Application{
		Flags: []Flag{
			&BoolFlag{
				Name:        "dest",
				Destination: &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest != true {
				t.Errorf("expected destination Bool true")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--dest"})
}

func TestParseMultiBoolFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_DEBUG", "1")
	a := Application{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, EnvVars: []string{"APP_DEBUG"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") != true {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") != true {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiBoolFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_DEBUG", "1")
	a := Application{
		Flags: []Flag{
			&BoolFlag{Name: "debug", Aliases: []string{"d"}, EnvVars: []string{"COMPAT_DEBUG", "APP_DEBUG"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") != true {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") != true {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiBoolTrue(t *testing.T) {
	a := Application{
		Flags: []Flag{
			&BoolFlag{Name: "implode", Aliases: []string{"i"}, DefaultValue: true},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("implode") {
				t.Errorf("main name not set")
			}
			if ctx.Bool("i") {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--implode=false"})
}

func TestParseDestinationBoolTrue(t *testing.T) {
	dest := true

	a := Application{
		Flags: []Flag{
			&BoolFlag{
				Name:         "dest",
				DefaultValue: true,
				Destination:  &dest,
			},
		},
		Action: func(ctx *Context) error {
			if dest {
				t.Errorf("expected destination Bool false")
			}
			return nil
		},
	}
	a.Run([]string{"run", "--dest=false"})
}

func TestParseMultiBoolTrueFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_DEBUG", "0")
	a := Application{
		Flags: []Flag{
			&BoolFlag{
				Name:         "debug",
				Aliases:      []string{"d"},
				DefaultValue: true,
				EnvVars:      []string{"APP_DEBUG"},
			},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseMultiBoolTrueFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_DEBUG", "0")
	a := Application{
		Flags: []Flag{
			&BoolFlag{
				Name:         "debug",
				Aliases:      []string{"d"},
				DefaultValue: true,
				EnvVars:      []string{"COMPAT_DEBUG", "APP_DEBUG"},
			},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("debug") {
				t.Errorf("main name not set from env")
			}
			if ctx.Bool("d") {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

type Parser [2]string

func (p *Parser) Set(value string) error {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return errors.New("invalid format")
	}

	(*p)[0] = parts[0]
	(*p)[1] = parts[1]

	return nil
}

func (p *Parser) String() string {
	return fmt.Sprintf("%s,%s", p[0], p[1])
}

func TestParseGeneric(t *testing.T) {
	a := Application{
		Flags: []Flag{
			&GenericFlag{Name: "serve", Aliases: []string{"s"}, Destination: &Parser{}},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("serve"), &Parser{"10", "20"}) {
				t.Errorf("main name not set")
			}
			if !reflect.DeepEqual(ctx.Generic("s"), &Parser{"10", "20"}) {
				t.Errorf("short name not set")
			}
			return nil
		},
	}
	a.Run([]string{"run", "-s", "10,20"})
}

func TestParseGenericFromEnv(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_SERVE", "20,30")
	a := Application{
		Flags: []Flag{
			&GenericFlag{
				Name:        "serve",
				Aliases:     []string{"s"},
				Destination: &Parser{},
				EnvVars:     []string{"APP_SERVE"},
			},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("serve"), &Parser{"20", "30"}) {
				t.Errorf("main name not set from env")
			}
			if !reflect.DeepEqual(ctx.Generic("s"), &Parser{"20", "30"}) {
				t.Errorf("short name not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestParseGenericFromEnvCascade(t *testing.T) {
	t.SkipNow()
	defer resetEnv(os.Environ())
	os.Clearenv()
	os.Setenv("APP_FOO", "99,2000")
	a := Application{
		Flags: []Flag{
			&GenericFlag{
				Name:        "foos",
				Destination: &Parser{},
				EnvVars:     []string{"COMPAT_FOO", "APP_FOO"},
			},
		},
		Action: func(ctx *Context) error {
			if !reflect.DeepEqual(ctx.Generic("foos"), &Parser{"99", "2000"}) {
				t.Errorf("value not set from env")
			}
			return nil
		},
	}
	a.Run([]string{"run"})
}

func TestStringSlice_Serialized_Set(t *testing.T) {
	sl0 := NewStringSlice("a", "b")
	ser0 := sl0.Serialized()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewStringSlice("c", "d")
	sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestIntSlice_Serialized_Set(t *testing.T) {
	sl0 := NewIntSlice(1, 2)
	ser0 := sl0.Serialized()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewIntSlice(3, 4)
	sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestInt64Slice_Serialized_Set(t *testing.T) {
	sl0 := NewInt64Slice(int64(1), int64(2))
	ser0 := sl0.Serialized()

	if len(ser0) < len(slPfx) {
		t.Fatalf("serialized shorter than expected: %q", ser0)
	}

	sl1 := NewInt64Slice(int64(3), int64(4))
	sl1.Set(ser0)

	if sl0.String() != sl1.String() {
		t.Fatalf("pre and post serialization do not match: %v != %v", sl0, sl1)
	}
}

func TestBlackfireCurlArgsParsing(t *testing.T) {
	hasRun := false
	app := Application{
		Commands: []*Command{
			{
				Category:    "client",
				Name:        "curl",
				Aliases:     []*Alias{{Name: "curl"}},
				Usage:       "Profile a URL via curl",
				Description: "Profile a URL via curl (see `man curl` for additional options)",
				FlagParsing: FlagParsingSkippedAfterFirstArg,
				Args: []*Arg{
					{Name: "cmd", Optional: true, Slice: true, Description: "The cURL command"},
				},
				Flags: []Flag{
					&StringFlag{Name: "client-id"},
					&StringFlag{Name: "client-token"},
					&IntFlag{Name: "samples", Usage: "Set the number of samples to collect (0 to use Blackfire default)"},
				},
				Action: func(c *Context) error {
					hasRun = true
					if samples := c.Int("samples"); samples != 3 {
						t.Errorf("Invalid samples flag value, 3 expected, %v received", samples)
					}

					expectedArgs := []string{"-X", "POST", "http://app.bkf"}
					args := c.Args().Tail()
					if !reflect.DeepEqual(args, expectedArgs) {
						t.Errorf("Invalid arguments, %v expected, %v received", expectedArgs, args)
					}

					return nil
				},
			},
		},
	}

	err := app.Run([]string{"blackfire", "curl", "--samples=3", "-X", "POST", "http://app.bkf"})
	if err != nil {
		t.Fatal(err)
	}

	if !hasRun {
		t.Fatal("Action didn't run")
	}
}
