// Package cli implements Pivot's command-line presentation layer.
package cli

import (
	"github.com/namdang-fdp/pivot/internal/adapters"
	"github.com/namdang-fdp/pivot/internal/application"
	"github.com/spf13/cobra"
)

// NewRootCommand constructs an isolated Pivot command tree.
func NewRootCommand() *cobra.Command {
	manifests := adapters.NewYAMLManifestRepository()
	registry := adapters.NewYAMLProjectRegistry()
	files := adapters.NewHostFilesystem()

	return newRootCommand(
		application.NewInitProjectService(manifests, files),
		application.NewAddProjectService(manifests, registry, files),
		application.NewListProjectsService(registry, files),
		application.NewDoctorProjectService(
			manifests,
			registry,
			files,
			adapters.NewHostCommandInspector(),
			adapters.NewLinuxPortInspector(),
		),
	)
}

func newRootCommand(
	initProjects initProjectUseCase,
	addProjects addProjectUseCase,
	listProjects listProjectsUseCase,
	doctorProjects doctorProjectUseCase,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pivot",
		Short:         "Safely orchestrate local project contexts",
		Long:          pivotBanner + "\n\n" + pivotTagline + "\n\nPivot safely orchestrates complete local development contexts.",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newVersionCommand(),
		newInitCommand(initProjects),
		newAddCommand(addProjects),
		newListCommand(listProjects),
		newDoctorCommand(doctorProjects),
	)
	return cmd
}
