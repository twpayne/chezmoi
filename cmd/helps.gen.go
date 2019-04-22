//go:generate go run github.com/twpayne/chezmoi/internal/extract-helps -i ../REFERENCE.md -o helps.gen.go

package cmd

type help struct {
	long    string
	example string
}

var helps = map[string]help{
	"add": {
		long:    "Add targets to the source state. If any target is already in the source state,\nthen its source state is replaced with its current state in the destination\ndirectory. The \"add\" command accepts additional flags:\n\n\"-e\", \"--empty\"\n\nSet the \"empty\" attribute on added files.\n\n\"-x\", \"--exact\"\n\nSet the \"exact\" attribute on added directories.\n\n\"-p\", \"--prompt\"\n\nInteractively prompt before adding each file.\n\n\"-r\", \"--recursive\"\n\nRecursively add all files, directories, and symlinks.\n\n\"-T\", \"--template\"\n\nSet the \"template\" attribute on added files and symlinks. In addition, \"chezmoi\"\nattempts to automatically generate the template by replacing any template data\nvalues with the equivalent template data keys. Longer subsitutions occur before\nshorter ones.\n",
		example: "  chezmoi add ~/.bashrc\n  chezmoi add --template ~/.gitconfig\n  chezmoi add --recursive ~/.vim",
	},
	"apply": {
		long:    "Ensure that targets are in the target state, updating them if necessary. If no\ntargets are specified, the state of all targets are ensured.\n",
		example: "  chezmoi apply\n  chezmoi apply --dry-run --verbose\n  chezmoi apply ~/.bashrc",
	},
	"archive": {
		long:    "Write a tar archive of the target state to stdout. This can be piped into \"tar\"\nto inspect the target state.\n",
		example: "  chezmoi archive | tar tvf -",
	},
	"cat": {
		long:    "Write the target state of targets  to stdout. targets must be files or symlinks.\nFor files, the target file contents are written. For symlinks, the target target\nis written.\n",
		example: "  chezmoi cat ~/.bashrc",
	},
	"cd": {
		long:    "Launch a shell in the source directory.\n",
		example: "  chezmoi cd",
	},
	"chattr": {
		long:    "Change the attributes of targets. attributes specifies which attributes to\nmodify. Add attributes by specifying them or their abbreviations directly,\noptionally prefixed with a plus sign (\"+\"). Remove attributes by prefixing them\nor their attributes with the string \"no\" or a minus sign (\"-\"). The available\nattributes and their abbreviations are:\n\n  Attribute    Abbreviation \n  \"empty\"      \"e\"          \n  \"encrypted\"  none         \n  \"exact\"      none         \n  \"executable\" \"x\"          \n  \"private\"    \"p\"          \n  \"template\"   \"t\"          \n  \nMultiple attributes modifications may be specified by separating them with a\ncomma (\",\").\n",
		example: "  chezmoi chattr template ~/.bashrc\n  chezmoi chattr noempty ~/.profile",
	},
	"completion": {
		long:    "Output shell completion code for the specified shell (\"bash\" or \"zsh\").\n",
		example: "  chezmoi completion bash\n  chezmoi completion zsh",
	},
	"data": {
		long:    "Write the computed template data in JSON format to stdout. The \"data\" command\naccepts additional flags:\n\n\"-f\", \"--format\" format\n\nPrint the computed template data in the given format. The accepted formats are\n\"json\" (JSON), \"toml\" (TOML), and \"yaml\" (YAML).\n",
		example: "  chezmoi data\n  chezmoi data --format=yaml",
	},
	"diff": {
		long:    "Print the approximate shell commands required to ensure that targets in the\ndestination directory match the target state. If no targets are specifed,\nprint the commands required for all targets. It is equivalent to \"chezmoi apply\n--dry-run --verbose\".\n",
		example: "  chezmoi diff\n  chezmoi diff ~/.bashrc",
	},
	"doctor": {
		long:    "Check for potential problems.\n",
		example: "  chezmoi doctor",
	},
	"dump": {
		long:    "Dump the target state in JSON format. If no targets are specified, then the\nentire target state. The \"dump\" command accepts additional arguments:\n\n\"-f\" / \"--format\" format\n\nPrint the target state in the given format. The accepted formats are \"json\"\n(JSON) and \"yaml\" (YAML).\n",
		example: "  chezmoi dump ~/.bashrc\n  chezmoi dump --format=yaml",
	},
	"edit": {
		long:    "Edit the source state of targets, which must be files or symlinks. The \"edit\"\ncommand accepts additional arguments:\n\n\"-a\", \"--apply\"\n\nApply target immediately after editing.\n\n\"-d\", \"--diff\"\n\nPrint the difference between the target state and the actual state after\nediting.\n\n\"-p\" / \"--prompt\"\n\nPrompt before applying each target.\n",
		example: "  chezmoi edit ~/.bashrc\n  chezmoi edit --apply --prompt ~/.bashrc",
	},
	"edit-config": {
		long:    "Edit the configuration file.\n",
		example: "  chezmoi edit-config",
	},
	"forget": {
		long:    "Remove targets from the source state, i.e. stop managing them.\n",
		example: "  chezmoi forget ~/.bashrc",
	},
	"help": {
		long: "Print the help associated with command.\n",
	},
	"import": {
		long: "FIXME document\n\n\"-x\" / \"--exact\"\n\n\"-r\", \"--remove-destination\"\n\n\"--strip-components\"\n",
	},
	"init": {
		long:    "Setup the source directory and update the destination directory to match the\ntarget state. If repo is given then it is checked out into the source directory,\notherwise a new repository is initialized in the source directory. If a file\ncalled \".chezmoi.format.tmpl\" exists, where \"format\" is one of the supported\nfile formats (e.g. \"json\", \"toml\", or \"yaml\") then a new configuration file is\ncreated using that file as a template. Finally, if the \"--apply\" flag is passed,\n\"chezmoi apply\" is run.\n",
		example: "  chezmoi init https://github.com/user/dotfiles.git",
	},
	"merge": {
		long:    "Perform a three-way merge between the destination state, the source state, and\nthe target state. The merge tool is defined by the \"merge.command\" configuration\nvariable, and defaults to \"vimdiff\". If multiple targets are specified the merge\ntool is invoked for each target. If the target state cannot be computed (for\nexample if source is a template containing errors or an encrypted file that\ncannot be decrypted) a two-way merge is performed instead.\n",
		example: "  chezmoi merge ~/.bashrc",
	},
	"remove": {
		long: "Remove targets from both the source state and the destination directory.\n",
	},
	"secret": {
		long: "Interact with a secret manager. See the \"Secret managers\" section for details.\n",
	},
	"source": {
		long:    "Execute the source version control system in the source directory with args.\nNote that any flags for the source version control system must be sepeated with\na \"--\" to stop \"chezmoi\" from reading them.\n",
		example: "  chezmoi source init\n  chezmoi source add .\n  chezmoi source commit -- -m \"Initial commit\"",
	},
	"source-path": {
		long:    "Print the path to each target's source state. If no targets are specified then\nprint the source directory.\n",
		example: "  chezmoi source-path ~/.bashrc",
	},
	"unmanaged": {
		long:    "List all unmanaged files in the destination directory.\n",
		example: "  chezmoi unmanaged",
	},
	"update": {
		long:    "Pull changes from the source VCS abd apply any changes.\n",
		example: "  chezmoi update",
	},
	"verify": {
		long:    "Verify that all targets match their target state. \"chezmoi\" exits with code 0\n(success) if all targets match their target state, or 1 (failure) otherwise. If\nno targets are specified then all targets are checked.\n",
		example: "  chezmoi verify\n  chezmoi verify ~/.bashrc\n  \nEditor configuration\n\nThe \"edit\" and \"edit-config\" commands use the editor specified by the \"VISUAL\"\nenvironment variable, the \"EDITOR\" environment variable, or \"vi\", whichever is\nspecified first.\n\nUmask\n\nFIXME document\n\nTemplates\n\n\"chezmoi\" provides the following automatically populated variables:\n\n  Variable                Value                                                                                     \n  \".chezmoi.arch\"         Architecture, e.g. \"amd64\", \"arm\", etc. as returned by runtime.GOARCH.                    \n  \".chezmoi.fullHostname\" The full hostname of the machine \"chezmoi\" is running on.                                 \n  \".chezmoi.group\"        The group of the user running \"chezmoi\".                                                  \n  \".chezmoi.homedir\"      The home directory of the user running \"chezmoi\".                                         \n  \".chezmoi.hostname\"     The hostname of the machine \"chezmoi\" is running on, up to the first \".\".                 \n  \".chezmoi.os\"           Operating system, e.g. \"darwin\", \"linux\", etc. as returned by runtime.GOOS.               \n  \".chezmoi.osRelease\"    The information from \"/etc/os-release\", Linux only, run \"chezmoi data\" to see its output. \n  \".chezmoi.username\"     The username of the user running \"chezmoi\".                                               \n  \nSecret managers\n\nFIXME document\n\nBitwarden\n\nFIXME document\n\nKeyring\n\nFIXME document\n\nLastPass\n\nFIXME document\n\n1Password\n\nFIXME document\n\npass\n\nFIXME document\n\nVault\n\nFIXME document\n\nGeneric\n\nvim: spell spelllang=en\n",
	},
}
