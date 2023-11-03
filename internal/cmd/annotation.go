package cmd

import "github.com/spf13/cobra"

// Annotations.
var (
	createSourceDirectoryIfNeeded = tagAnnotation("chezmoi_create_source_directory_if_needed")
	doesNotRequireValidConfig     = tagAnnotation("chezmoi_runs_with_invalid_config")
	dryRun                        = tagAnnotation("chezmoi_dry_run")
	modifiesConfigFile            = tagAnnotation("chezmoi_modifies_config_file")
	modifiesDestinationDirectory  = tagAnnotation("chezmoi_modifies_destination_directory")
	modifiesSourceDirectory       = tagAnnotation("chezmoi_modifies_source_directory")
	outputsDiff                   = tagAnnotation("chezmoi_outputs_diff")
	requiresConfigDirectory       = tagAnnotation("chezmoi_requires_config_directory")
	requiresSourceDirectory       = tagAnnotation("chezmoi_requires_source_directory")
	requiresWorkingTree           = tagAnnotation("chezmoi_requires_working_tree")
	runsCommands                  = tagAnnotation("chezmoi_runs_commands")
)

// Persistent state modes.
const (
	persistentStateModeKey = "chezmoi_persistent_state_mode"

	persistentStateModeEmpty         persistentStateMode = "empty"
	persistentStateModeReadOnly      persistentStateMode = "read-only"
	persistentStateModeReadMockWrite persistentStateMode = "read-mock-write"
	persistentStateModeReadWrite     persistentStateMode = "read-write"
)

type annotation interface {
	key() string
	value() string
}

type annotationsSet map[string]string

func getAnnotations(cmd *cobra.Command) annotationsSet {
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

func (a annotationsSet) persistentStateMode() persistentStateMode {
	return persistentStateMode(a[persistentStateModeKey])
}

type persistentStateMode string

func (m persistentStateMode) key() string {
	return persistentStateModeKey
}

func (m persistentStateMode) value() string {
	return string(m)
}

type tagAnnotation string

func (a tagAnnotation) key() string {
	return string(a)
}

func (a tagAnnotation) value() string {
	return "true"
}
