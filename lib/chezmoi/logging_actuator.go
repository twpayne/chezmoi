package chezmoi

import (
	"fmt"
	"log"
	"os"
)

type LoggingActuator struct {
	a Actuator
}

func NewLoggingActuator(a Actuator) *LoggingActuator {
	return &LoggingActuator{
		a: a,
	}
}

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

func (a *LoggingActuator) WriteFile(name string, contents []byte, mode os.FileMode) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", mode, name)
	err := a.a.WriteFile(name, contents, mode)
	if err == nil {
		log.Print(action)
	} else {
		log.Printf("%s: %v", action, err)
	}
	return err
}
