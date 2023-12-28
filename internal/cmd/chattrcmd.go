package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type chattrCmdConfig struct {
	recursive bool
}

type boolModifier int

const (
	boolModifierSet            boolModifier = 1
	boolModifierLeaveUnchanged boolModifier = 0
	boolModifierClear          boolModifier = -1
)

type conditionModifier int

const (
	conditionModifierLeaveUnchanged conditionModifier = iota
	conditionModifierClearOnce
	conditionModifierSetOnce
	conditionModifierClearOnChange
	conditionModifierSetOnChange
)

type orderModifier int

const (
	orderModifierSetBefore      orderModifier = -2
	orderModifierClearBefore    orderModifier = -1
	orderModifierLeaveUnchanged orderModifier = 0
	orderModifierClearAfter     orderModifier = 1
	orderModifierSetAfter       orderModifier = 2
)

type sourceFileTypeModifier int

const (
	sourceFileTypeModifierLeaveUnchanged sourceFileTypeModifier = iota
	sourceFileTypeModifierSetCreate
	sourceFileTypeModifierClearCreate
	sourceFileTypeModifierSetModify
	sourceFileTypeModifierClearModify
	sourceFileTypeModifierSetRemove
	sourceFileTypeModifierClearRemove
	sourceFileTypeModifierSetScript
	sourceFileTypeModifierClearScript
	sourceFileTypeModifierSetSymlink
	sourceFileTypeModifierClearSymlink
)

type modifier struct {
	sourceFileType sourceFileTypeModifier
	condition      conditionModifier
	empty          boolModifier
	encrypted      boolModifier
	exact          boolModifier
	executable     boolModifier
	external       boolModifier
	order          orderModifier
	private        boolModifier
	readOnly       boolModifier
	remove         boolModifier
	template       boolModifier
}

func (c *Config) newChattrCmd() *cobra.Command {
	chattrCmd := &cobra.Command{
		Use:               "chattr attributes target...",
		Short:             "Change the attributes of a target in the source state",
		Long:              mustLongHelp("chattr"),
		Example:           example("chattr"),
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: c.chattrCmdValidArgs,
		RunE:              c.makeRunEWithSourceState(c.runChattrCmd),
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			requiresSourceDirectory,
		),
	}

	flags := chattrCmd.Flags()
	flags.BoolVarP(&c.chattr.recursive, "recursive", "r", c.chattr.recursive, "Recurse into subdirectories")

	return chattrCmd
}

// chattrCmdValidArgs returns the completions for the chattr command.
func (c *Config) chattrCmdValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		prefixes := []string{"", "-", "+", "no"}
		attributes := []string{
			"after",
			"before",
			"create",
			"empty",
			"encrypted",
			"exact",
			"executable",
			"external",
			"modify",
			"once",
			"onchange",
			"private",
			"readonly",
			"remove",
			"script",
			"symlink",
			"template",
		}
		validModifiers := make([]string, 0, len(prefixes)*len(attributes))
		for _, prefix := range prefixes {
			for _, attribute := range attributes {
				modifier := prefix + attribute
				validModifiers = append(validModifiers, modifier)
			}
		}

		modifiers := strings.Split(toComplete, ",")
		modifierToComplete := modifiers[len(modifiers)-1]
		completionPrefix := toComplete[:len(toComplete)-len(modifierToComplete)]
		var completions []string
		for _, modifier := range validModifiers {
			if strings.HasPrefix(modifier, modifierToComplete) {
				completion := completionPrefix + modifier
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	default:
		return c.targetValidArgs(cmd, args, toComplete)
	}
}

func (c *Config) runChattrCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// LATER should the core functionality of chattr move to chezmoi.SourceState?

	m, err := parseModifier(args[0])
	if err != nil {
		return err
	}

	targetRelPaths, err := c.targetRelPaths(sourceState, args[1:], targetRelPathsOptions{
		mustBeManaged: true,
		recursive:     c.chattr.recursive,
	})
	if err != nil {
		return err
	}

	// Sort targets in reverse so we update children before their parent
	// directories.
	sort.Sort(sort.Reverse(targetRelPaths))

	encryptedSuffix := sourceState.Encryption().EncryptedSuffix()
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		sourceRelPath := sourceStateEntry.SourceRelPath()
		parentSourceRelPath, fileSourceRelPath := sourceRelPath.Split()
		parentRelPath := parentSourceRelPath.RelPath()
		fileRelPath := fileSourceRelPath.RelPath()
		switch sourceStateEntry := sourceStateEntry.(type) {
		case *chezmoi.SourceStateDir:
			relPath := m.modifyDirAttr(sourceStateEntry.Attr).SourceName()
			if newBaseNameRelPath := chezmoi.NewRelPath(relPath); newBaseNameRelPath != fileRelPath {
				oldSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, fileRelPath)
				newSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, newBaseNameRelPath)
				if err := c.sourceSystem.Rename(oldSourceAbsPath, newSourceAbsPath); err != nil {
					return err
				}
			}
		case *chezmoi.SourceStateFile:
			newAttr := m.modifyFileAttr(sourceStateEntry.Attr)
			newBaseNameRelPath := chezmoi.NewRelPath(newAttr.SourceName(encryptedSuffix))
			oldSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, fileRelPath)
			newSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, newBaseNameRelPath)
			switch encryptedBefore, encryptedAfter := sourceStateEntry.Attr.Encrypted, newAttr.Encrypted; {
			case encryptedBefore && !encryptedAfter:
				// Write the plaintext and then remove the ciphertext.
				plaintext, err := sourceStateEntry.Contents()
				if err != nil {
					return err
				}
				if err := c.sourceSystem.WriteFile(newSourceAbsPath, plaintext, 0o666&^c.Umask); err != nil {
					return err
				}
				if err := c.sourceSystem.Remove(oldSourceAbsPath); err != nil {
					return err
				}
			case !encryptedBefore && encryptedAfter:
				// Write the ciphertext and then remove the plaintext.
				plaintext, err := sourceStateEntry.Contents()
				if err != nil {
					return err
				}
				ciphertext, err := sourceState.Encryption().Encrypt(plaintext)
				if err != nil {
					return err
				}
				if err := c.sourceSystem.WriteFile(newSourceAbsPath, ciphertext, 0o666&^c.Umask); err != nil {
					return err
				}
				if err := c.sourceSystem.Remove(oldSourceAbsPath); err != nil {
					return err
				}
			case newBaseNameRelPath != fileRelPath:
				// Contents have not changed so a rename is sufficient.
				if err := c.sourceSystem.Rename(oldSourceAbsPath, newSourceAbsPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// modify returns the modified value of b.
func (m boolModifier) modify(b bool) bool {
	switch m {
	case boolModifierSet:
		return true
	case boolModifierLeaveUnchanged:
		return b
	case boolModifierClear:
		return false
	default:
		panic(fmt.Sprintf("%d: unknown bool modifier", m))
	}
}

// modify returns the modified value of condition.
func (m conditionModifier) modify(condition chezmoi.ScriptCondition) chezmoi.ScriptCondition {
	switch m {
	case conditionModifierLeaveUnchanged:
		return condition
	case conditionModifierClearOnce:
		if condition == chezmoi.ScriptConditionOnce {
			return chezmoi.ScriptConditionAlways
		}
		return condition
	case conditionModifierSetOnce:
		return chezmoi.ScriptConditionOnce
	case conditionModifierClearOnChange:
		if condition == chezmoi.ScriptConditionOnChange {
			return chezmoi.ScriptConditionAlways
		}
		return condition
	case conditionModifierSetOnChange:
		return chezmoi.ScriptConditionOnChange
	default:
		panic(fmt.Sprintf("%d: unknown order modifier", m))
	}
}

// modify returns the modified value of order.
func (m orderModifier) modify(order chezmoi.ScriptOrder) chezmoi.ScriptOrder {
	switch m {
	case orderModifierSetBefore:
		return chezmoi.ScriptOrderBefore
	case orderModifierClearBefore:
		if order == chezmoi.ScriptOrderBefore {
			return chezmoi.ScriptOrderDuring
		}
		return order
	case orderModifierLeaveUnchanged:
		return order
	case orderModifierClearAfter:
		if order == chezmoi.ScriptOrderAfter {
			return chezmoi.ScriptOrderDuring
		}
		return order
	case orderModifierSetAfter:
		return chezmoi.ScriptOrderAfter
	default:
		panic(fmt.Sprintf("%d: unknown order modifier", m))
	}
}

// modify returns the modified value of type.
func (m sourceFileTypeModifier) modify(sourceFileType chezmoi.SourceFileTargetType) chezmoi.SourceFileTargetType {
	switch m {
	case sourceFileTypeModifierLeaveUnchanged:
		return sourceFileType
	case sourceFileTypeModifierSetCreate:
		return chezmoi.SourceFileTypeCreate
	case sourceFileTypeModifierClearCreate:
		if sourceFileType == chezmoi.SourceFileTypeCreate {
			return chezmoi.SourceFileTypeFile
		}
		return sourceFileType
	case sourceFileTypeModifierSetRemove:
		return chezmoi.SourceFileTypeRemove
	case sourceFileTypeModifierClearRemove:
		if sourceFileType == chezmoi.SourceFileTypeRemove {
			return chezmoi.SourceFileTypeFile
		}
		return sourceFileType
	case sourceFileTypeModifierSetModify:
		return chezmoi.SourceFileTypeModify
	case sourceFileTypeModifierClearModify:
		if sourceFileType == chezmoi.SourceFileTypeModify {
			return chezmoi.SourceFileTypeFile
		}
		return sourceFileType
	case sourceFileTypeModifierSetScript:
		return chezmoi.SourceFileTypeScript
	case sourceFileTypeModifierClearScript:
		if sourceFileType == chezmoi.SourceFileTypeScript {
			return chezmoi.SourceFileTypeFile
		}
		return sourceFileType
	case sourceFileTypeModifierSetSymlink:
		return chezmoi.SourceFileTypeSymlink
	case sourceFileTypeModifierClearSymlink:
		if sourceFileType == chezmoi.SourceFileTypeSymlink {
			return chezmoi.SourceFileTypeFile
		}
		return sourceFileType
	default:
		panic(fmt.Sprintf("%d: unknown type modifier", m))
	}
}

// parseModifier parses the modifier from s.
func parseModifier(s string) (*modifier, error) {
	m := &modifier{}
	for _, modifierStr := range strings.Split(s, ",") {
		modifierStr = strings.TrimSpace(modifierStr)
		if modifierStr == "" {
			continue
		}
		var bm boolModifier
		var attribute string
		switch {
		case modifierStr[0] == '-':
			bm = boolModifierClear
			attribute = modifierStr[1:]
		case modifierStr[0] == '+':
			bm = boolModifierSet
			attribute = modifierStr[1:]
		case strings.HasPrefix(modifierStr, "no"):
			bm = boolModifierClear
			attribute = modifierStr[2:]
		default:
			bm = boolModifierSet
			attribute = modifierStr
		}
		switch attribute {
		case "after", "a":
			switch bm {
			case boolModifierClear:
				m.order = orderModifierClearAfter
			case boolModifierLeaveUnchanged:
				m.order = orderModifierLeaveUnchanged
			case boolModifierSet:
				m.order = orderModifierSetAfter
			}
		case "before", "b":
			switch bm {
			case boolModifierClear:
				m.order = orderModifierClearBefore
			case boolModifierLeaveUnchanged:
				m.order = orderModifierLeaveUnchanged
			case boolModifierSet:
				m.order = orderModifierSetBefore
			}
		case "create":
			switch bm {
			case boolModifierClear:
				m.sourceFileType = sourceFileTypeModifierClearCreate
			case boolModifierSet:
				m.sourceFileType = sourceFileTypeModifierSetCreate
			}
		case "empty", "e":
			m.empty = bm
		case "encrypted":
			m.encrypted = bm
		case "exact":
			m.exact = bm
		case "executable", "x":
			m.executable = bm
		case "external":
			m.external = bm
		case "modify":
			switch bm {
			case boolModifierClear:
				m.sourceFileType = sourceFileTypeModifierClearModify
			case boolModifierSet:
				m.sourceFileType = sourceFileTypeModifierSetModify
			}
		case "once", "o":
			switch bm {
			case boolModifierClear:
				m.condition = conditionModifierClearOnce
			case boolModifierSet:
				m.condition = conditionModifierSetOnce
			}
		case "onchange":
			switch bm {
			case boolModifierClear:
				m.condition = conditionModifierClearOnChange
			case boolModifierSet:
				m.condition = conditionModifierSetOnChange
			}
		case "private", "p":
			m.private = bm
		case "readonly", "r":
			m.readOnly = bm
		case "remove":
			switch bm {
			case boolModifierClear:
				m.remove = bm
				m.sourceFileType = sourceFileTypeModifierClearRemove
			case boolModifierSet:
				m.remove = bm
				m.sourceFileType = sourceFileTypeModifierSetRemove
			}
		case "script":
			switch bm {
			case boolModifierClear:
				m.sourceFileType = sourceFileTypeModifierClearScript
			case boolModifierSet:
				m.sourceFileType = sourceFileTypeModifierSetScript
			}
		case "symlink":
			switch bm {
			case boolModifierClear:
				m.sourceFileType = sourceFileTypeModifierClearSymlink
			case boolModifierSet:
				m.sourceFileType = sourceFileTypeModifierSetSymlink
			}
		case "template", "t":
			m.template = bm
		default:
			return nil, fmt.Errorf("%s: unknown attribute", attribute)
		}
	}
	return m, nil
}

// modifyDirAttr returns the modified value of dirAttr.
func (m *modifier) modifyDirAttr(dirAttr chezmoi.DirAttr) chezmoi.DirAttr {
	return chezmoi.DirAttr{
		TargetName: dirAttr.TargetName,
		Exact:      m.exact.modify(dirAttr.Exact),
		External:   m.external.modify(dirAttr.External),
		Private:    m.private.modify(dirAttr.Private),
		ReadOnly:   m.readOnly.modify(dirAttr.ReadOnly),
		Remove:     m.remove.modify(dirAttr.Remove),
	}
}

// modifyFileAttr returns the modified value of fileAttr.
func (m *modifier) modifyFileAttr(fileAttr chezmoi.FileAttr) chezmoi.FileAttr {
	switch m.sourceFileType.modify(fileAttr.Type) {
	case chezmoi.SourceFileTypeFile:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeFile,
			Empty:      m.empty.modify(fileAttr.Empty),
			Encrypted:  m.encrypted.modify(fileAttr.Encrypted),
			Executable: m.executable.modify(fileAttr.Executable),
			Private:    m.private.modify(fileAttr.Private),
			ReadOnly:   m.readOnly.modify(fileAttr.ReadOnly),
			Template:   m.template.modify(fileAttr.Template),
		}
	case chezmoi.SourceFileTypeModify:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeModify,
			Executable: m.executable.modify(fileAttr.Executable),
			Private:    m.private.modify(fileAttr.Private),
			ReadOnly:   m.readOnly.modify(fileAttr.ReadOnly),
			Template:   m.template.modify(fileAttr.Template),
		}
	case chezmoi.SourceFileTypeCreate:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeCreate,
			Empty:      m.encrypted.modify(fileAttr.Empty),
			Encrypted:  m.encrypted.modify(fileAttr.Encrypted),
			Executable: m.executable.modify(fileAttr.Executable),
			Private:    m.private.modify(fileAttr.Private),
			ReadOnly:   m.readOnly.modify(fileAttr.ReadOnly),
			Template:   m.template.modify(fileAttr.Template),
		}
	case chezmoi.SourceFileTypeScript:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeScript,
			Condition:  m.condition.modify(fileAttr.Condition),
			Order:      m.order.modify(fileAttr.Order),
		}
	case chezmoi.SourceFileTypeSymlink:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeSymlink,
			Template:   m.template.modify(fileAttr.Template),
		}
	case chezmoi.SourceFileTypeRemove:
		return chezmoi.FileAttr{
			TargetName: fileAttr.TargetName,
			Type:       chezmoi.SourceFileTypeRemove,
		}
	default:
		panic(fmt.Sprintf("%d: unknown source file type", fileAttr.Type))
	}
}
