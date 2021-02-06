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
            [CompletionResult]::new('--debug', 'debug', [CompletionResultType]::ParameterName, 'write debug logs')
            [CompletionResult]::new('-D', 'D', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('--destination', 'destination', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('-n', 'n', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--dry-run', 'dry-run', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--follow', 'follow', [CompletionResultType]::ParameterName, 'follow symlinks')
            [CompletionResult]::new('-N', 'N', [CompletionResultType]::ParameterName, 'don''t run scripts')
            [CompletionResult]::new('--noscript', 'noscript', [CompletionResultType]::ParameterName, 'don''t run scripts')
            [CompletionResult]::new('--remove', 'remove', [CompletionResultType]::ParameterName, 'remove targets')
            [CompletionResult]::new('-S', 'S', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--source', 'source', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('-v', 'v', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('--verbose', 'verbose', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('add', 'add', [CompletionResultType]::ParameterValue, 'Add an existing file, directory, or symlink to the source state')
            [CompletionResult]::new('apply', 'apply', [CompletionResultType]::ParameterValue, 'Update the destination directory to match the target state')
            [CompletionResult]::new('archive', 'archive', [CompletionResultType]::ParameterValue, 'Write a tar archive of the target state to stdout')
            [CompletionResult]::new('cat', 'cat', [CompletionResultType]::ParameterValue, 'Print the target contents of a file or symlink')
            [CompletionResult]::new('cd', 'cd', [CompletionResultType]::ParameterValue, 'Launch a shell in the source directory')
            [CompletionResult]::new('chattr', 'chattr', [CompletionResultType]::ParameterValue, 'Change the attributes of a target in the source state')
            [CompletionResult]::new('completion', 'completion', [CompletionResultType]::ParameterValue, 'Generate shell completion code for the specified shell (bash, fish, or zsh)')
            [CompletionResult]::new('data', 'data', [CompletionResultType]::ParameterValue, 'Print the template data')
            [CompletionResult]::new('diff', 'diff', [CompletionResultType]::ParameterValue, 'Print the diff between the target state and the destination state')
            [CompletionResult]::new('docs', 'docs', [CompletionResultType]::ParameterValue, 'Print documentation')
            [CompletionResult]::new('doctor', 'doctor', [CompletionResultType]::ParameterValue, 'Check your system for potential problems')
            [CompletionResult]::new('dump', 'dump', [CompletionResultType]::ParameterValue, 'Write a dump of the target state to stdout')
            [CompletionResult]::new('edit', 'edit', [CompletionResultType]::ParameterValue, 'Edit the source state of a target')
            [CompletionResult]::new('edit-config', 'edit-config', [CompletionResultType]::ParameterValue, 'Edit the configuration file')
            [CompletionResult]::new('execute-template', 'execute-template', [CompletionResultType]::ParameterValue, 'Write the result of executing the given template(s) to stdout')
            [CompletionResult]::new('forget', 'forget', [CompletionResultType]::ParameterValue, 'Remove a target from the source state')
            [CompletionResult]::new('git', 'git', [CompletionResultType]::ParameterValue, 'Run git in the source directory')
            [CompletionResult]::new('help', 'help', [CompletionResultType]::ParameterValue, 'Print help about a command')
            [CompletionResult]::new('hg', 'hg', [CompletionResultType]::ParameterValue, 'Run mercurial in the source directory')
            [CompletionResult]::new('import', 'import', [CompletionResultType]::ParameterValue, 'Import a tar archive into the source state')
            [CompletionResult]::new('init', 'init', [CompletionResultType]::ParameterValue, 'Setup the source directory and update the destination directory to match the target state')
            [CompletionResult]::new('managed', 'managed', [CompletionResultType]::ParameterValue, 'List the managed files in the destination directory')
            [CompletionResult]::new('merge', 'merge', [CompletionResultType]::ParameterValue, 'Perform a three-way merge between the destination state, the source state, and the target state')
            [CompletionResult]::new('purge', 'purge', [CompletionResultType]::ParameterValue, 'Purge all of chezmoi''s configuration and data')
            [CompletionResult]::new('remove', 'remove', [CompletionResultType]::ParameterValue, 'Remove a target from the source state and the destination directory')
            [CompletionResult]::new('secret', 'secret', [CompletionResultType]::ParameterValue, 'Interact with a secret manager')
            [CompletionResult]::new('source', 'source', [CompletionResultType]::ParameterValue, 'Run the source version control system command in the source directory')
            [CompletionResult]::new('source-path', 'source-path', [CompletionResultType]::ParameterValue, 'Print the path of a target in the source state')
            [CompletionResult]::new('unmanaged', 'unmanaged', [CompletionResultType]::ParameterValue, 'List the unmanaged files in the destination directory')
            [CompletionResult]::new('update', 'update', [CompletionResultType]::ParameterValue, 'Pull changes from the source VCS and apply any changes')
            [CompletionResult]::new('upgrade', 'upgrade', [CompletionResultType]::ParameterValue, 'Upgrade chezmoi to the latest released version')
            [CompletionResult]::new('verify', 'verify', [CompletionResultType]::ParameterValue, 'Exit with success if the destination state matches the target state, fail otherwise')
            break
        }
        'chezmoi;add' {
            break
        }
        'chezmoi;apply' {
            break
        }
        'chezmoi;archive' {
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
            [CompletionResult]::new('--debug', 'debug', [CompletionResultType]::ParameterName, 'write debug logs')
            [CompletionResult]::new('-D', 'D', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('--destination', 'destination', [CompletionResultType]::ParameterName, 'destination directory')
            [CompletionResult]::new('-n', 'n', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--dry-run', 'dry-run', [CompletionResultType]::ParameterName, 'dry run')
            [CompletionResult]::new('--follow', 'follow', [CompletionResultType]::ParameterName, 'follow symlinks')
            [CompletionResult]::new('-h', 'h', [CompletionResultType]::ParameterName, 'help for completion')
            [CompletionResult]::new('--help', 'help', [CompletionResultType]::ParameterName, 'help for completion')
            [CompletionResult]::new('-N', 'N', [CompletionResultType]::ParameterName, 'don''t run scripts')
            [CompletionResult]::new('--noscript', 'noscript', [CompletionResultType]::ParameterName, 'don''t run scripts')
            [CompletionResult]::new('-o', 'o', [CompletionResultType]::ParameterName, 'output filename')
            [CompletionResult]::new('--output', 'output', [CompletionResultType]::ParameterName, 'output filename')
            [CompletionResult]::new('--remove', 'remove', [CompletionResultType]::ParameterName, 'remove targets')
            [CompletionResult]::new('-S', 'S', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('--source', 'source', [CompletionResultType]::ParameterName, 'source directory')
            [CompletionResult]::new('-v', 'v', [CompletionResultType]::ParameterName, 'verbose')
            [CompletionResult]::new('--verbose', 'verbose', [CompletionResultType]::ParameterName, 'verbose')
            break
        }
        'chezmoi;data' {
            break
        }
        'chezmoi;diff' {
            break
        }
        'chezmoi;docs' {
            break
        }
        'chezmoi;doctor' {
            break
        }
        'chezmoi;dump' {
            break
        }
        'chezmoi;edit' {
            break
        }
        'chezmoi;edit-config' {
            break
        }
        'chezmoi;execute-template' {
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
        'chezmoi;hg' {
            break
        }
        'chezmoi;import' {
            break
        }
        'chezmoi;init' {
            break
        }
        'chezmoi;managed' {
            break
        }
        'chezmoi;merge' {
            break
        }
        'chezmoi;purge' {
            break
        }
        'chezmoi;remove' {
            break
        }
        'chezmoi;secret' {
            [CompletionResult]::new('bitwarden', 'bitwarden', [CompletionResultType]::ParameterValue, 'Execute the Bitwarden CLI (bw)')
            [CompletionResult]::new('generic', 'generic', [CompletionResultType]::ParameterValue, 'Execute a generic secret command')
            [CompletionResult]::new('gopass', 'gopass', [CompletionResultType]::ParameterValue, 'Execute the gopass CLI')
            [CompletionResult]::new('keepassxc', 'keepassxc', [CompletionResultType]::ParameterValue, 'Execute the KeePassXC CLI (keepassxc-cli)')
            [CompletionResult]::new('keyring', 'keyring', [CompletionResultType]::ParameterValue, 'Interact with keyring')
            [CompletionResult]::new('lastpass', 'lastpass', [CompletionResultType]::ParameterValue, 'Execute the LastPass CLI (lpass)')
            [CompletionResult]::new('onepassword', 'onepassword', [CompletionResultType]::ParameterValue, 'Execute the 1Password CLI (op)')
            [CompletionResult]::new('pass', 'pass', [CompletionResultType]::ParameterValue, 'Execute the pass CLI')
            [CompletionResult]::new('vault', 'vault', [CompletionResultType]::ParameterValue, 'Execute the Hashicorp Vault CLI (vault)')
            break
        }
        'chezmoi;secret;bitwarden' {
            break
        }
        'chezmoi;secret;generic' {
            break
        }
        'chezmoi;secret;gopass' {
            break
        }
        'chezmoi;secret;keepassxc' {
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
        'chezmoi;secret;lastpass' {
            break
        }
        'chezmoi;secret;onepassword' {
            break
        }
        'chezmoi;secret;pass' {
            break
        }
        'chezmoi;secret;vault' {
            break
        }
        'chezmoi;source' {
            break
        }
        'chezmoi;source-path' {
            break
        }
        'chezmoi;unmanaged' {
            break
        }
        'chezmoi;update' {
            break
        }
        'chezmoi;upgrade' {
            break
        }
        'chezmoi;verify' {
            break
        }
    })
    $completions.Where{ $_.CompletionText -like "$wordToComplete*" } |
        Sort-Object -Property ListItemText
}