package chezmoi

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// A GitDiffMutator wraps a Mutator and logs all of the actions it would execute
// as a git diff.
type GitDiffMutator struct {
	m              Mutator
	prefix         string
	unifiedEncoder *diff.UnifiedEncoder
}

// NewGitDiffMutator returns a new GitDiffMutator.
func NewGitDiffMutator(unifiedEncoder *diff.UnifiedEncoder, m Mutator, prefix string) *GitDiffMutator {
	return &GitDiffMutator{
		m:              m,
		prefix:         prefix,
		unifiedEncoder: unifiedEncoder,
	}
}

// Chmod implements Mutator.Chmod.
func (m *GitDiffMutator) Chmod(name string, mode os.FileMode) error {
	fromFileMode, info, err := m.getFileMode(name)
	if err != nil {
		return err
	}
	// Assume that we're only changing permissions.
	toFileMode, err := filemode.NewFromOSFileMode(info.Mode()&^os.ModePerm | mode)
	if err != nil {
		return err
	}
	path := m.trimPrefix(name)
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				from: &gitDiffFile{
					fileMode: fromFileMode,
					path:     path,
					hash:     plumbing.ZeroHash,
				},
				to: &gitDiffFile{
					fileMode: toFileMode,
					path:     path,
					hash:     plumbing.ZeroHash,
				},
			},
		},
	})
}

// IdempotentCmdOutput implements Mutator.IdempotentCmdOutput.
func (m *GitDiffMutator) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return m.m.IdempotentCmdOutput(cmd)
}

// Mkdir implements Mutator.Mkdir.
func (m *GitDiffMutator) Mkdir(name string, perm os.FileMode) error {
	toFileMode, err := filemode.NewFromOSFileMode(os.ModeDir | perm)
	if err != nil {
		return err
	}
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				to: &gitDiffFile{
					fileMode: toFileMode,
					path:     m.trimPrefix(name),
					hash:     plumbing.ZeroHash,
				},
			},
		},
	})
}

// RemoveAll implements Mutator.RemoveAll.
func (m *GitDiffMutator) RemoveAll(name string) error {
	fromFileMode, _, err := m.getFileMode(name)
	if err != nil {
		return err
	}
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				from: &gitDiffFile{
					fileMode: fromFileMode,
					path:     m.trimPrefix(name),
					hash:     plumbing.ZeroHash,
				},
			},
		},
	})
}

// RunCmd implements Mutator.RunCmd.
func (m *GitDiffMutator) RunCmd(cmd *exec.Cmd) error {
	// FIXME write scripts to diff
	return nil
}

// Stat implements Mutator.Stat.
func (m *GitDiffMutator) Stat(name string) (os.FileInfo, error) {
	return m.m.Stat(name)
}

// Rename implements Mutator.Rename.
func (m *GitDiffMutator) Rename(oldpath, newpath string) error {
	fileMode, _, err := m.getFileMode(oldpath)
	if err != nil {
		return err
	}
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				from: &gitDiffFile{
					fileMode: fileMode,
					path:     m.trimPrefix(oldpath),
					hash:     plumbing.ZeroHash,
				},
				to: &gitDiffFile{
					fileMode: fileMode,
					path:     m.trimPrefix(newpath),
					hash:     plumbing.ZeroHash,
				},
			},
		},
	})
}

// WriteFile implements Mutator.WriteFile.
func (m *GitDiffMutator) WriteFile(filename string, data []byte, perm os.FileMode, currData []byte) error {
	fileMode, _, err := m.getFileMode(filename)
	if err != nil {
		return err
	}
	path := m.trimPrefix(filename)
	isBinary := isBinary(currData) || isBinary(data)
	var chunks []diff.Chunk
	if !isBinary {
		chunks = diffChunks(string(currData), string(data))
	}
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				isBinary: isBinary,
				from: &gitDiffFile{
					fileMode: fileMode,
					path:     path,
					hash:     plumbing.ComputeHash(plumbing.BlobObject, currData),
				},
				to: &gitDiffFile{
					fileMode: fileMode,
					path:     path,
					hash:     plumbing.ComputeHash(plumbing.BlobObject, data),
				},
				chunks: chunks,
			},
		},
	})
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *GitDiffMutator) WriteSymlink(oldname, newname string) error {
	return m.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				to: &gitDiffFile{
					fileMode: filemode.Symlink,
					path:     m.trimPrefix(newname),
					hash:     plumbing.ComputeHash(plumbing.BlobObject, []byte(oldname)),
				},
				chunks: []diff.Chunk{
					&gitDiffChunk{
						content:   oldname,
						operation: diff.Add,
					},
				},
			},
		},
	})
}

func (m *GitDiffMutator) getFileMode(name string) (filemode.FileMode, os.FileInfo, error) {
	info, err := m.m.Stat(name)
	if os.IsNotExist(err) {
		return filemode.Empty, nil, nil
	} else if err != nil {
		return filemode.Empty, nil, err
	}
	fileMode, err := filemode.NewFromOSFileMode(info.Mode())
	return fileMode, info, err
}

func (m *GitDiffMutator) trimPrefix(path string) string {
	return strings.TrimPrefix(path, m.prefix)
}

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
	path     string
}

func (f *gitDiffFile) Hash() plumbing.Hash     { return f.hash }
func (f *gitDiffFile) Mode() filemode.FileMode { return f.fileMode }
func (f *gitDiffFile) Path() string            { return f.path }

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
