package chezmoi

import (
	"io/fs"
	"strings"
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
	SourceFileTypeScript
	SourceFileTypeSymlink
)

// DirAttr holds attributes parsed from a source directory name.
type DirAttr struct {
	TargetName string
	Exact      bool
	Private    bool
}

// A FileAttr holds attributes parsed from a source file name.
type FileAttr struct {
	TargetName string
	Type       SourceFileTargetType
	Empty      bool
	Encrypted  bool
	Executable bool
	Once       bool
	Order      int
	Private    bool
	Template   bool
}

// parseDirAttr parses a single directory name in the source state.
func parseDirAttr(sourceName string) DirAttr {
	var (
		name    = sourceName
		exact   = false
		private = false
	)
	if strings.HasPrefix(name, exactPrefix) {
		name = mustTrimPrefix(name, exactPrefix)
		exact = true
	}
	if strings.HasPrefix(name, privatePrefix) {
		name = mustTrimPrefix(name, privatePrefix)
		private = true
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
	}
}

// SourceName returns da's source name.
func (da DirAttr) SourceName() string {
	sourceName := ""
	if da.Exact {
		sourceName += exactPrefix
	}
	if da.Private {
		sourceName += privatePrefix
	}
	switch {
	case strings.HasPrefix(da.TargetName, "."):
		sourceName += dotPrefix + mustTrimPrefix(da.TargetName, ".")
	case dirPrefixRegexp.MatchString(da.TargetName):
		sourceName += literalPrefix + da.TargetName
	default:
		sourceName += da.TargetName
	}
	return sourceName
}

// perm returns da's file mode.
func (da DirAttr) perm() fs.FileMode {
	perm := fs.FileMode(0o777)
	if da.Private {
		perm &^= 0o77
	}
	return perm
}

// parseFileAttr parses a source file name in the source state.
func parseFileAttr(sourceName, encryptedSuffix string) FileAttr {
	var (
		sourceFileType = SourceFileTypeFile
		name           = sourceName
		empty          = false
		encrypted      = false
		executable     = false
		once           = false
		private        = false
		template       = false
		order          = 0
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
		if strings.HasPrefix(name, executablePrefix) {
			name = mustTrimPrefix(name, executablePrefix)
			executable = true
		}
	case strings.HasPrefix(name, runPrefix):
		sourceFileType = SourceFileTypeScript
		name = mustTrimPrefix(name, runPrefix)
		if strings.HasPrefix(name, oncePrefix) {
			name = mustTrimPrefix(name, oncePrefix)
			once = true
		}
		switch {
		case strings.HasPrefix(name, beforePrefix):
			name = mustTrimPrefix(name, beforePrefix)
			order = -1
		case strings.HasPrefix(name, afterPrefix):
			name = mustTrimPrefix(name, afterPrefix)
			order = 1
		}
	case strings.HasPrefix(name, symlinkPrefix):
		sourceFileType = SourceFileTypeSymlink
		name = mustTrimPrefix(name, symlinkPrefix)
	case strings.HasPrefix(name, modifyPrefix):
		sourceFileType = SourceFileTypeModify
		name = mustTrimPrefix(name, modifyPrefix)
		if strings.HasPrefix(name, privatePrefix) {
			name = mustTrimPrefix(name, privatePrefix)
			private = true
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
	if strings.HasSuffix(name, TemplateSuffix) {
		name = mustTrimSuffix(name, TemplateSuffix)
		template = true
	}
	return FileAttr{
		TargetName: name,
		Type:       sourceFileType,
		Empty:      empty,
		Encrypted:  encrypted,
		Executable: executable,
		Once:       once,
		Private:    private,
		Template:   template,
		Order:      order,
	}
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
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeModify:
		sourceName = modifyPrefix
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeScript:
		sourceName = runPrefix
		if fa.Once {
			sourceName += oncePrefix
		}
		switch fa.Order {
		case -1:
			sourceName += beforePrefix
		case 1:
			sourceName += afterPrefix
		}
	case SourceFileTypeSymlink:
		sourceName = symlinkPrefix
	}
	switch {
	case strings.HasPrefix(fa.TargetName, "."):
		sourceName += dotPrefix + mustTrimPrefix(fa.TargetName, ".")
	case filePrefixRegexp.MatchString(fa.TargetName):
		sourceName += literalPrefix + fa.TargetName
	default:
		sourceName += fa.TargetName
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
	return perm
}
