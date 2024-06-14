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
	"flag"
	"strconv"
	"time"

	"github.com/posener/complete"
)

// BoolFlag is a flag with type bool
type BoolFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, bool) error
	Destination   *bool
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *BoolFlag) String() string {
	return FlagStringer(f)
}

func (f *BoolFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{"true", "false"}
}

func (f *BoolFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Bool(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *BoolFlag) Names() []string {
	return flagNames(f)
}

// Bool looks up the value of a local BoolFlag, returns
// false if not found
func (c *Context) Bool(name string) bool {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupBool(name, f)
	}
	return false
}

func lookupBool(name string, f *flag.Flag) bool {
	if f == nil {
		return false
	}

	if parsed, err := strconv.ParseBool(f.Value.String()); err == nil {
		return parsed
	}

	return false
}

// DurationFlag is a flag with type time.Duration (see https://golang.org/pkg/time/#ParseDuration)
type DurationFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  time.Duration
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, time.Duration) error
	Destination   *time.Duration
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *DurationFlag) String() string {
	return FlagStringer(f)
}

func (f *DurationFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *DurationFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Duration(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *DurationFlag) Names() []string {
	return flagNames(f)
}

// Duration looks up the value of a local DurationFlag, returns
// 0 if not found
func (c *Context) Duration(name string) time.Duration {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupDuration(name, f)
	}
	return 0
}

func lookupDuration(name string, f *flag.Flag) time.Duration {
	if f == nil {
		return 0
	}

	if parsed, err := time.ParseDuration(f.Value.String()); err == nil {
		return parsed
	}

	return 0
}

// Float64Flag is a flag with type float64
type Float64Flag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  float64
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, float64) error
	Destination   *float64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64Flag) String() string {
	return FlagStringer(f)
}

func (f *Float64Flag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *Float64Flag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Float64(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *Float64Flag) Names() []string {
	return flagNames(f)
}

// Float64 looks up the value of a local Float64Flag, returns
// 0 if not found
func (c *Context) Float64(name string) float64 {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupFloat64(name, f)
	}
	return 0
}

func lookupFloat64(name string, f *flag.Flag) float64 {
	if f == nil {
		return 0
	}

	if parsed, err := strconv.ParseFloat(f.Value.String(), 64); err == nil {
		return parsed
	}

	return 0
}

// GenericFlag is a flag with type Generic
type GenericFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, interface{}) error
	Destination   Generic
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *GenericFlag) String() string {
	return FlagStringer(f)
}

func (f *GenericFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *GenericFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Generic(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *GenericFlag) Names() []string {
	return flagNames(f)
}

// Generic looks up the value of a local GenericFlag, returns
// nil if not found
func (c *Context) Generic(name string) interface{} {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupGeneric(name, f)
	}
	return nil
}

func lookupGeneric(name string, f *flag.Flag) interface{} {
	if f == nil {
		return nil
	}

	if parsed, err := f.Value, error(nil); err == nil {
		return parsed
	}

	return nil
}

// Int64Flag is a flag with type int64
type Int64Flag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  int64
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, int64) error
	Destination   *int64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64Flag) String() string {
	return FlagStringer(f)
}

func (f *Int64Flag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *Int64Flag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Int64(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *Int64Flag) Names() []string {
	return flagNames(f)
}

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (c *Context) Int64(name string) int64 {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupInt64(name, f)
	}
	return 0
}

func lookupInt64(name string, f *flag.Flag) int64 {
	if f == nil {
		return 0
	}

	if parsed, err := strconv.ParseInt(f.Value.String(), 0, 64); err == nil {
		return parsed
	}

	return 0
}

// IntFlag is a flag with type int
type IntFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  int
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, int) error
	Destination   *int
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntFlag) String() string {
	return FlagStringer(f)
}

func (f *IntFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *IntFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Int(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *IntFlag) Names() []string {
	return flagNames(f)
}

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (c *Context) Int(name string) int {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupInt(name, f)
	}
	return 0
}

func lookupInt(name string, f *flag.Flag) int {
	if f == nil {
		return 0
	}

	if parsed, err := strconv.ParseInt(f.Value.String(), 0, 64); err == nil {
		return int(parsed)
	}

	return 0
}

// IntSliceFlag is a flag with type *IntSlice
type IntSliceFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, []int) error
	Destination   *IntSlice
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntSliceFlag) String() string {
	return FlagStringer(f)
}

func (f *IntSliceFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *IntSliceFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.IntSlice(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *IntSliceFlag) Names() []string {
	return flagNames(f)
}

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (c *Context) IntSlice(name string) []int {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupIntSlice(name, f)
	}
	return nil
}

func lookupIntSlice(name string, f *flag.Flag) []int {
	if f == nil {
		return nil
	}

	if asserted, ok := f.Value.(*IntSlice); !ok {
		return nil
	} else if parsed, err := asserted.Value(), error(nil); err == nil {
		return parsed
	}

	return nil
}

// Int64SliceFlag is a flag with type *Int64Slice
type Int64SliceFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, []int64) error
	Destination   *Int64Slice
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Int64SliceFlag) String() string {
	return FlagStringer(f)
}

func (f *Int64SliceFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *Int64SliceFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Int64Slice(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *Int64SliceFlag) Names() []string {
	return flagNames(f)
}

// Int64Slice looks up the value of a local Int64SliceFlag, returns
// nil if not found
func (c *Context) Int64Slice(name string) []int64 {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupInt64Slice(name, f)
	}
	return nil
}

func lookupInt64Slice(name string, f *flag.Flag) []int64 {
	if f == nil {
		return nil
	}

	if asserted, ok := f.Value.(*Int64Slice); !ok {
		return nil
	} else if parsed, err := asserted.Value(), error(nil); err == nil {
		return parsed
	}

	return nil
}

// Float64SliceFlag is a flag with type *Float64Slice
type Float64SliceFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, []float64) error
	Destination   *Float64Slice
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64SliceFlag) String() string {
	return FlagStringer(f)
}

func (f *Float64SliceFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *Float64SliceFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Float64Slice(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *Float64SliceFlag) Names() []string {
	return flagNames(f)
}

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (c *Context) Float64Slice(name string) []float64 {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupFloat64Slice(name, f)
	}
	return nil
}

func lookupFloat64Slice(name string, f *flag.Flag) []float64 {
	if f == nil {
		return nil
	}

	if asserted, ok := f.Value.(*Float64Slice); !ok {
		return nil
	} else if parsed, err := asserted.Value(), error(nil); err == nil {
		return parsed
	}

	return nil
}

// StringFlag is a flag with type string
type StringFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  string
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, string) error
	Destination   *string
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringFlag) String() string {
	return FlagStringer(f)
}

func (f *StringFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *StringFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.String(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *StringFlag) Names() []string {
	return flagNames(f)
}

// String looks up the value of a local StringFlag, returns
// "" if not found
func (c *Context) String(name string) string {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupString(name, f)
	}
	return ""
}

func lookupString(name string, f *flag.Flag) string {
	if f == nil {
		return ""
	}

	if parsed, err := f.Value.String(), error(nil); err == nil {
		return parsed
	}

	return ""
}

// StringSliceFlag is a flag with type *StringSlice
type StringSliceFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, []string) error
	Destination   *StringSlice
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringSliceFlag) String() string {
	return FlagStringer(f)
}

func (f *StringSliceFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *StringSliceFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.StringSlice(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *StringSliceFlag) Names() []string {
	return flagNames(f)
}

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (c *Context) StringSlice(name string) []string {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupStringSlice(name, f)
	}
	return nil
}

func lookupStringSlice(name string, f *flag.Flag) []string {
	if f == nil {
		return nil
	}

	if asserted, ok := f.Value.(*StringSlice); !ok {
		return nil
	} else if parsed, err := asserted.Value(), error(nil); err == nil {
		return parsed
	}

	return nil
}

// StringMapFlag is a flag with type *StringMap
type StringMapFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, map[string]string) error
	Destination   *StringMap
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringMapFlag) String() string {
	return FlagStringer(f)
}

func (f *StringMapFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *StringMapFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.StringMap(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *StringMapFlag) Names() []string {
	return flagNames(f)
}

// StringMap looks up the value of a local StringMapFlag, returns
// nil if not found
func (c *Context) StringMap(name string) map[string]string {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupStringMap(name, f)
	}
	return nil
}

func lookupStringMap(name string, f *flag.Flag) map[string]string {
	if f == nil {
		return nil
	}

	if asserted, ok := f.Value.(*StringMap); !ok {
		return nil
	} else if parsed, err := asserted.Value(), error(nil); err == nil {
		return parsed
	}

	return nil
}

// Uint64Flag is a flag with type uint64
type Uint64Flag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  uint64
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, uint64) error
	Destination   *uint64
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Uint64Flag) String() string {
	return FlagStringer(f)
}

func (f *Uint64Flag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *Uint64Flag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Uint64(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *Uint64Flag) Names() []string {
	return flagNames(f)
}

// Uint64 looks up the value of a local Uint64Flag, returns
// 0 if not found
func (c *Context) Uint64(name string) uint64 {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupUint64(name, f)
	}
	return 0
}

func lookupUint64(name string, f *flag.Flag) uint64 {
	if f == nil {
		return 0
	}

	if parsed, err := strconv.ParseUint(f.Value.String(), 0, 64); err == nil {
		return parsed
	}

	return 0
}

// UintFlag is a flag with type uint
type UintFlag struct {
	Name          string
	Aliases       []string
	Usage         string
	EnvVars       []string
	Hidden        bool
	DefaultValue  uint
	DefaultText   string
	Required      bool
	ArgsPredictor func(*Context, complete.Args) []string
	Validator     func(*Context, uint) error
	Destination   *uint
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *UintFlag) String() string {
	return FlagStringer(f)
}

func (f *UintFlag) PredictArgs(c *Context, a complete.Args) []string {
	if f.ArgsPredictor != nil {
		return f.ArgsPredictor(c, a)
	}
	return []string{}
}

func (f *UintFlag) Validate(c *Context) error {
	if f.Validator != nil {
		return f.Validator(c, c.Uint(f.Name))
	}
	return nil
}

// Names returns the names of the flag
func (f *UintFlag) Names() []string {
	return flagNames(f)
}

// Uint looks up the value of a local UintFlag, returns
// 0 if not found
func (c *Context) Uint(name string) uint {
	if f := lookupRawFlag(name, c); f != nil {
		return lookupUint(name, f)
	}
	return 0
}

func lookupUint(name string, f *flag.Flag) uint {
	if f == nil {
		return 0
	}

	if parsed, err := strconv.ParseUint(f.Value.String(), 0, 64); err == nil {
		return uint(parsed)
	}

	return 0
}
