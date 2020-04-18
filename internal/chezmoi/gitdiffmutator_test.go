package chezmoi

import "github.com/go-git/go-git/v5/plumbing/format/diff"

var (
	_ Mutator        = &GitDiffMutator{}
	_ diff.Chunk     = &gitDiffChunk{}
	_ diff.File      = &gitDiffFile{}
	_ diff.FilePatch = &gitDiffFilePatch{}
	_ diff.Patch     = &gitDiffPatch{}
)
