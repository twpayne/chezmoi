package chezmoi

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// A LoggingActuator wraps an Actuator and logs all of the actions it executes
// and any errors.
type LoggingActuator struct {
	a Actuator
}

// NewLoggingActuator returns a new LoggingActuator.
func NewLoggingActuator(a Actuator) *LoggingActuator {
	return &LoggingActuator{
		a: a,
	}
}

// Chmod implements Actuator.Chmod.
func (a *LoggingActuator) Chmod(name string, mode os.FileMode) error {
	action := fmt.Sprintf("chmod %o %s", mode, name)
	err := a.a.Chmod(name, mode)
	if err == nil {
		log.Print(action)
	} else {
		log.Printf("%s: %v", action, err)
	}
	return err
}

// Mkdir implements Actuator.Mkdir.
func (a *LoggingActuator) Mkdir(name string, mode os.FileMode) error {
	action := fmt.Sprintf("mkdir -m %o %s", mode, name)
	err := a.a.Mkdir(name, mode)
	if err == nil {
		log.Print(action)
	} else {
		log.Printf("%s: %v", action, err)
	}
	return err
}

// RemoveAll implements Actuator.RemoveAll.
func (a *LoggingActuator) RemoveAll(name string) error {
	action := fmt.Sprintf("rm -rf %s", name)
	err := a.a.RemoveAll(name)
	if err == nil {
		log.Print(action)
	} else {
		log.Printf("%s: %v", action, err)
	}
	return err
}

// WriteFile implements Actuator.WriteFile.
func (a *LoggingActuator) WriteFile(name string, contents []byte, mode os.FileMode, currentContents []byte) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", mode, name)
	err := a.a.WriteFile(name, contents, mode, currentContents)
	if err == nil {
		log.Print(action)
		dmp := diffmatchpatch.New()
		textA, textB, lineArray := dmp.DiffLinesToChars(string(currentContents), string(contents))
		charDiffs := dmp.DiffMain(textA, textB, false)
		diffs := dmp.DiffCharsToLines(charDiffs, lineArray)
		// FIXME print standard diff
		for _, diff := range diffs {
			if diff.Type == diffmatchpatch.DiffEqual {
				continue
			}
			lines := strings.Split(diff.Text, "\n")
			for i := 0; i < len(lines)-1; i++ {
				switch diff.Type {
				case diffmatchpatch.DiffDelete:
					log.Printf("-%s", lines[i])
				case diffmatchpatch.DiffInsert:
					log.Printf("+%s", lines[i])
				}
			}
		}
	} else {
		log.Printf("%s: %v", action, err)
	}
	return err
}
