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

__chezmoi_handle_reply()
{
    __chezmoi_debug "${FUNCNAME[0]}"
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
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
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
        completions=("${must_have_one_noun[@]}")
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
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
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1
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

    flags+=("--empty")
    flags+=("-e")
    flags+=("--encrypt")
    flags+=("--exact")
    flags+=("-x")
    flags+=("--follow")
    flags+=("-f")
    flags+=("--prompt")
    flags+=("-p")
    flags+=("--recursive")
    flags+=("-r")
    flags+=("--template")
    flags+=("-T")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
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
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
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
    two_word_flags+=("-f")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    flags+=("--recursive")
    flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    flags+=("--diff")
    flags+=("-d")
    flags+=("--prompt")
    flags+=("-p")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    flags+=("-x")
    flags+=("--remove-destination")
    flags+=("-r")
    flags+=("--strip-components=")
    two_word_flags+=("--strip-components")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_bitwarden()
{
    last_command="chezmoi_secret_bitwarden"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_generic()
{
    last_command="chezmoi_secret_generic"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_keepassxc()
{
    last_command="chezmoi_secret_keepassxc"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--service=")
    two_word_flags+=("--service")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--password=")
    two_word_flags+=("--password")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--service=")
    two_word_flags+=("--service")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_flag+=("--service=")
    must_have_one_flag+=("--user=")
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_lastpass()
{
    last_command="chezmoi_secret_lastpass"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_onepassword()
{
    last_command="chezmoi_secret_onepassword"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_pass()
{
    last_command="chezmoi_secret_pass"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret_vault()
{
    last_command="chezmoi_secret_vault"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_secret()
{
    last_command="chezmoi_secret"

    command_aliases=()

    commands=()
    commands+=("bitwarden")
    commands+=("generic")
    commands+=("keepassxc")
    commands+=("keyring")
    commands+=("lastpass")
    commands+=("onepassword")
    commands+=("pass")
    commands+=("vault")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_chezmoi_source()
{
    last_command="chezmoi_source"

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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--force")
    flags+=("-f")
    flags+=("--method=")
    two_word_flags+=("--method")
    two_word_flags+=("-m")
    flags+=("--owner=")
    two_word_flags+=("--owner")
    two_word_flags+=("-o")
    flags+=("--repo=")
    two_word_flags+=("--repo")
    two_word_flags+=("-r")
    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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

    flags+=("--color=")
    two_word_flags+=("--color")
    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
    commands+=("doctor")
    commands+=("dump")
    commands+=("edit")
    commands+=("edit-config")
    commands+=("forget")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("unmanage")
        aliashash["unmanage"]="forget"
    fi
    commands+=("import")
    commands+=("init")
    commands+=("merge")
    commands+=("remove")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("rm")
        aliashash["rm"]="remove"
    fi
    commands+=("secret")
    commands+=("source")
    commands+=("source-path")
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
    two_word_flags+=("-c")
    flags+=("--destination=")
    two_word_flags+=("--destination")
    two_word_flags+=("-D")
    flags+=("--dry-run")
    flags+=("-n")
    flags+=("--remove")
    flags+=("--source=")
    two_word_flags+=("--source")
    two_word_flags+=("-S")
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
