package console

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type ArgDefinition []*Arg

func (def ArgDefinition) Usage() string {
	if len(def) < 1 {
		return ""
	}
	buf := bytes.Buffer{}

	buf.WriteString(" [--]")

	for _, arg := range def {
		element := "<" + arg.Name + ">"
		if arg.Optional {
			element = "[" + element + "]"
		} else if arg.Slice {
			element = "(" + element + ")"
		}

		if arg.Slice {
			element += "..."
		}

		buf.WriteString(" ")
		buf.WriteString(element)
	}

	return strings.TrimRight(buf.String(), " ")
}

type Arg struct {
	Name, Default string
	Description   string
	Optional      bool
	Slice         bool
}

func (a *Arg) String() string {
	requiredString := ""
	if !a.Optional {
		requiredString = " <comment>(required)</>"
	}

	defaultValueString := ""
	if a.Default != "" {
		defaultValueString = fmt.Sprintf(` <comment>[default: "%s"]</>`, a.Default)
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s%s", a.Description, defaultValueString, requiredString))
	return fmt.Sprintf("<info>%s</>\t%s", a.Name, usageWithDefault)
}

func checkArgsModes(args []*Arg) {
	arguments := make(map[string]bool)
	hasSliceArgument := false
	hasOptional := false

	for _, arg := range args {
		if arguments[arg.Name] {
			panic(fmt.Sprintf(`An argument with name "%s" already exists.`, arg.Name))
		}

		if hasSliceArgument {
			panic("Cannot add an argument after an array argument.")
		}
		if !arg.Optional && hasOptional {
			panic("Cannot add a required argument after an optional one.")
		}

		if arg.Slice {
			hasSliceArgument = true
		}
		if arg.Optional {
			hasOptional = true
		}

		arguments[arg.Name] = true
	}
}

func checkRequiredArgs(command *Command, context *Context) error {
	args := context.Args()
	hasSliceArgument := false
	maximumArgsLen := 0

	for _, arg := range command.Args {
		if arg.Slice {
			hasSliceArgument = true
		} else {
			maximumArgsLen++
		}

		if arg.Optional {
			continue
		}

		if arg.Slice {
			if len(args.Tail()) < 1 {
				return errors.Errorf(`Required argument "%s" is not set`, arg.Name)
			}
			break
		}

		if args.Get(arg.Name) == "" {
			return errors.Errorf(`Required argument "%s" is not set`, arg.Name)
		}
	}

	if !hasSliceArgument && args.Len() > maximumArgsLen {
		return errors.New("Too many arguments")
	}

	return nil
}
