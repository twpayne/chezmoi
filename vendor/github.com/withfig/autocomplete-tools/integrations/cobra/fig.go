package cobracompletefig

import (
	"fmt"
	"strings"
)

type BaseSuggestion struct {
	displayName string
	description string
	isDangerous bool
	hidden      bool
}

type SuggestionType int

const (
	FOLDER SuggestionType = iota
	FILE
	ARG
	SUBCOMMAND
	OPTION
	SPECIAL
	SHORTCUT
)

type Suggestion struct {
	*BaseSuggestion
	name           string
	suggestionType SuggestionType
}

type Names []string

type Subcommand struct {
	*BaseSuggestion
	name        Names
	subcommands Subcommands
	options     Options
	args        Args
}

type Subcommands []Subcommand
type Options []Option
type Args []Arg

type Spec struct {
	*Subcommand
	name string
}

type Option struct {
	*BaseSuggestion
	name         Names
	args         Args
	isRequired   bool
	isRepeatable bool
	isPersistent bool
	exclusiveOn  []string
	dependsOn    []string
}

type Arg struct {
	name        string
	description string
	isDangerous bool
	suggestions []Suggestion //or maybe []string
	template    []Template
	isVariadic  bool
	isOptional  bool
	isCommand   bool
	isModule    bool
	isScript    bool
	defaultVal  string //This is default in the fig spec, but that word is reserved in go
}

type Template int

const (
	FOLDERS Template = iota
	FILEPATHS
)

func sanitize(str string) string {
	sanitized := str
	sanitized = strings.ReplaceAll(sanitized, "\"", "\\\"")
	sanitized = strings.ReplaceAll(sanitized, "'", "\\'")
	sanitized = strings.ReplaceAll(sanitized, "\n", " ")
	return sanitized
}

func (names *Names) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, name := range *names {
		sb.WriteString(fmt.Sprintf(`"%v",`, name))
	}
	sb.WriteString("]")
	return sb.String()
}

func (spec *Spec) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("const completionSpec: Fig.Spec = {")
	sb.WriteString(fmt.Sprintf(`name: "%v",`, spec.name))
	sb.WriteString(fmt.Sprintf(`description: "%v",`, sanitize(spec.description)))
	if len(spec.subcommands) > 0 {
		sb.WriteString(fmt.Sprintf(`subcommands: %v,`, spec.subcommands.ToTypescript()))
	}
	if len(spec.options) > 0 {
		sb.WriteString(fmt.Sprintf(`options: %v,`, spec.options.ToTypescript()))
	}
	if len(spec.args) > 0 {
		sb.WriteString(fmt.Sprintf(`args: %v,`, spec.args.ToTypescript()))
	}
	sb.WriteString("}; export default completionSpec;")
	return sb.String()
}

func (subcommands *Subcommands) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, subcommand := range *subcommands {
		sb.WriteString(fmt.Sprintf(`%v,`, subcommand.ToTypescript()))
	}
	sb.WriteString("]")
	return sb.String()
}

func (subcommand *Subcommand) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString(fmt.Sprintf(`name: %v,`, subcommand.name.ToTypescript()))
	sb.WriteString(fmt.Sprintf(`description: "%v",`, sanitize(subcommand.description)))
	if subcommand.hidden {
		sb.WriteString(fmt.Sprintf(`hidden: %t,`, subcommand.hidden))
	}
	if len(subcommand.subcommands) > 0 {
		sb.WriteString(fmt.Sprintf(`subcommands: %v,`, subcommand.subcommands.ToTypescript()))
	}
	if len(subcommand.options) > 0 {
		sb.WriteString(fmt.Sprintf(`options: %v,`, subcommand.options.ToTypescript()))
	}
	if len(subcommand.args) > 0 {
		sb.WriteString(fmt.Sprintf(`args: %v,`, subcommand.args.ToTypescript()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (options *Options) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, option := range *options {
		sb.WriteString(fmt.Sprintf(`%v,`, option.ToTypescript()))
	}
	sb.WriteString("]")
	return sb.String()
}

func (option *Option) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString(fmt.Sprintf(`name: %v,`, option.name.ToTypescript()))
	sb.WriteString(fmt.Sprintf(`description: "%v",`, sanitize(option.description)))
	if option.isPersistent {
		sb.WriteString(fmt.Sprintf(`isPersistent: %t,`, option.isPersistent))
	}
	if option.hidden {
		sb.WriteString(fmt.Sprintf(`hidden: %t,`, option.hidden))
	}
	if option.isRepeatable {
		sb.WriteString(fmt.Sprintf(`isRepeatable: %t,`, option.isRepeatable))
	}
	if option.displayName != "" {
		sb.WriteString(fmt.Sprintf(`displayName: "%v",`, sanitize(option.displayName)))
	}
	if len(option.args) > 0 {
		sb.WriteString(fmt.Sprintf(`args: %v,`, option.args.ToTypescript()))
	}
	if option.isRequired {
		sb.WriteString(`isRequired: true,`)
	}
	sb.WriteString("}")
	return sb.String()
}

func (args *Args) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("[")
	for _, arg := range *args {
		sb.WriteString(fmt.Sprintf(`%v,`, arg.ToTypescript()))
	}
	sb.WriteString("]")
	return sb.String()
}

func (arg *Arg) ToTypescript() string {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString(fmt.Sprintf(`name: "%v",`, arg.name))
	if arg.description != "" {
		sb.WriteString(fmt.Sprintf(`description: "%v",`, sanitize(arg.description)))
	}
	if arg.defaultVal != "" {
		sb.WriteString(fmt.Sprintf(`default: "%v",`, arg.defaultVal))
	}
	if len(arg.template) > 0 {
		sb.WriteString(`template: [`)
		for _, val := range arg.template {
			switch val {
			case FOLDERS:
				sb.WriteString(`"folders",`)
			case FILEPATHS:
				sb.WriteString(`"filepaths",`)
			}
		}
		sb.WriteString(`],`)
	}
	sb.WriteString("}")
	return sb.String()
}
