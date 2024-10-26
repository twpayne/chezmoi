package cmd

import (
	"slices"

	"github.com/spf13/cobra"
)

// Annotations.
var (
	createSourceDirectoryIfNeeded = tagAnnotation("chezmoi_create_source_directory_if_needed")
	doesNotRequireValidConfig     = tagAnnotation("chezmoi_runs_with_invalid_config")
	dryRun                        = tagAnnotation("chezmoi_dry_run")
	modifiesConfigFile            = tagAnnotation("chezmoi_modifies_config_file")
	modifiesDestinationDirectory  = tagAnnotation("chezmoi_modifies_destination_directory")
	modifiesSourceDirectory       = tagAnnotation("chezmoi_modifies_source_directory")
	outputsDiff                   = tagAnnotation("chezmoi_outputs_diff")
	persistentStateModeKey        = tagAnnotation("chezmoi_persistent_state_mode")
	requiresConfigDirectory       = tagAnnotation("chezmoi_requires_config_directory")
	requiresSourceDirectory       = tagAnnotation("chezmoi_requires_source_directory")
	requiresWorkingTree           = tagAnnotation("chezmoi_requires_working_tree")
	runsCommands                  = tagAnnotation("chezmoi_runs_commands")
)

// Persistent state modes.
const (
	persistentStateModeEmpty         persistentStateModeValue = "empty"
	persistentStateModeNone          persistentStateModeValue = "none"
	persistentStateModeReadOnly      persistentStateModeValue = "read-only"
	persistentStateModeReadMockWrite persistentStateModeValue = "read-mock-write"
	persistentStateModeReadWrite     persistentStateModeValue = "read-write"
)

type annotation interface {
	key() string
	value() string
}

type annotationsSet map[string]string

func getAnnotations(cmd *cobra.Command) annotationsSet {
	thirdPartyCommandNames := []string{
		"__complete",
	}
	if !slices.Contains(thirdPartyCommandNames, cmd.Name()) {
		if cmd.Annotations == nil {
			panic(cmd.Name() + ": no annotations")
		}
		if cmd.Annotations[string(persistentStateModeKey)] == "" {
			panic(cmd.Name() + ": persistent state mode not set")
		}
	}
	return annotationsSet(cmd.Annotations)
}

func newAnnotations(annotations ...annotation) annotationsSet {
	result := make(map[string]string, len(annotations))
	for _, annotation := range annotations {
		result[annotation.key()] = annotation.value()
	}
	return result
}

func (a annotationsSet) hasTag(tag annotation) bool {
	return a[tag.key()] == tag.value()
}

func (a annotationsSet) persistentStateMode() persistentStateModeValue {
	return persistentStateModeValue(a[string(persistentStateModeKey)])
}

type persistentStateModeValue string

func (m persistentStateModeValue) key() string {
	return string(persistentStateModeKey)
}

func (m persistentStateModeValue) value() string {
	return string(m)
}

type tagAnnotation string

func (a tagAnnotation) key() string {
	return string(a)
}

func (a tagAnnotation) value() string {
	return "true"
}
