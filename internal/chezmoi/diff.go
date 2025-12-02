package chezmoi

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	znkrdiff "znkr.io/diff"
	znkrtextdiff "znkr.io/diff/textdiff"
)

var gitDiffOperation = [...]diff.Operation{
	znkrdiff.Delete: diff.Delete,
	znkrdiff.Match:  diff.Equal,
	znkrdiff.Insert: diff.Add,
}

// A gitDiffChunk implements the
// github.com/go-git/go-git/v5/plumbing/format/diff.Chunk interface.
type gitDiffChunk struct {
	content   string
	operation diff.Operation
}

func (c *gitDiffChunk) Content() string      { return c.content }
func (c *gitDiffChunk) Type() diff.Operation { return c.operation }

// A gitDiffFile implements the
// github.com/go-git/go-git/v5/plumbing/format/diff.File interface.
type gitDiffFile struct {
	hash     plumbing.Hash
	fileMode filemode.FileMode
	relPath  RelPath
}

func (f *gitDiffFile) Hash() plumbing.Hash     { return f.hash }
func (f *gitDiffFile) Mode() filemode.FileMode { return f.fileMode }
func (f *gitDiffFile) Path() string            { return f.relPath.String() }

// A gitDiffFilePatch implements the
// github.com/go-git/go-git/v5/plumbing/format/diff.FilePatch interface.
type gitDiffFilePatch struct {
	binary   bool
	from, to diff.File
	chunks   []diff.Chunk
}

func (fp *gitDiffFilePatch) IsBinary() bool              { return fp.binary }
func (fp *gitDiffFilePatch) Files() (from, to diff.File) { return fp.from, fp.to }
func (fp *gitDiffFilePatch) Chunks() []diff.Chunk        { return fp.chunks }

// A gitDiffPatch implements the
// github.com/go-git/go-git/v5/plumbing/format/diff.Patch interface.
type gitDiffPatch struct {
	filePatches []diff.FilePatch
	message     string
}

func (p *gitDiffPatch) FilePatches() []diff.FilePatch { return p.filePatches }
func (p *gitDiffPatch) Message() string               { return p.message }

// DiffPatch returns a github.com/go-git/go-git/plumbing/format/diff.Patch for
// path from the given data and mode to the given data and mode.
func DiffPatch(path RelPath, fromData []byte, fromMode fs.FileMode, toData []byte, toMode fs.FileMode) (diff.Patch, error) {
	isBinary := isBinary(fromData) || isBinary(toData)

	var from diff.File
	if fromData != nil || fromMode != 0 {
		fromFileMode, err := diffFileMode(fromMode)
		if err != nil {
			return nil, err
		}
		from = &gitDiffFile{
			fileMode: fromFileMode,
			relPath:  path,
			hash:     plumbing.ComputeHash(plumbing.BlobObject, fromData),
		}
	}

	var to diff.File
	if toData != nil || toMode != 0 {
		toFileMode, err := diffFileMode(toMode)
		if err != nil {
			return nil, err
		}
		to = &gitDiffFile{
			fileMode: toFileMode,
			relPath:  path,
			hash:     plumbing.ComputeHash(plumbing.BlobObject, toData),
		}
	}

	var chunks []diff.Chunk
	if !isBinary {
		chunks = diffChunks(string(fromData), string(toData))
	}

	return &gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				binary: isBinary,
				from:   from,
				to:     to,
				chunks: chunks,
			},
		},
	}, nil
}

// diffChunks returns the
// github.com/go-git/go-git/v5/plumbing/format/diff.Chunks required to transform
// from into to.
//
// Nothing documents this, but go-git seems to depend on never encountering
// consecutive edits of the same operation. This forces us to join the
// individual edits into chunks.
func diffChunks(from, to string) []diff.Chunk {
	edits := znkrtextdiff.Edits(from, to, znkrdiff.Minimal())
	if len(edits) == 0 {
		return nil
	}
	var chunks []diff.Chunk
	lastOp := edits[0].Op
	var sb strings.Builder
	for _, edit := range edits {
		if edit.Op != lastOp {
			if content := sb.String(); content != "" {
				chunks = append(chunks, &gitDiffChunk{
					content:   content,
					operation: gitDiffOperation[lastOp],
				})
			}
			lastOp = edit.Op
			sb.Reset()
		}
		sb.WriteString(edit.Line)
	}
	if content := sb.String(); content != "" {
		chunks = append(chunks, &gitDiffChunk{
			content:   content,
			operation: gitDiffOperation[lastOp],
		})
	}
	return chunks
}

// diffFileMode converts an io/fs.FileMode into a
// github.com/go-git/go-git/v5/plumbing/format/diff.FileMode.
func diffFileMode(mode fs.FileMode) (filemode.FileMode, error) {
	fileMode, err := filemode.NewFromOSFileMode(mode)
	if err != nil {
		return 0, err
	}
	return (fileMode &^ filemode.FileMode(fs.ModePerm)) | filemode.FileMode(mode.Perm()), nil
}

// isBinary returns true if data contains binary (non-human-readable) data.
func isBinary(data []byte) bool {
	return len(data) != 0 && !strings.HasPrefix(http.DetectContentType(data), "text/")
}
