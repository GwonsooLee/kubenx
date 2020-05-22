/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io"
	"bytes"
	"context"
	"github.com/spf13/cobra"
)

const defaultBoilerPlate = `
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var (
	completionShells = map[string]func(out io.Writer, boilerPlate string, cmd *cobra.Command) error{
		"bash": runCompletionBash,
		"zsh":  runCompletionZsh,
	}
)

//Create Command for get pod
func NewCmdCompletion() *cobra.Command {
	return NewCmd("completion").
		WithDescription("Generate bash/zsh completion script").
		WithLongDescription(`To load completion run
. <(kubenx completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(kubenx completion)
`		).
		RunWithArgsAndCmd(execCompletion)
}

// Function for get command
func execCompletion(_ context.Context, out io.Writer, cmd *cobra.Command, args []string) error {
	err := RunCompletion(out, "", cmd, args)
	CheckErr(err)

	return err
}

// RunCompletion checks given arguments and executes command
func RunCompletion(out io.Writer, boilerPlate string, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return UsageErrorf(cmd, "Shell not specified.")
	}
	if len(args) > 1 {
		return UsageErrorf(cmd, "Too many arguments. Expected only the shell type.")
	}
	run, found := completionShells[args[0]]
	if !found {
		return UsageErrorf(cmd, "Unsupported shell type %q.", args[0])
	}

	return run(out, boilerPlate, cmd.Parent())
}

func runCompletionBash(out io.Writer, boilerPlate string, kubenx *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return kubenx.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, boilerPlate string, kubenx *cobra.Command) error {
	zshHead := "#compdef \n"

	out.Write([]byte(zshHead))

	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	zshInitialization := `
__kubenx_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__kubenx_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kubenx_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__kubenx_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__kubenx_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__kubenx_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__kubenx_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__kubenx_filedir() {
	local RET OLD_IFS w qw
	__kubenx_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __kubenx_debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__kubenx_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__kubenx_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__kubenx_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kubenx_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kubenx_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kubenx_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kubenx_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kubenx_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kubenx_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	kubenx.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__kubenx_bash_source <(__kubenx_convert_bash_to_zsh)
_complete kubenx 2>/dev/null
`
	out.Write([]byte(zshTail))
	return nil
}
