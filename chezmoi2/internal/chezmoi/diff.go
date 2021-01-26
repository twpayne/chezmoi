package chezmoi

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var gitDiffOperation = map[diffmatchpatch.Operation]diff.Operation{
	diffmatchpatch.DiffDelete: diff.Delete,
	diffmatchpatch.DiffEqual:  diff.Equal,
	diffmatchpatch.DiffInsert: diff.Add,
}

type gitDiffChunk struct {
	content   string
	operation diff.Operation
}

func (c *gitDiffChunk) Content() string      { return c.content }
func (c *gitDiffChunk) Type() diff.Operation { return c.operation }

type gitDiffFile struct {
	hash     plumbing.Hash
	fileMode filemode.FileMode
	relPath  RelPath
}

func (f *gitDiffFile) Hash() plumbing.Hash     { return f.hash }
func (f *gitDiffFile) Mode() filemode.FileMode { return f.fileMode }
func (f *gitDiffFile) Path() string            { return string(f.relPath) }

type gitDiffFilePatch struct {
	isBinary bool
	from, to diff.File
	chunks   []diff.Chunk
}

func (fp *gitDiffFilePatch) IsBinary() bool                { return fp.isBinary }
func (fp *gitDiffFilePatch) Files() (diff.File, diff.File) { return fp.from, fp.to }
func (fp *gitDiffFilePatch) Chunks() []diff.Chunk          { return fp.chunks }

type gitDiffPatch struct {
	filePatches []diff.FilePatch
	message     string
}

func (p *gitDiffPatch) FilePatches() []diff.FilePatch { return p.filePatches }
func (p *gitDiffPatch) Message() string               { return p.message }

// DiffPatch returns a github.com/go-git/go-git/plumbing/format/diff.Patch for
// path from the given data and mode to the given data and mode.
func DiffPatch(path RelPath, fromData []byte, fromMode os.FileMode, toData []byte, toMode os.FileMode) (diff.Patch, error) {
	isBinary := isBinary(fromData) || isBinary(toData)

	var from diff.File
	if fromData != nil || fromMode != 0 {
		fromFileMode, err := filemode.NewFromOSFileMode(fromMode)
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
		toFileMode, err := filemode.NewFromOSFileMode(toMode)
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
				isBinary: isBinary,
				from:     from,
				to:       to,
				chunks:   chunks,
			},
		},
	}, nil
}

func diffChunks(from, to string) []diff.Chunk {
	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = time.Second
	fromRunes, toRunes, runesToLines := dmp.DiffLinesToRunes(from, to)
	diffs := dmp.DiffCharsToLines(dmp.DiffMainRunes(fromRunes, toRunes, false), runesToLines)
	chunks := make([]diff.Chunk, 0, len(diffs))
	for _, d := range diffs {
		chunk := &gitDiffChunk{
			content:   d.Text,
			operation: gitDiffOperation[d.Type],
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func isBinary(data []byte) bool {
	return len(data) != 0 && !strings.HasPrefix(http.DetectContentType(data), "text/")
}
