package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

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
	order          orderModifier
	private        boolModifier
	readOnly       boolModifier
	template       boolModifier
}

func (c *Config) newChattrCmd() *cobra.Command {
	attrs := []string{
		"after", "a",
		"before", "b",
		"create",
		"empty", "e",
		"encrypted",
		"exact",
		"executable", "x",
		"modify",
		"once", "o",
		"onchange",
		"private", "p",
		"readonly", "r",
		"script",
		"symlink",
		"template", "t",
	}
	validArgs := make([]string, 0, 4*len(attrs))
	for _, attribute := range attrs {
		validArgs = append(validArgs, attribute, "-"+attribute, "+"+attribute, "no"+attribute)
	}

	chattrCmd := &cobra.Command{
		Use:       "chattr attributes target...",
		Short:     "Change the attributes of a target in the source state",
		Long:      mustLongHelp("chattr"),
		Example:   example("chattr"),
		Args:      cobra.MinimumNArgs(2),
		ValidArgs: validArgs,
		RunE:      c.makeRunEWithSourceState(c.runChattrCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
		},
	}

	return chattrCmd
}

func (c *Config) runChattrCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// LATER should the core functionality of chattr move to chezmoi.SourceState?

	m, err := parseModifier(args[0])
	if err != nil {
		return err
	}

	targetRelPaths, err := c.targetRelPaths(sourceState, args[1:], targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	// Sort targets in reverse so we update children before their parent
	// directories.
	sort.Sort(targetRelPaths)

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
			// FIXME encrypted attribute changes
			// FIXME when changing encrypted attribute add new file before removing old one
			relPath := m.modifyFileAttr(sourceStateEntry.Attr).SourceName(encryptedSuffix)
			if newBaseNameRelPath := chezmoi.NewRelPath(relPath); newBaseNameRelPath != fileRelPath {
				oldSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, fileRelPath)
				newSourceAbsPath := c.SourceDirAbsPath.Join(parentRelPath, newBaseNameRelPath)
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

// parseModifier parses the attrMmodifier from s.
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
		Private:    m.private.modify(dirAttr.Private),
		ReadOnly:   m.readOnly.modify(dirAttr.ReadOnly),
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
	default:
		panic(fmt.Sprintf("%d: unknown source file type", fileAttr.Type))
	}
}
