# Copyright (c) 2021-present Fabien Potencier <fabien@symfony.com>
#
# This file is part of Symfony CLI project
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.
#
# Fish completions for the CLI binary
#
# References:
#   - https://github.com/symfony/symfony/blob/6.4/src/Symfony/Component/Console/Resources/completion.fish
#   - https://github.com/posener/complete/blob/master/install/fish.go
#   - https://github.com/fish-shell/fish-shell/blob/master/share/completions/sudo.fish
#

function __complete_{{ .App.HelpName }}
    set -lx COMP_LINE (commandline -cp)
    test -z (commandline -ct)
    and set COMP_LINE "$COMP_LINE "
    {{ .CurrentBinaryInvocation }} self:autocomplete
end

complete -f -c '{{ .App.HelpName }}' -n "not __fish_seen_subcommand_from console php pecl composer run local:run" -a '(__complete_{{ .App.HelpName }})'

if type -t _sf_console >/dev/null 2>/dev/null
  # this wrapper function allows us to call Symfony autocompletion letting it
  # knows how to call the `bin/console` using the Symfony CLI binary (to ensure
  # the right env and PHP versions are used)
  function __complete_{{ .App.HelpName }}_console
      set -x _SF_CMD "{{ .CurrentBinaryInvocation }}" "console"
      _sf_console
  end

  complete -f -c '{{ .App.HelpName }}' -n "__fish_seen_subcommand_from console" -a '(__complete_{{ .App.HelpName }}_console)' -f
end
