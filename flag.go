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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/posener/complete"
	"github.com/symfony-cli/terminal"
)

const defaultPlaceholder = "value"

var (
	slPfx = fmt.Sprintf("sl:::%d:::", time.Now().UTC().UnixNano())

	commaWhitespace = regexp.MustCompile("[, ]+.*")
)

// VersionFlag prints the version for the application
var VersionFlag = &BoolFlag{
	Name:  "V",
	Usage: "Print the version",
}

// HelpFlag prints the help for all commands and subcommands.
// Set to nil to disable the flag.
var HelpFlag = &BoolFlag{
	Name:    "help",
	Aliases: []string{"h"},
	Usage:   "Show help",
}

// FlagStringer converts a flag definition to a string. This is used by help
// to display a flag.
var FlagStringer FlagStringFunc = stringifyFlag

// Serializeder is used to circumvent the limitations of flag.FlagSet.Set
type Serializeder interface {
	Serialized() string
}

// FlagsByName is a slice of Flag.
type FlagsByName []Flag

func (f FlagsByName) Len() int {
	return len(f)
}

func (f FlagsByName) Less(i, j int) bool {
	if len(f[j].Names()) == 0 {
		return false
	} else if len(f[i].Names()) == 0 {
		return true
	}
	return f[i].Names()[0] < f[j].Names()[0]
}

func (f FlagsByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Flag is a common interface related to parsing flags in cli.
// For more advanced flag parsing techniques, it is recommended that
// this interface be implemented.
type Flag interface {
	fmt.Stringer

	PredictArgs(*Context, complete.Args) []string
	Validate(*Context) error
	// Apply Flag settings to the given flag set
	Apply(*flag.FlagSet)
	Names() []string
}

func flagSet(fsName string, flags []Flag) *flag.FlagSet {
	allFlagsNames := make(map[string]interface{})
	set := flag.NewFlagSet(fsName, flag.ContinueOnError)

	for _, f := range flags {
		currentFlagNames := flagNames(f)
		for _, name := range currentFlagNames {
			if _, alreadyThere := allFlagsNames[name]; alreadyThere {
				var msg string
				if fsName == "" {
					msg = fmt.Sprintf("flag redefined: %s", name)
				} else {
					msg = fmt.Sprintf("%s flag redefined: %s", fsName, name)
				}
				fmt.Fprintln(terminal.Stderr, msg)
				panic(msg) // Happens only if flags are declared with identical names
			}
			allFlagsNames[name] = nil
		}
		f.Apply(set)
	}
	return set
}

// Generic is a generic parseable type identified by a specific flag
type Generic interface {
	Set(value string) error
	String() string
}

// Apply takes the flagset and calls Set on the generic flag with the value
// provided by the user for parsing by the flag
func (f *GenericFlag) Apply(set *flag.FlagSet) {
	set.Var(f.Destination, f.Name, f.Usage)
}

// StringMap wraps a map[string]string to satisfy flag.Value
type StringMap struct {
	m          map[string]string
	hasBeenSet bool
}

// NewStringMap creates a *StringMap with default values
func NewStringMap(m map[string]string) *StringMap {
	return &StringMap{m: m}
}

// Set appends the string value to the list of values
func (m *StringMap) Set(value string) error {
	if !m.hasBeenSet {
		m.m = make(map[string]string)
		m.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &m.m)
		m.hasBeenSet = true
		return nil
	}

	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return errors.New("please use key=value format")
	}
	m.m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (m *StringMap) String() string {
	if m == nil {
		return ""
	}
	if len(m.m) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	keys := make([]string, 0, len(m.m))
	for key := range m.m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		buffer.WriteString(fmt.Sprintf(`"%s=%s", `, key, m.m[key]))
	}
	return strings.Trim(buffer.String(), ", ")
}

// Serialized allows StringSlice to fulfill Serializeder
func (m *StringMap) Serialized() string {
	jsonBytes, _ := json.Marshal(m)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the map set by this flag
func (m *StringMap) Value() map[string]string {
	return m.m
}

// Apply populates the flag given the flag set and environment
func (f *StringMapFlag) Apply(set *flag.FlagSet) {
	if f.Destination == nil {
		f.Destination = NewStringMap(make(map[string]string))
	}
	set.Var(f.Destination, f.Name, f.Usage)
}

// StringSlice wraps a []string to satisfy flag.Value
type StringSlice struct {
	slice      []string
	hasBeenSet bool
}

// NewStringSlice creates a *StringSlice with default values
func NewStringSlice(defaults ...string) *StringSlice {
	return &StringSlice{slice: append([]string{}, defaults...)}
}

// Set appends the string value to the list of values
func (f *StringSlice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []string{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	f.slice = append(f.slice, value)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *StringSlice) String() string {
	return fmt.Sprintf("%s", f.slice)
}

// Serialized allows StringSlice to fulfill Serializeder
func (f *StringSlice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of strings set by this flag
func (f *StringSlice) Value() []string {
	return f.slice
}

// Apply populates the flag given the flag set and environment
func (f *StringSliceFlag) Apply(set *flag.FlagSet) {
	if f.Destination == nil {
		f.Destination = NewStringSlice()
	}

	set.Var(f.Destination, f.Name, f.Usage)
}

// IntSlice wraps an []int to satisfy flag.Value
type IntSlice struct {
	slice      []int
	hasBeenSet bool
}

// NewIntSlice makes an *IntSlice with default values
func NewIntSlice(defaults ...int) *IntSlice {
	return &IntSlice{slice: append([]int{}, defaults...)}
}

// NewInt64Slice makes an *Int64Slice with default values
func NewInt64Slice(defaults ...int64) *Int64Slice {
	return &Int64Slice{slice: append([]int64{}, defaults...)}
}

// SetInt directly adds an integer to the list of values
func (i *IntSlice) SetInt(value int) {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	i.slice = append(i.slice, value)
}

// Set parses the value into an integer and appends it to the list of values
func (i *IntSlice) Set(value string) error {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &i.slice)
		i.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return errors.WithStack(err)
	}

	i.slice = append(i.slice, int(tmp))
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *IntSlice) String() string {
	return fmt.Sprintf("%#v", i.slice)
}

// Serialized allows IntSlice to fulfill Serializeder
func (i *IntSlice) Serialized() string {
	jsonBytes, _ := json.Marshal(i.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *IntSlice) Value() []int {
	return i.slice
}

// Apply populates the flag given the flag set and environment
func (f *IntSliceFlag) Apply(set *flag.FlagSet) {
	if f.Destination == nil {
		f.Destination = NewIntSlice()
	}

	set.Var(f.Destination, f.Name, f.Usage)
}

// Int64Slice is an opaque type for []int to satisfy flag.Value
type Int64Slice struct {
	slice      []int64
	hasBeenSet bool
}

// Set parses the value into an integer and appends it to the list of values
func (f *Int64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []int64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return errors.WithStack(err)
	}

	f.slice = append(f.slice, tmp)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Int64Slice) String() string {
	return fmt.Sprintf("%#v", f.slice)
}

// Serialized allows Int64Slice to fulfill Serializeder
func (f *Int64Slice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (f *Int64Slice) Value() []int64 {
	return f.slice
}

// Apply populates the flag given the flag set and environment
func (f *Int64SliceFlag) Apply(set *flag.FlagSet) {
	if f.Destination == nil {
		f.Destination = NewInt64Slice()
	}

	set.Var(f.Destination, f.Name, f.Usage)
}

// Apply populates the flag given the flag set and environment
func (f *BoolFlag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.BoolVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Bool(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *StringFlag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.StringVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.String(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *IntFlag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.IntVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Int(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *Int64Flag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.Int64Var(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Int64(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *UintFlag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.UintVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Uint(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *Uint64Flag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.Uint64Var(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Uint64(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *DurationFlag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.DurationVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Duration(f.Name, f.DefaultValue, f.Usage)
	}
}

// Apply populates the flag given the flag set and environment
func (f *Float64Flag) Apply(set *flag.FlagSet) {
	if f.Destination != nil {
		set.Float64Var(f.Destination, f.Name, f.DefaultValue, f.Usage)
	} else {
		set.Float64(f.Name, f.DefaultValue, f.Usage)
	}
}

// NewFloat64Slice makes a *Float64Slice with default values
func NewFloat64Slice(defaults ...float64) *Float64Slice {
	return &Float64Slice{slice: append([]float64{}, defaults...)}
}

// Float64Slice is an opaque type for []float64 to satisfy flag.Value
type Float64Slice struct {
	slice      []float64
	hasBeenSet bool
}

// Set parses the value into a float64 and appends it to the list of values
func (f *Float64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []float64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return errors.WithStack(err)
	}

	f.slice = append(f.slice, tmp)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Float64Slice) String() string {
	return fmt.Sprintf("%#v", f.slice)
}

// Serialized allows Float64Slice to fulfill Serializeder
func (f *Float64Slice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of float64s set by this flag
func (f *Float64Slice) Value() []float64 {
	return f.slice
}

// Apply populates the flag given the flag set and environment
func (f *Float64SliceFlag) Apply(set *flag.FlagSet) {
	if f.Destination == nil {
		f.Destination = NewFloat64Slice()
	}

	set.Var(f.Destination, f.Name, f.Usage)
}

func visibleFlags(fl []Flag) []Flag {
	visible := []Flag{}
	for _, flag := range fl {
		if !flagValue(flag).FieldByName("Hidden").Bool() {
			visible = append(visible, flag)
		}
	}
	return visible
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

// Returns the placeholder, if any, and the unquoted usage string.
func unquoteUsage(usage string) (string, string) {
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name := usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break
		}
	}
	return "", usage
}

func prefixedNames(names []string, placeholder string) string {
	var prefixed string
	for i, name := range names {
		if name == "" {
			continue
		}

		prefixed += prefixFor(name) + name
		if placeholder != "" {
			prefixed += "=" + placeholder
		}
		if i < len(names)-1 {
			prefixed += ", "
		}
	}
	return prefixed
}

func withEnvHint(envVars []string, str string) string {
	envText := ""
	if len(envVars) > 0 {
		prefix := "$"
		suffix := ""
		sep := ", $"
		if runtime.GOOS == "windows" {
			prefix = "%"
			suffix = "%"
			sep = "%, %"
		}
		envText = fmt.Sprintf(" [%s%s%s]", prefix, strings.Join(envVars, sep), suffix)
	}
	return str + envText
}

func flagName(f Flag) string {
	return flagStringField(f, "Name")
}

func flagNames(f Flag) []string {
	aliases := append([]string{flagStringField(f, "Name")}, flagStringSliceField(f, "Aliases")...)

	ret := make([]string, 0, len(aliases))

	for _, part := range aliases {
		// v1 -> v2 migration warning zone:
		// Strip off anything after the first found comma or space, which
		// *hopefully* makes it a tiny bit more obvious that unexpected behavior is
		// caused by using the v1 form of stringly typed "Name".
		ret = append(ret, commaWhitespace.ReplaceAllString(part, ""))
	}

	return ret
}

func flagStringSliceField(f Flag, name string) []string {
	fv := flagValue(f)
	field := fv.FieldByName(name)

	if field.IsValid() {
		return field.Interface().([]string)
	}

	return []string{}
}

func flagStringField(f Flag, name string) string {
	fv := flagValue(f)
	field := fv.FieldByName(name)

	if field.IsValid() {
		return field.String()
	}

	return ""
}

func flagValue(f Flag) reflect.Value {
	fv := reflect.ValueOf(f)
	for fv.Kind() == reflect.Ptr {
		fv = reflect.Indirect(fv)
	}
	return fv
}

func flagIsRequired(f Flag) bool {
	field := flagValue(f).FieldByName("Required")
	if field.IsValid() && field.Kind() == reflect.Bool {
		return field.Bool()
	}

	return false
}

func stringifyFlag(f Flag) string {
	fv := flagValue(f)

	switch f := f.(type) {
	case *IntSliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyIntSliceFlag(f))
	case *Int64SliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyInt64SliceFlag(f))
	case *Float64SliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyFloat64SliceFlag(f))
	case *StringSliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyStringSliceFlag(f))
	case *StringMapFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyStringMapFlag(f))
	}

	placeholder, usage := unquoteUsage(fv.FieldByName("Usage").String())

	needsPlaceholder := false
	defaultValueString := ""
	val := fv.FieldByName("DefaultValue")
	if !val.IsValid() {
		val = fv.FieldByName("Destination")
	}
	if val.IsValid() {
		needsPlaceholder = val.Kind() != reflect.Bool

		if val.Kind() == reflect.String && val.String() != "" {
			defaultValueString = fmt.Sprintf("%q", val.String())
		} else if val.Kind() != reflect.Bool || val.Bool() {
			defaultValueString = fmt.Sprintf("%v", val.Interface())
		}
	}

	helpText := fv.FieldByName("DefaultText")
	if helpText.IsValid() && helpText.String() != "" {
		needsPlaceholder = val.Kind() != reflect.Bool
		defaultValueString = helpText.String()
	}

	if defaultValueString != "" {
		defaultValueString = fmt.Sprintf(" <comment>[default: %s]</>", defaultValueString)
	}
	requiredString := ""
	if flagIsRequired(f) {
		requiredString = " <comment>(required)</>"
	}

	if needsPlaceholder && placeholder == "" {
		placeholder = defaultPlaceholder
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s%s", usage, defaultValueString, requiredString))

	return withEnvHint(flagStringSliceField(f, "EnvVars"),
		fmt.Sprintf("<info>%s</>\t%s", prefixedNames(f.Names(), placeholder), usageWithDefault))
}

func stringifyIntSliceFlag(f *IntSliceFlag) string {
	defaultVals := []string{}
	if f.Destination != nil && len(f.Destination.Value()) > 0 {
		for _, i := range f.Destination.Value() {
			defaultVals = append(defaultVals, fmt.Sprintf("%d", i))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyInt64SliceFlag(f *Int64SliceFlag) string {
	defaultVals := []string{}
	if f.Destination != nil && len(f.Destination.Value()) > 0 {
		for _, i := range f.Destination.Value() {
			defaultVals = append(defaultVals, fmt.Sprintf("%d", i))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyFloat64SliceFlag(f *Float64SliceFlag) string {
	defaultVals := []string{}
	if f.Destination != nil && len(f.Destination.Value()) > 0 {
		for _, i := range f.Destination.Value() {
			defaultVals = append(defaultVals, strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", i), "0"), "."))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyStringSliceFlag(f *StringSliceFlag) string {
	defaultVals := []string{}
	if f.Destination != nil && len(f.Destination.Value()) > 0 {
		for _, s := range f.Destination.Value() {
			if len(s) > 0 {
				defaultVals = append(defaultVals, fmt.Sprintf("%q", s))
			}
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifySliceFlag(usage string, names, defaultVals []string) string {
	placeholder, usage := unquoteUsage(usage)
	if placeholder == "" {
		placeholder = defaultPlaceholder
	}

	defaultVal := ""
	if len(defaultVals) > 0 {
		defaultVal = fmt.Sprintf(" <comment>[default: %s]</>", strings.Join(defaultVals, ", "))
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s", usage, defaultVal))
	return fmt.Sprintf("<info>%s</>\t%s", prefixedNames(names, placeholder), usageWithDefault)
}

func stringifyStringMapFlag(f *StringMapFlag) string {
	return stringifyMapFlag(f.Usage, f.Names(), f.Destination)
}

func stringifyMapFlag(usage string, names []string, defaultVals fmt.Stringer) string {
	placeholder, usage := unquoteUsage(usage)
	if placeholder == "" {
		placeholder = "key=value"
	}

	defaultVal := ""
	if v := defaultVals.String(); len(v) > 0 {
		defaultVal = fmt.Sprintf(" <comment>[default: %s]</>", v)
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s", usage, defaultVal))
	return fmt.Sprintf("<info>%s</>\t%s", prefixedNames(names, placeholder), usageWithDefault)
}

func hasFlag(flags []Flag, fl Flag) bool {
	for _, existing := range flags {
		if fl == existing {
			return true
		}
	}

	return false
}
