package chezmoi

import (
	"io/fs"
	"strings"

	"github.com/rs/zerolog"
)

// A SourceFileTargetType is a the type of a target represented by a file in the
// source state. A file in the source state can represent a file, script, or
// symlink in the target state.
type SourceFileTargetType int

// Source file types.
const (
	SourceFileTypeCreate SourceFileTargetType = iota
	SourceFileTypeFile
	SourceFileTypeModify
	SourceFileTypeRemove
	SourceFileTypeScript
	SourceFileTypeSymlink
)

var sourceFileTypeStrs = map[SourceFileTargetType]string{
	SourceFileTypeCreate:  "create",
	SourceFileTypeFile:    "file",
	SourceFileTypeModify:  "modify",
	SourceFileTypeRemove:  "remove",
	SourceFileTypeScript:  "script",
	SourceFileTypeSymlink: "symlink",
}

// A ScriptOrder defines when a script should be executed.
type ScriptOrder int

// Script orders.
const (
	ScriptOrderBefore ScriptOrder = -1
	ScriptOrderDuring ScriptOrder = 0
	ScriptOrderAfter  ScriptOrder = 1
)

// A ScriptCondition defines under what conditions a script should be executed.
type ScriptCondition string

// Script conditions.
const (
	ScriptConditionNone     ScriptCondition = ""
	ScriptConditionAlways   ScriptCondition = "always"
	ScriptConditionOnce     ScriptCondition = "once"
	ScriptConditionOnChange ScriptCondition = "onchange"
)

// DirAttr holds attributes parsed from a source directory name.
type DirAttr struct {
	TargetName string
	Exact      bool
	Private    bool
	ReadOnly   bool
	Remove     bool
}

// A FileAttr holds attributes parsed from a source file name.
type FileAttr struct {
	TargetName string
	Type       SourceFileTargetType
	Condition  ScriptCondition
	Empty      bool
	Encrypted  bool
	Executable bool
	Order      ScriptOrder
	Private    bool
	ReadOnly   bool
	Template   bool
}

// parseDirAttr parses a single directory name in the source state.
func parseDirAttr(sourceName string) DirAttr {
	var (
		name     = sourceName
		exact    = false
		private  = false
		readOnly = false
		remove   = false
	)
	if strings.HasPrefix(name, removePrefix) {
		name = mustTrimPrefix(name, removePrefix)
		remove = true
	}
	if strings.HasPrefix(name, exactPrefix) {
		name = mustTrimPrefix(name, exactPrefix)
		exact = true
	}
	if strings.HasPrefix(name, privatePrefix) {
		name = mustTrimPrefix(name, privatePrefix)
		private = true
	}
	if strings.HasPrefix(name, readOnlyPrefix) {
		name = mustTrimPrefix(name, readOnlyPrefix)
		readOnly = true
	}
	switch {
	case strings.HasPrefix(name, dotPrefix):
		name = "." + mustTrimPrefix(name, dotPrefix)
	case strings.HasPrefix(name, literalPrefix):
		name = name[len(literalPrefix):]
	}
	return DirAttr{
		TargetName: name,
		Exact:      exact,
		Private:    private,
		ReadOnly:   readOnly,
		Remove:     remove,
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.ObjectMarshaler.MarshalZerologObject.
func (da DirAttr) MarshalZerologObject(e *zerolog.Event) {
	e.Str("targetName", da.TargetName)
	e.Bool("exact", da.Exact)
	e.Bool("private", da.Private)
	e.Bool("readOnly", da.ReadOnly)
}

// SourceName returns da's source name.
func (da DirAttr) SourceName() string {
	sourceName := ""
	if da.Remove {
		sourceName += removePrefix
	}
	if da.Exact {
		sourceName += exactPrefix
	}
	if da.Private {
		sourceName += privatePrefix
	}
	if da.ReadOnly {
		sourceName += readOnlyPrefix
	}
	switch {
	case strings.HasPrefix(da.TargetName, "."):
		sourceName += dotPrefix + mustTrimPrefix(da.TargetName, ".")
	case dirPrefixRx.MatchString(da.TargetName):
		sourceName += literalPrefix + da.TargetName
	default:
		sourceName += da.TargetName
	}
	return sourceName
}

// perm returns da's file mode.
func (da DirAttr) perm() fs.FileMode {
	perm := fs.ModePerm
	if da.Private {
		perm &^= 0o77
	}
	if da.ReadOnly {
		perm &^= 0o222
	}
	return perm
}

// parseFileAttr parses a source file name in the source state.
func parseFileAttr(sourceName, encryptedSuffix string) FileAttr {
	var (
		sourceFileType = SourceFileTypeFile
		name           = sourceName
		condition      = ScriptConditionNone
		empty          = false
		encrypted      = false
		executable     = false
		order          = ScriptOrderDuring
		private        = false
		readOnly       = false
		template       = false
	)
	switch {
	case strings.HasPrefix(name, createPrefix):
		sourceFileType = SourceFileTypeCreate
		name = mustTrimPrefix(name, createPrefix)
		if strings.HasPrefix(name, encryptedPrefix) {
			name = mustTrimPrefix(name, encryptedPrefix)
			encrypted = true
		}
		if strings.HasPrefix(name, privatePrefix) {
			name = mustTrimPrefix(name, privatePrefix)
			private = true
		}
		if strings.HasPrefix(name, readOnlyPrefix) {
			name = mustTrimPrefix(name, readOnlyPrefix)
			readOnly = true
		}
		if strings.HasPrefix(name, executablePrefix) {
			name = mustTrimPrefix(name, executablePrefix)
			executable = true
		}
	case strings.HasPrefix(name, removePrefix):
		sourceFileType = SourceFileTypeRemove
		name = mustTrimPrefix(name, removePrefix)
	case strings.HasPrefix(name, runPrefix):
		sourceFileType = SourceFileTypeScript
		name = mustTrimPrefix(name, runPrefix)
		switch {
		case strings.HasPrefix(name, oncePrefix):
			name = mustTrimPrefix(name, oncePrefix)
			condition = ScriptConditionOnce
		case strings.HasPrefix(name, onChangePrefix):
			name = mustTrimPrefix(name, onChangePrefix)
			condition = ScriptConditionOnChange
		default:
			condition = ScriptConditionAlways
		}
		switch {
		case strings.HasPrefix(name, beforePrefix):
			name = mustTrimPrefix(name, beforePrefix)
			order = ScriptOrderBefore
		case strings.HasPrefix(name, afterPrefix):
			name = mustTrimPrefix(name, afterPrefix)
			order = ScriptOrderAfter
		}
	case strings.HasPrefix(name, symlinkPrefix):
		sourceFileType = SourceFileTypeSymlink
		name = mustTrimPrefix(name, symlinkPrefix)
	case strings.HasPrefix(name, modifyPrefix):
		sourceFileType = SourceFileTypeModify
		name = mustTrimPrefix(name, modifyPrefix)
		if strings.HasPrefix(name, encryptedPrefix) {
			name = mustTrimPrefix(name, encryptedPrefix)
			encrypted = true
		}
		if strings.HasPrefix(name, privatePrefix) {
			name = mustTrimPrefix(name, privatePrefix)
			private = true
		}
		if strings.HasPrefix(name, readOnlyPrefix) {
			name = mustTrimPrefix(name, readOnlyPrefix)
			readOnly = true
		}
		if strings.HasPrefix(name, executablePrefix) {
			name = mustTrimPrefix(name, executablePrefix)
			executable = true
		}
	default:
		if strings.HasPrefix(name, encryptedPrefix) {
			name = mustTrimPrefix(name, encryptedPrefix)
			encrypted = true
		}
		if strings.HasPrefix(name, privatePrefix) {
			name = mustTrimPrefix(name, privatePrefix)
			private = true
		}
		if strings.HasPrefix(name, readOnlyPrefix) {
			name = mustTrimPrefix(name, readOnlyPrefix)
			readOnly = true
		}
		if strings.HasPrefix(name, emptyPrefix) {
			name = mustTrimPrefix(name, emptyPrefix)
			empty = true
		}
		if strings.HasPrefix(name, executablePrefix) {
			name = mustTrimPrefix(name, executablePrefix)
			executable = true
		}
	}
	switch {
	case strings.HasPrefix(name, dotPrefix):
		name = "." + mustTrimPrefix(name, dotPrefix)
	case strings.HasPrefix(name, literalPrefix):
		name = name[len(literalPrefix):]
	}
	if encrypted {
		name = strings.TrimSuffix(name, encryptedSuffix)
	}
	switch {
	case strings.HasSuffix(name, literalSuffix):
		name = mustTrimSuffix(name, literalSuffix)
	case strings.HasSuffix(name, TemplateSuffix):
		name = mustTrimSuffix(name, TemplateSuffix)
		template = true
		if strings.HasSuffix(name, literalSuffix) {
			name = mustTrimSuffix(name, literalSuffix)
		}
	}
	return FileAttr{
		TargetName: name,
		Type:       sourceFileType,
		Condition:  condition,
		Empty:      empty,
		Encrypted:  encrypted,
		Executable: executable,
		Order:      order,
		Private:    private,
		ReadOnly:   readOnly,
		Template:   template,
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (fa FileAttr) MarshalZerologObject(e *zerolog.Event) {
	e.Str("TargetName", fa.TargetName)
	e.Str("Type", sourceFileTypeStrs[fa.Type])
	e.Str("Condition", string(fa.Condition))
	e.Bool("Empty", fa.Empty)
	e.Bool("Encrypted", fa.Encrypted)
	e.Bool("Executable", fa.Executable)
	e.Int("Order", int(fa.Order))
	e.Bool("Private", fa.Private)
	e.Bool("ReadOnly", fa.ReadOnly)
	e.Bool("Template", fa.Template)
}

// SourceName returns fa's source name.
func (fa FileAttr) SourceName(encryptedSuffix string) string {
	sourceName := ""
	switch fa.Type {
	case SourceFileTypeCreate:
		sourceName = createPrefix
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeFile:
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeModify:
		sourceName = modifyPrefix
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeRemove:
		sourceName = removePrefix
	case SourceFileTypeScript:
		sourceName = runPrefix
		switch fa.Condition {
		case ScriptConditionOnce:
			sourceName += oncePrefix
		case ScriptConditionOnChange:
			sourceName += onChangePrefix
		}
		switch fa.Order {
		case ScriptOrderBefore:
			sourceName += beforePrefix
		case ScriptOrderAfter:
			sourceName += afterPrefix
		}
	case SourceFileTypeSymlink:
		sourceName = symlinkPrefix
	}
	switch {
	case strings.HasPrefix(fa.TargetName, "."):
		sourceName += dotPrefix + mustTrimPrefix(fa.TargetName, ".")
	case filePrefixRx.MatchString(fa.TargetName):
		sourceName += literalPrefix + fa.TargetName
	default:
		sourceName += fa.TargetName
	}
	if fileSuffixRx.MatchString(fa.TargetName) {
		sourceName += literalSuffix
	}
	if fa.Template {
		sourceName += TemplateSuffix
	}
	if fa.Encrypted {
		sourceName += encryptedSuffix
	}
	return sourceName
}

// perm returns fa's permissions.
func (fa FileAttr) perm() fs.FileMode {
	perm := fs.FileMode(0o666)
	if fa.Executable {
		perm |= 0o111
	}
	if fa.Private {
		perm &^= 0o77
	}
	if fa.ReadOnly {
		perm &^= 0o222
	}
	return perm
}
