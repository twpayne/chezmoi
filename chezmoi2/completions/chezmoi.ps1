using namespace System.Management.Automation
using namespace System.Management.Automation.Language
Register-ArgumentCompleter -Native -CommandName 'chezmoi' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commandElements = $commandAst.CommandElements
    $command = @(
        'chezmoi'
        for ($i = 1; $i -lt $commandElements.Count; $i++) {
            $element = $commandElements[$i]
            if ($element -isnot [StringConstantExpressionAst] -or
                $element.StringConstantType -ne [StringConstantType]::BareWord -or
                $element.Value.StartsWith('-')) {
                break
            }
            $element.Value
        }
    ) -join ';'
    $completions = @(switch ($command) {
        'chezmoi' {
            [CompletionResult]::new('--color', 'color', [CompletionResultType]::ParameterName, 'colorize diffs')
            [CompletionResult]::new('-c', 'c', [CompletionResultType]::ParameterName, 'config file')
            [CompletionResult]::new('--config', 'config', [CompletionResultType]::ParameterName, 'config file')
            [CompletionResult]::new('--cpu-profile', 'cpu-profile', [CompletionResultType]::ParameterName, 'write CPU profile to file')
            [CompletionResult]::new('--debug', 'debug', [CompletionResultType]::ParameterName, 'write debug logs')
            [CompletionResult]::new('-D', 'D', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('--destination', 'destination', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('-n', 'n', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--dry-run', 'dry-run', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--force', 'force', [CompletionResultType]::ParameterName, 'force')
            [CompletionResult]::new('--format', 'format', [CompletionResultType]::ParameterName, 'format (json or yaml)')
            [CompletionResult]::new('-k', 'k', [CompletionResultType]::ParameterName, 'keep going as far as possible after an error')
            [CompletionResult]::new('--keep-going', 'keep-going', [CompletionResultType]::ParameterName, 'keep going as far as possible after an error')
            [CompletionResult]::new('--no-tty', 'no-tty', [CompletionResultType]::ParameterName, 'don''t attempt to get a TTY for reading passwords')
            [CompletionResult]::new('-o', 'o', [CompletionResultType]::ParameterName, 'output file')
            [CompletionResult]::new('--output', 'output', [CompletionResultType]::ParameterName, 'output file')
            [CompletionResult]::new('--remove', 'remove', [CompletionResultType]::ParameterName, 'remove targets')
            [CompletionResult]::new('-S', 'S', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--source', 'source', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--use-builtin-git', 'use-builtin-git', [CompletionResultType]::ParameterName, 'use builtin git')
            [CompletionResult]::new('-v', 'v', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('--verbose', 'verbose', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('add', 'add', [CompletionResultType]::ParameterValue, 'Add an existing file, directory, or symlink to the source state')
            [CompletionResult]::new('apply', 'apply', [CompletionResultType]::ParameterValue, 'Update the destination directory to match the target state')
            [CompletionResult]::new('archive', 'archive', [CompletionResultType]::ParameterValue, 'Generate a tar archive of the target state')
            [CompletionResult]::new('cat', 'cat', [CompletionResultType]::ParameterValue, 'Print the target contents of a file or symlink')
            [CompletionResult]::new('cd', 'cd', [CompletionResultType]::ParameterValue, 'Launch a shell in the source directory')
            [CompletionResult]::new('chattr', 'chattr', [CompletionResultType]::ParameterValue, 'Change the attributes of a target in the source state')
            [CompletionResult]::new('completion', 'completion', [CompletionResultType]::ParameterValue, 'Generate shell completion code')
            [CompletionResult]::new('data', 'data', [CompletionResultType]::ParameterValue, 'Print the template data')
            [CompletionResult]::new('diff', 'diff', [CompletionResultType]::ParameterValue, 'Print the diff between the target state and the destination state')
            [CompletionResult]::new('docs', 'docs', [CompletionResultType]::ParameterValue, 'Print documentation')
            [CompletionResult]::new('doctor', 'doctor', [CompletionResultType]::ParameterValue, 'Check your system for potential problems')
            [CompletionResult]::new('dump', 'dump', [CompletionResultType]::ParameterValue, 'Generate a dump of the target state')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit the source state of a target')
            [CompletionResult]::new('edit-config', 'edit-config', [CompletionResultType]::ParameterValue, 'Edit the configuration file')
            [CompletionResult]::new('execute-template', 'execute-template', [CompletionResultType]::ParameterValue, 'Execute the given template(s)')
            [CompletionResult]::new('forget', 'forget', [CompletionResultType]::ParameterValue, 'Remove a target from the source state')
            [CompletionResult]::new('git', 'git', [CompletionResultType]::ParameterValue, 'Run git in the source directory')
            [CompletionResult]::new('help', 'help', [CompletionResultType]::ParameterValue, 'Print help about a command')
            [CompletionResult]::new('import', 'import', [CompletionResultType]::ParameterValue, 'Import a tar archive into the source state')
            [CompletionResult]::new('init', 'init', [CompletionResultType]::ParameterValue, 'Setup the source directory and update the destination directory to match the target state')
            [CompletionResult]::new('managed', 'managed', [CompletionResultType]::ParameterValue, 'List the managed entries in the destination directory')
            [CompletionResult]::new('merge', 'merge', [CompletionResultType]::ParameterValue, 'Perform a three-way merge between the destination state, the source state, and the target state')
            [CompletionResult]::new('purge', 'purge', [CompletionResultType]::ParameterValue, 'Purge chezmoi''s configuration and data')
            [CompletionResult]::new('remove', 'remove', [CompletionResultType]::ParameterValue, 'Remove a target from the source state and the destination directory')
            [CompletionResult]::new('secret', 'secret', [CompletionResultType]::ParameterValue, 'Interact with a secret manager')
            [CompletionResult]::new('source-path', 'source-path', [CompletionResultType]::ParameterValue, 'Print the path of a target in the source state')
            [CompletionResult]::new('state', 'state', [CompletionResultType]::ParameterValue, 'Manipulate the persistent state')
            [CompletionResult]::new('status', 'status', [CompletionResultType]::ParameterValue, 'Show the status of targets')
            [CompletionResult]::new('unmanaged', 'unmanaged', [CompletionResultType]::ParameterValue, 'List the unmanaged files in the destination directory')
            [CompletionResult]::new('update', 'update', [CompletionResultType]::ParameterValue, 'Pull and apply any changes')
            [CompletionResult]::new('verify', 'verify', [CompletionResultType]::ParameterValue, 'Exit with success if the destination state matches the target state, fail otherwise')
            break
        }
        'chezmoi;add' {
            [CompletionResult]::new('-a', 'a', [CompletionResultType]::ParameterName, 'auto generate the template when adding files as templates')
            [CompletionResult]::new('--autotemplate', 'autotemplate', [CompletionResultType]::ParameterName, 'auto generate the template when adding files as templates')
            [CompletionResult]::new('-e', 'e', [CompletionResultType]::ParameterName, 'add empty files')
            [CompletionResult]::new('--empty', 'empty', [CompletionResultType]::ParameterName, 'add empty files')
            [CompletionResult]::new('--encrypt', 'encrypt', [CompletionResultType]::ParameterName, 'encrypt files')
            [CompletionResult]::new('-x', 'x', [CompletionResultType]::ParameterName, 'add directories exactly')
            [CompletionResult]::new('--exact', 'exact', [CompletionResultType]::ParameterName, 'add directories exactly')
            [CompletionResult]::new('--exists', 'exists', [CompletionResultType]::ParameterName, 'add files that should exist, irrespective of their contents')
            [CompletionResult]::new('-f', 'f', [CompletionResultType]::ParameterName, 'add symlink targets instead of symlinks')
            [CompletionResult]::new('--follow', 'follow', [CompletionResultType]::ParameterName, 'add symlink targets instead of symlinks')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('-T', 'T', [CompletionResultType]::ParameterName, 'add files as templates')
            [CompletionResult]::new('--template', 'template', [CompletionResultType]::ParameterName, 'add files as templates')
            break
        }
        'chezmoi;apply' {
            [CompletionResult]::new('--ignore-encrypted', 'ignore-encrypted', [CompletionResultType]::ParameterName, 'ignore encrypted files')
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--source-path', 'source-path', [CompletionResultType]::ParameterName, 'specify targets by source path')
            break
        }
        'chezmoi;archive' {
            [CompletionResult]::new('--format', 'format', [CompletionResultType]::ParameterName, 'format (tar or zip)')
            [CompletionResult]::new('-z', 'z', [CompletionResultType]::ParameterName, 'compress the output with gzip')
            [CompletionResult]::new('--gzip', 'gzip', [CompletionResultType]::ParameterName, 'compress the output with gzip')
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
        'chezmoi;cat' {
            break
        }
        'chezmoi;cd' {
            break
        }
        'chezmoi;chattr' {
            break
        }
        'chezmoi;completion' {
            [CompletionResult]::new('--color', 'color', [CompletionResultType]::ParameterName, 'colorize diffs')
            [CompletionResult]::new('-c', 'c', [CompletionResultType]::ParameterName, 'config file')
            [CompletionResult]::new('--config', 'config', [CompletionResultType]::ParameterName, 'config file')
            [CompletionResult]::new('--cpu-profile', 'cpu-profile', [CompletionResultType]::ParameterName, 'write CPU profile to file')
            [CompletionResult]::new('--debug', 'debug', [CompletionResultType]::ParameterName, 'write debug logs')
            [CompletionResult]::new('-D', 'D', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('--destination', 'destination', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('-n', 'n', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--dry-run', 'dry-run', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--force', 'force', [CompletionResultType]::ParameterName, 'force')
            [CompletionResult]::new('--format', 'format', [CompletionResultType]::ParameterName, 'format (json or yaml)')
            [CompletionResult]::new('-h', 'h', [CompletionResultType]::ParameterName, 'help for completion')
            [CompletionResult]::new('--help', 'help', [CompletionResultType]::ParameterName, 'help for completion')
            [CompletionResult]::new('-k', 'k', [CompletionResultType]::ParameterName, 'keep going as far as possible after an error')
            [CompletionResult]::new('--keep-going', 'keep-going', [CompletionResultType]::ParameterName, 'keep going as far as possible after an error')
            [CompletionResult]::new('--no-tty', 'no-tty', [CompletionResultType]::ParameterName, 'don''t attempt to get a TTY for reading passwords')
            [CompletionResult]::new('-o', 'o', [CompletionResultType]::ParameterName, 'output file')
            [CompletionResult]::new('--output', 'output', [CompletionResultType]::ParameterName, 'output file')
            [CompletionResult]::new('--remove', 'remove', [CompletionResultType]::ParameterName, 'remove targets')
            [CompletionResult]::new('-S', 'S', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--source', 'source', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--use-builtin-git', 'use-builtin-git', [CompletionResultType]::ParameterName, 'use builtin git')
            [CompletionResult]::new('-v', 'v', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('--verbose', 'verbose', [CompletionResultType]::ParameterName, 'verbose')
            break
        }
        'chezmoi;data' {
            break
        }
        'chezmoi;diff' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--no-pager', 'no-pager', [CompletionResultType]::ParameterName, 'disable pager')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
        'chezmoi;docs' {
            break
        }
        'chezmoi;doctor' {
            break
        }
        'chezmoi;dump' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
        'chezmoi;edit' {
            [CompletionResult]::new('-a', 'a', [CompletionResultType]::ParameterName, 'apply edit after editing')
            [CompletionResult]::new('--apply', 'apply', [CompletionResultType]::ParameterName, 'apply edit after editing')
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            break
        }
        'chezmoi;edit-config' {
            break
        }
        'chezmoi;execute-template' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'simulate chezmoi init')
            [CompletionResult]::new('--init', 'init', [CompletionResultType]::ParameterName, 'simulate chezmoi init')
            [CompletionResult]::new('--promptBool', 'promptBool', [CompletionResultType]::ParameterName, 'simulate promptBool')
            [CompletionResult]::new('--promptInt', 'promptInt', [CompletionResultType]::ParameterName, 'simulate promptInt')
            [CompletionResult]::new('-p', 'p', [CompletionResultType]::ParameterName, 'simulate promptString')
            [CompletionResult]::new('--promptString', 'promptString', [CompletionResultType]::ParameterName, 'simulate promptString')
            break
        }
        'chezmoi;forget' {
            break
        }
        'chezmoi;git' {
            break
        }
        'chezmoi;help' {
            break
        }
        'chezmoi;import' {
            [CompletionResult]::new('-d', 'd', [CompletionResultType]::ParameterName, 'destination prefix')
            [CompletionResult]::new('--destination', 'destination', [CompletionResultType]::ParameterName, 'destination prefix')
            [CompletionResult]::new('-x', 'x', [CompletionResultType]::ParameterName, 'import directories exactly')
            [CompletionResult]::new('--exact', 'exact', [CompletionResultType]::ParameterName, 'import directories exactly')
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'remove destination before import')
            [CompletionResult]::new('--remove-destination', 'remove-destination', [CompletionResultType]::ParameterName, 'remove destination before import')
            [CompletionResult]::new('--strip-components', 'strip-components', [CompletionResultType]::ParameterName, 'strip components')
            break
        }
        'chezmoi;init' {
            [CompletionResult]::new('-a', 'a', [CompletionResultType]::ParameterName, 'update destination directory')
            [CompletionResult]::new('--apply', 'apply', [CompletionResultType]::ParameterName, 'update destination directory')
            [CompletionResult]::new('-d', 'd', [CompletionResultType]::ParameterName, 'create a shallow clone')
            [CompletionResult]::new('--depth', 'depth', [CompletionResultType]::ParameterName, 'create a shallow clone')
            [CompletionResult]::new('--ignore-encrypted', 'ignore-encrypted', [CompletionResultType]::ParameterName, 'ignore encrypted files')
            [CompletionResult]::new('--one-shot', 'one-shot', [CompletionResultType]::ParameterName, 'one shot')
            [CompletionResult]::new('-p', 'p', [CompletionResultType]::ParameterName, 'purge config and source directories')
            [CompletionResult]::new('--purge', 'purge', [CompletionResultType]::ParameterName, 'purge config and source directories')
            [CompletionResult]::new('-P', 'P', [CompletionResultType]::ParameterName, 'purge chezmoi binary')
            [CompletionResult]::new('--purge-binary', 'purge-binary', [CompletionResultType]::ParameterName, 'purge chezmoi binary')
            break
        }
        'chezmoi;managed' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            break
        }
        'chezmoi;merge' {
            break
        }
        'chezmoi;purge' {
            [CompletionResult]::new('-P', 'P', [CompletionResultType]::ParameterName, 'purge chezmoi executable')
            [CompletionResult]::new('--binary', 'binary', [CompletionResultType]::ParameterName, 'purge chezmoi executable')
            break
        }
        'chezmoi;remove' {
            break
        }
        'chezmoi;secret' {
            [CompletionResult]::new('keyring', 'keyring', [CompletionResultType]::ParameterValue, 'Interact with keyring')
            break
        }
        'chezmoi;secret;keyring' {
            [CompletionResult]::new('get', 'get', [CompletionResultType]::ParameterValue, 'Get a value from keyring')
            [CompletionResult]::new('set', 'set', [CompletionResultType]::ParameterValue, 'Set a value in keyring')
            break
        }
        'chezmoi;secret;keyring;get' {
            break
        }
        'chezmoi;secret;keyring;set' {
            break
        }
        'chezmoi;source-path' {
            break
        }
        'chezmoi;state' {
            [CompletionResult]::new('dump', 'dump', [CompletionResultType]::ParameterValue, 'Generate a dump of the persistent state')
            [CompletionResult]::new('reset', 'reset', [CompletionResultType]::ParameterValue, 'Reset the persistent state')
            break
        }
        'chezmoi;state;dump' {
            break
        }
        'chezmoi;state;reset' {
            break
        }
        'chezmoi;status' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
        'chezmoi;unmanaged' {
            break
        }
        'chezmoi;update' {
            [CompletionResult]::new('-a', 'a', [CompletionResultType]::ParameterName, 'apply after pulling')
            [CompletionResult]::new('--apply', 'apply', [CompletionResultType]::ParameterName, 'apply after pulling')
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
        'chezmoi;verify' {
            [CompletionResult]::new('-i', 'i', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('--include', 'include', [CompletionResultType]::ParameterName, 'include entry types')
            [CompletionResult]::new('-r', 'r', [CompletionResultType]::ParameterName, 'recursive')
            [CompletionResult]::new('--recursive', 'recursive', [CompletionResultType]::ParameterName, 'recursive')
            break
        }
    })
    $completions.Where{ $_.CompletionText -like "$wordToComplete*" } |
        Sort-Object -Property ListItemText
}