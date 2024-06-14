//go:build darwin || linux || freebsd || openbsd

package console

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/symfony-cli/terminal"
)

// completionTemplates holds our shell completions templates.
//
//go:embed resources/completion.*
var completionTemplates embed.FS

var shellAutoCompleteInstallCommand = &Command{
	Category: "self",
	Name:     "completion",
	Aliases: []*Alias{
		{Name: "completion"},
	},
	Usage: "Dumps the completion script for the current shell",
	ShellComplete: func(*Context, string) []string {
		return []string{"bash", "zsh", "fish"}
	},
	Description: `The <info>{{.HelpName}}</> command dumps the shell completion script required
to use shell autocompletion (currently, bash, zsh and fish completion are supported).

<comment>Static installation
-------------------</>

Dump the script to a global completion file and restart your shell:

   <info>{{.HelpName}} {{ call .Shell }} | sudo tee {{ call .CompletionFile }}</>

Or dump the script to a local file and source it:

   <info>{{.HelpName}} {{ call .Shell }} > completion.sh</>

   <comment># source the file whenever you use the project</>
   <info>source completion.sh</>

   <comment># or add this line at the end of your "{{ call .RcFile }}" file:</>
   <info>source /path/to/completion.sh</>

<comment>Dynamic installation
--------------------</>

Add this to the end of your shell configuration file (e.g. <info>"{{ call .RcFile }}"</>):

   <info>eval "$({{.HelpName}} {{ call .Shell }})"</>`,
	DescriptionFunc: func(command *Command, application *Application) string {
		var buf bytes.Buffer

		tpl := template.Must(template.New("description").Parse(command.Description))

		if err := tpl.Execute(&buf, struct {
			// allows to directly access any field from the command inside the template
			*Command
			Shell          func() string
			RcFile         func() string
			CompletionFile func() string
		}{
			Command: command,
			Shell:   guessShell,
			RcFile: func() string {
				switch guessShell() {
				case "fish":
					return "~/.config/fish/config.fish"
				case "zsh":
					return "~/.zshrc"
				default:
					return "~/.bashrc"
				}
			},
			CompletionFile: func() string {
				switch guessShell() {
				case "fish":
					return fmt.Sprintf("/etc/fish/completions/%s.fish", application.HelpName)
				case "zsh":
					return fmt.Sprintf("$fpath[1]/_%s", application.HelpName)
				default:
					return fmt.Sprintf("/etc/bash_completion.d/%s", application.HelpName)
				}
			},
		}); err != nil {
			panic(err)
		}

		return buf.String()
	},
	Args: []*Arg{
		{
			Name:        "shell",
			Description: `The shell type (e.g. "bash"), the value of the "$SHELL" env var will be used if this is not given`,
			Optional:    true,
		},
	},
	Action: func(c *Context) error {
		shell := c.Args().Get("shell")
		if shell == "" {
			shell = guessShell()
		}

		templates, err := template.ParseFS(completionTemplates, "resources/*")
		if err != nil {
			return errors.WithStack(err)
		}

		if tpl := templates.Lookup(fmt.Sprintf("completion.%s", shell)); tpl != nil {
			return errors.WithStack(tpl.Execute(terminal.Stdout, c))
		}

		var supportedShell []string

		for _, tmpl := range templates.Templates() {
			if tmpl.Tree == nil || tmpl.Root == nil {
				continue
			}
			supportedShell = append(supportedShell, strings.TrimLeft(path.Ext(tmpl.Name()), "."))
		}

		if shell == "" {
			return errors.Errorf(`shell not detected, supported shells: "%s"`, strings.Join(supportedShell, ", "))
		}

		return errors.Errorf(`shell "%s" is not supported, supported shells: "%s"`, shell, strings.Join(supportedShell, ", "))
	},
}

func guessShell() string {
	return path.Base(os.Getenv("SHELL"))
}
