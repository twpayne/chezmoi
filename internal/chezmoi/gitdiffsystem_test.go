package chezmoi

import (
	"github.com/go-git/go-git/v5/plumbing/format/diff"
)

var (
	_ System         = &GitDiffSystem{}
	_ diff.Chunk     = &GitDiffChunk{}
	_ diff.File      = &GitDiffFile{}
	_ diff.FilePatch = &GitDiffFilePatch{}
	_ diff.Patch     = &GitDiffPatch{}
)
