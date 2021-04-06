# bash completion for chezmoi                              -*- shell-script -*-

__chezmoi_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__chezmoi_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__chezmoi_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__chezmoi_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__chezmoi_handle_go_custom_completion()
{
    __chezmoi_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly chezmoi allows to handle aliases
    args=("${words[@]:1}")
    requestComp="${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __chezmoi_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __chezmoi_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __chezmoi_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __chezmoi_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __chezmoi_debug "${FUNCNAME[0]}: the completions are: ${out[*]}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __chezmoi_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __chezmoi_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __chezmoi_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out[*]}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __chezmoi_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subDir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out[0]}")
        if [ -n "$subdir" ]; then
            __chezmoi_debug "Listing directories in $subdir"
            __chezmoi_handle_subdirs_in_dir_flag "$subdir"
        else
            __chezmoi_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out[*]}" -- "$cur")
    fi
}

__chezmoi_handle_reply()
{
    __chezmoi_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __chezmoi_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __chezmoi_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __chezmoi_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
		if declare -F __chezmoi_custom_func >/dev/null; then
			# try command name qualified custom func
			__chezmoi_custom_func
		else
			# otherwise fall back to unqualified for compatibility
			declare -F __custom_func >/dev/null && __custom_func
		fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__chezmoi_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__chezmoi_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__chezmoi_handle_flag()
{
    __chezmoi_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __chezmoi_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __chezmoi_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __chezmoi_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __chezmoi_contains_word "${words[c]}" "${two_word_flags[@]}"; then
			  __chezmoi_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__chezmoi_handle_noun()
{
    __chezmoi_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __chezmoi_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __chezmoi_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__chezmoi_handle_command()
{
    __chezmoi_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_chezmoi_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __chezmoi_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__chezmoi_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __chezmoi_handle_reply
        return
    fi
    __chezmoi_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __chezmoi_handle_flag
    elif __chezmoi_contains_word "${words[c]}" "${commands[@]}"; then
        __chezmoi_handle_command
    elif [[ $c -eq 0 ]]; then
        __chezmoi_handle_command
    elif __chezmoi_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __chezmoi_handle_command
        else
            __chezmoi_handle_noun
        fi
    else
        __chezmoi_handle_noun
    fi
    __chezmoi_handle_word
}

_chezmoi_add()
{
    last_command="chezmoi_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--autotemplate")
    flags+=("-a")
    local_nonpersistent_flags+=("--autotemplate")
    local_nonpersistent_flags+=("-a")
    flags+=("--create")
    local_nonpersistent_flags+=("--create")
    flags+=("--empty")
    flags+=("-e")
    local_nonpersistent_flags+=("--empty")
    local_nonpersistent_flags+=("-e")
    flags+=("--encrypt")
    local_nonpersistent_flags+=("--encrypt")
    flags+=("--exact")
    local_nonpersistent_flags+=("--exact")
    flags+=("--follow")
    flags+=("-f")
    local_nonpersistent_flags+=("--follow")
    local_nonpersistent_flags+=("-f")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--template")
    flags+=("-T")
    local_nonpersistent_flags+=("--template")
    local_nonpersistent_flags+=("-T")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_apply()
{
    last_command="chezmoi_apply"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_archive()
{
    last_command="chezmoi_archive"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    flags+=("--gzip")
    flags+=("-z")
    local_nonpersistent_flags+=("--gzip")
    local_nonpersistent_flags+=("-z")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_cat()
{
    last_command="chezmoi_cat"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_cd()
{
    last_command="chezmoi_cd"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_chattr()
{
    last_command="chezmoi_chattr"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("+a")
    must_have_one_noun+=("+after")
    must_have_one_noun+=("+b")
    must_have_one_noun+=("+before")
    must_have_one_noun+=("+e")
    must_have_one_noun+=("+empty")
    must_have_one_noun+=("+encrypted")
    must_have_one_noun+=("+exact")
    must_have_one_noun+=("+executable")
    must_have_one_noun+=("+o")
    must_have_one_noun+=("+once")
    must_have_one_noun+=("+p")
    must_have_one_noun+=("+private")
    must_have_one_noun+=("+t")
    must_have_one_noun+=("+template")
    must_have_one_noun+=("+x")
    must_have_one_noun+=("-a")
    must_have_one_noun+=("-after")
    must_have_one_noun+=("-b")
    must_have_one_noun+=("-before")
    must_have_one_noun+=("-e")
    must_have_one_noun+=("-empty")
    must_have_one_noun+=("-encrypted")
    must_have_one_noun+=("-exact")
    must_have_one_noun+=("-executable")
    must_have_one_noun+=("-o")
    must_have_one_noun+=("-once")
    must_have_one_noun+=("-p")
    must_have_one_noun+=("-private")
    must_have_one_noun+=("-t")
    must_have_one_noun+=("-template")
    must_have_one_noun+=("-x")
    must_have_one_noun+=("a")
    must_have_one_noun+=("after")
    must_have_one_noun+=("b")
    must_have_one_noun+=("before")
    must_have_one_noun+=("e")
    must_have_one_noun+=("empty")
    must_have_one_noun+=("encrypted")
    must_have_one_noun+=("exact")
    must_have_one_noun+=("executable")
    must_have_one_noun+=("noa")
    must_have_one_noun+=("noafter")
    must_have_one_noun+=("nob")
    must_have_one_noun+=("nobefore")
    must_have_one_noun+=("noe")
    must_have_one_noun+=("noempty")
    must_have_one_noun+=("noencrypted")
    must_have_one_noun+=("noexact")
    must_have_one_noun+=("noexecutable")
    must_have_one_noun+=("noo")
    must_have_one_noun+=("noonce")
    must_have_one_noun+=("nop")
    must_have_one_noun+=("noprivate")
    must_have_one_noun+=("not")
    must_have_one_noun+=("notemplate")
    must_have_one_noun+=("nox")
    must_have_one_noun+=("o")
    must_have_one_noun+=("once")
    must_have_one_noun+=("p")
    must_have_one_noun+=("private")
    must_have_one_noun+=("t")
    must_have_one_noun+=("template")
    must_have_one_noun+=("x")
    noun_aliases=()
}

_chezmoi_completion()
{
    last_command="chezmoi_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("powershell")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_chezmoi_data()
{
    last_command="chezmoi_data"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_diff()
{
    last_command="chezmoi_diff"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_docs()
{
    last_command="chezmoi_docs"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_doctor()
{
    last_command="chezmoi_doctor"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_dump()
{
    last_command="chezmoi_dump"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_edit()
{
    last_command="chezmoi_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--apply")
    flags+=("-a")
    local_nonpersistent_flags+=("--apply")
    local_nonpersistent_flags+=("-a")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_edit-config()
{
    last_command="chezmoi_edit-config"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_execute-template()
{
    last_command="chezmoi_execute-template"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--init")
    flags+=("-i")
    local_nonpersistent_flags+=("--init")
    local_nonpersistent_flags+=("-i")
    flags+=("--promptBool=")
    two_word_flags+=("--promptBool")
    local_nonpersistent_flags+=("--promptBool")
    local_nonpersistent_flags+=("--promptBool=")
    flags+=("--promptInt=")
    two_word_flags+=("--promptInt")
    local_nonpersistent_flags+=("--promptInt")
    local_nonpersistent_flags+=("--promptInt=")
    flags+=("--promptString=")
    two_word_flags+=("--promptString")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--promptString")
    local_nonpersistent_flags+=("--promptString=")
    local_nonpersistent_flags+=("-p")
    flags+=("--stdinisatty")
    local_nonpersistent_flags+=("--stdinisatty")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_forget()
{
    last_command="chezmoi_forget"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_git()
{
    last_command="chezmoi_git"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_help()
{
    last_command="chezmoi_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_import()
{
    last_command="chezmoi_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--exact")
    local_nonpersistent_flags+=("--exact")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--remove-destination")
    flags+=("-r")
    local_nonpersistent_flags+=("--remove-destination")
    local_nonpersistent_flags+=("-r")
    flags+=("--strip-components=")
    two_word_flags+=("--strip-components")
    local_nonpersistent_flags+=("--strip-components")
    local_nonpersistent_flags+=("--strip-components=")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_init()
{
    last_command="chezmoi_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--apply")
    flags+=("-a")
    local_nonpersistent_flags+=("--apply")
    local_nonpersistent_flags+=("-a")
    flags+=("--depth=")
    two_word_flags+=("--depth")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--depth")
    local_nonpersistent_flags+=("--depth=")
    local_nonpersistent_flags+=("-d")
    flags+=("--one-shot")
    local_nonpersistent_flags+=("--one-shot")
    flags+=("--purge")
    flags+=("-p")
    local_nonpersistent_flags+=("--purge")
    local_nonpersistent_flags+=("-p")
    flags+=("--purge-binary")
    flags+=("-P")
    local_nonpersistent_flags+=("--purge-binary")
    local_nonpersistent_flags+=("-P")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_managed()
{
    last_command="chezmoi_managed"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_merge()
{
    last_command="chezmoi_merge"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_purge()
{
    last_command="chezmoi_purge"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--binary")
    flags+=("-P")
    local_nonpersistent_flags+=("--binary")
    local_nonpersistent_flags+=("-P")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_remove()
{
    last_command="chezmoi_remove"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_keyring_get()
{
    last_command="chezmoi_secret_keyring_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--service=")
    two_word_flags+=("--service")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--user=")
    two_word_flags+=("--user")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_keyring_set()
{
    last_command="chezmoi_secret_keyring_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--service=")
    two_word_flags+=("--service")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--user=")
    two_word_flags+=("--user")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_keyring()
{
    last_command="chezmoi_secret_keyring"

    command_aliases=()

    commands=()
    commands+=("get")
    commands+=("set")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--service=")
    two_word_flags+=("--service")
    flags+=("--user=")
    two_word_flags+=("--user")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--service=")
    must_have_one_flag+=("--user=")
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret()
{
    last_command="chezmoi_secret"

    command_aliases=()

    commands=()
    commands+=("keyring")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_source-path()
{
    last_command="chezmoi_source-path"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_state_dump()
{
    last_command="chezmoi_state_dump"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_state_reset()
{
    last_command="chezmoi_state_reset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_state()
{
    last_command="chezmoi_state"

    command_aliases=()

    commands=()
    commands+=("dump")
    commands+=("reset")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_status()
{
    last_command="chezmoi_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_unmanaged()
{
    last_command="chezmoi_unmanaged"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_update()
{
    last_command="chezmoi_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--apply")
    flags+=("-a")
    local_nonpersistent_flags+=("--apply")
    local_nonpersistent_flags+=("-a")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_upgrade()
{
    last_command="chezmoi_upgrade"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--method=")
    two_word_flags+=("--method")
    local_nonpersistent_flags+=("--method")
    local_nonpersistent_flags+=("--method=")
    flags+=("--owner=")
    two_word_flags+=("--owner")
    local_nonpersistent_flags+=("--owner")
    local_nonpersistent_flags+=("--owner=")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    local_nonpersistent_flags+=("--repo")
    local_nonpersistent_flags+=("--repo=")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_verify()
{
    last_command="chezmoi_verify"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--recursive")
    flags+=("-r")
    local_nonpersistent_flags+=("--recursive")
    local_nonpersistent_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_root_command()
{
    last_command="chezmoi"

    command_aliases=()

    commands=()
    commands+=("add")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("manage")
        aliashash["manage"]="add"
    fi
    commands+=("apply")
    commands+=("archive")
    commands+=("cat")
    commands+=("cd")
    commands+=("chattr")
    commands+=("completion")
    commands+=("data")
    commands+=("diff")
    commands+=("docs")
    commands+=("doctor")
    commands+=("dump")
    commands+=("edit")
    commands+=("edit-config")
    commands+=("execute-template")
    commands+=("forget")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("unmanage")
        aliashash["unmanage"]="forget"
    fi
    commands+=("git")
    commands+=("help")
    commands+=("import")
    commands+=("init")
    commands+=("managed")
    commands+=("merge")
    commands+=("purge")
    commands+=("remove")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("rm")
        aliashash["rm"]="remove"
    fi
    commands+=("secret")
    commands+=("source-path")
    commands+=("state")
    commands+=("status")
    commands+=("unmanaged")
    commands+=("update")
    commands+=("upgrade")
    commands+=("verify")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags_with_completion+=("--config")
    flags_completion+=("_filedir")
    two_word_flags+=("-c")
    flags_with_completion+=("-c")
    flags_completion+=("_filedir")
    flags+=("--cpu-profile=")
    two_word_flags+=("--cpu-profile")
    flags_with_completion+=("--cpu-profile")
    flags_completion+=("_filedir")
    flags+=("--debug")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    flags_with_completion+=("--destination")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-D")
    flags_with_completion+=("-D")
    flags_completion+=("_filedir -d")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-x")
    flags+=("--force")
    flags+=("--keep-going")
    flags+=("-k")
    flags+=("--no-pager")
    flags+=("--no-tty")
    flags+=("--output=")
    two_word_flags+=("--output")
    flags_with_completion+=("--output")
    flags_completion+=("_filedir")
    two_word_flags+=("-o")
    flags_with_completion+=("-o")
    flags_completion+=("_filedir")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    flags_with_completion+=("--source")
    flags_completion+=("_filedir -d")
    two_word_flags+=("-S")
    flags_with_completion+=("-S")
    flags_completion+=("_filedir -d")
    flags+=("--source-path")
    flags+=("--use-builtin-git=")
    two_word_flags+=("--use-builtin-git")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_chezmoi()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __chezmoi_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("chezmoi")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function
    local last_command
    local nouns=()

    __chezmoi_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_chezmoi chezmoi
else
    complete -o default -o nospace -F __start_chezmoi chezmoi
fi

# ex: ts=4 sw=4 et filetype=sh
