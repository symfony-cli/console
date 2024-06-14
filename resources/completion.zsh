#compdef {{ .App.HelpName }}

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
# zsh completions for {{ .App.HelpName }}
#
# References:
#   - https://github.com/symfony/symfony/blob/6.4/src/Symfony/Component/Console/Resources/completion.zsh
#   - https://github.com/posener/complete/blob/master/install/zsh.go
#   - https://stackoverflow.com/a/13547531
#

# this wrapper function allows us to let Symfony knows how to call the
# `bin/console` using the Symfony CLI binary (to ensure the right env and PHP
# versions are used)
_{{ .App.HelpName }}_console() {
  # shellcheck disable=SC2068
  {{ .CurrentBinaryInvocation }} console $@
}

_complete_{{ .App.HelpName }}() {
    local lastParam flagPrefix requestComp out comp
    local -a completions

    # The user could have moved the cursor backwards on the command-line.
    # We need to trigger completion from the $CURRENT location, so we need
    # to truncate the command-line ($words) up to the $CURRENT location.
    # (We cannot use $CURSOR as its value does not work when a command is an alias.)
    words=("${=words[1,CURRENT]}") lastParam=${words[-1]}

    # For zsh, when completing a flag with an = (e.g., {{ .App.HelpName }} -n=<TAB>)
    # completions must be prefixed with the flag
    setopt local_options BASH_REMATCH
    if [[ "${lastParam}" =~ '-.*=' ]]; then
        # We are dealing with a flag with an =
        flagPrefix="-P ${BASH_REMATCH}"
    fi

    # detect if we are in a wrapper command and need to "forward" completion to it
    for ((i = 1; i <= $#words; i++)); do
        if [[ "${words[i]}" != -* ]]; then
              case "${words[i]}" in
              console|php|pecl|composer|run|local:run)
                shift words
                (( CURRENT-- ))
                _SF_CMD="_{{ .App.HelpName }}_console" _normal
                return
            esac;
        fi
    done

    while IFS='\n' read -r comp; do
        if [ -n "$comp" ]; then
            # We first need to escape any : as part of the completion itself.
            comp=${comp//:/\\:}
            completions+=${comp}
        fi
    done < <(COMP_LINE="$words" COMP_POINT="$((CURRENT-1))" ${words[0]} ${_SF_CMD:-${words[1]}} self:autocomplete)

    # Let inbuilt _describe handle completions
    eval _describe "completions" completions $flagPrefix
}

compdef _complete_{{ .App.HelpName }} {{ .App.HelpName }}
