package chezmoi

import (
	"errors"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/v2/internal/chezmoitest"
)

var _ System = &RealSystem{}

func TestRealSystemGlob(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user": map[string]any{
			"bar":            "",
			"baz":            "",
			"foo":            "",
			"dir/bar":        "",
			"dir/foo":        "",
			"dir/subdir/foo": "",
		},
	}, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
		for _, tc := range []struct {
			pattern         string
			expectedMatches []string
		}{
			{
				pattern: "/home/user/foo",
				expectedMatches: []string{
					"/home/user/foo",
				},
			},
			{
				pattern: "/home/user/**/foo",
				expectedMatches: []string{
					"/home/user/dir/foo",
					"/home/user/dir/subdir/foo",
					"/home/user/foo",
				},
			},
			{
				pattern: "/home/user/**/ba*",
				expectedMatches: []string{
					"/home/user/bar",
					"/home/user/baz",
					"/home/user/dir/bar",
				},
			},
		} {
			t.Run(tc.pattern, func(t *testing.T) {
				actualMatches, err := system.Glob(tc.pattern)
				assert.NoError(t, err)
				slices.Sort(actualMatches)
				assert.Equal(t, tc.expectedMatches, pathsToSlashes(actualMatches))
			})
		}
	})
}

func pathsToSlashes(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, filepath.ToSlash(path))
	}
	return result
}

func TestPrepareScriptCmd(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/work": map[string]any{},
	}, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
		scriptTempDir, err := system.RawPath(NewAbsPath("/scripts"))
		assert.NoError(t, err)
		workingDir := NewAbsPath("/work")
		workingDirRawAbsPath, err := system.RawPath(workingDir)
		assert.NoError(t, err)

		for _, tc := range []struct {
			name           string
			setWorkingDir  bool
			expectedCmdDir string
		}{
			{
				name: "without_working_dir",
			},
			{
				name:           "with_working_dir",
				setWorkingDir:  true,
				expectedCmdDir: workingDirRawAbsPath.String(),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var createScriptTempDirOnce sync.Once
				interpreter := &Interpreter{
					Command: "interpreter",
					Args:    []string{"arg"},
				}
				args := runScriptArgs{
					scriptName:    "script.sh",
					workingDir:    workingDir,
					setWorkingDir: tc.setWorkingDir,
					data:          []byte("# script contents\n"),
					interpreter:   interpreter,
					sourceRelPath: NewSourceRelPath("run_script.sh"),
				}
				state := runScriptState{
					createScriptTempDirOnce: &createScriptTempDirOnce,
					scriptTempDir:           scriptTempDir,
					system:                  system,
				}

				preparedScript, err := prepareScriptCmd(args, state)
				assert.NoError(t, err)

				// Assert the returned command matches expected values
				scriptName := preparedScript.cmd.Args[len(preparedScript.cmd.Args)-1]
				assert.Equal(t, interpreter.Command, preparedScript.cmd.Path)
				assert.Equal(t, []string{"interpreter", "arg", scriptName}, preparedScript.cmd.Args)
				assert.Equal(t, tc.expectedCmdDir, preparedScript.cmd.Dir)
				assert.SliceContains(t, preparedScript.cmd.Env, "CHEZMOI_SOURCE_FILE=run_script.sh")

				// Assert the directory was created with the appropriate permissions
				scriptTempDirFileInfo, err := system.Stat(NewAbsPath("/scripts"))
				assert.NoError(t, err)
				assert.True(t, scriptTempDirFileInfo.IsDir())
				if runtime.GOOS != "windows" {
					assert.Equal(t, 0o700, scriptTempDirFileInfo.Mode().Perm())
				}

				// Assert the temporary script file has appropriate path and name
				assert.HasPrefix(t, filepath.ToSlash(scriptName), scriptTempDir.String()+"/")
				assert.HasSuffix(t, scriptName, ".script.sh")

				// Assert the temporary script contents and permissions
				scriptAbsPath := NewAbsPath("/scripts").JoinString(filepath.Base(scriptName))
				data, err := system.ReadFile(scriptAbsPath)
				assert.NoError(t, err)
				assert.Equal(t, []byte("# script contents\n"), data)
				if runtime.GOOS != "windows" {
					fileInfo, err := system.Stat(scriptAbsPath)
					assert.NoError(t, err)
					assert.Equal(t, 0o700, fileInfo.Mode().Perm())
				}

				// Assert the cleanup function properly deleted the temporary script
				assert.NoError(t, preparedScript.cleanup())
				_, err = system.Stat(scriptAbsPath)
				assert.True(t, errors.Is(err, fs.ErrNotExist))
			})
		}
	})
}

func TestPrepareScriptCmdCleansUpOnError(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/work": map[string]any{},
	}, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
		scriptTempDir, err := system.RawPath(NewAbsPath("/scripts"))
		assert.NoError(t, err)
		workingDir := NewAbsPath("/work")
		workingDirFileInfo, err := system.Stat(workingDir)
		assert.NoError(t, err)

		var createScriptTempDirOnce sync.Once

		// Manually trigger an error to validate the premature exit scenario
		rawPathErr := errors.New("raw path error")
		rawPathErrorSystem := &rawPathErrorSystem{
			expectedRawPath: workingDir,
			fileInfo:        workingDirFileInfo,
			err:             rawPathErr,
		}

		_, err = prepareScriptCmd(
			runScriptArgs{
				scriptName:    "script.sh",
				workingDir:    workingDir,
				setWorkingDir: true,
				data:          []byte("# script contents\n"),
				interpreter:   &Interpreter{},
				sourceRelPath: NewSourceRelPath("run_script.sh"),
			},
			runScriptState{
				createScriptTempDirOnce: &createScriptTempDirOnce,
				scriptTempDir:           scriptTempDir,
				system:                  rawPathErrorSystem,
			},
		)
		assert.True(t, errors.Is(err, rawPathErr))

		dirEntries, err := system.ReadDir(NewAbsPath("/scripts"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(dirEntries))
	})
}

type rawPathErrorSystem struct {
	emptySystemMixin
	noUpdateSystemMixin

	expectedRawPath AbsPath
	fileInfo        fs.FileInfo
	err             error
}

func (s *rawPathErrorSystem) RawPath(absPath AbsPath) (AbsPath, error) {
	if absPath == s.expectedRawPath {
		return EmptyAbsPath, s.err
	}
	return absPath, nil
}

func (s *rawPathErrorSystem) Stat(AbsPath) (fs.FileInfo, error) {
	return s.fileInfo, nil
}
