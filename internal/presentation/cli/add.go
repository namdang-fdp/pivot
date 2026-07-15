package cli

import (
	"context"
	"fmt"

	"github.com/namdang-fdp/pivot/internal/application"
	"github.com/spf13/cobra"
)

type addProjectUseCase interface {
	AddProject(context.Context, string) (application.AddProjectResult, error)
}

func newAddCommand(service addProjectUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "add [path]",
		Short: "Validate and register a Pivot project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			result, err := service.AddProject(cmd.Context(), path)
			if err != nil {
				return err
			}
			if result.AlreadyRegistered {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Project %s is already registered at %s\n", result.Project.ID, result.Project.Path)
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Registered project %s (%s) at %s\n", result.Project.ID, result.Project.Name, result.Project.Path)
			return err
		},
	}
}
