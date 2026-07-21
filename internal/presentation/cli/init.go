package cli

import (
	"context"
	"fmt"

	"github.com/namdang-fdp/pivot/internal/application"
	"github.com/spf13/cobra"
)

type initProjectUseCase interface {
	InitProject(context.Context, application.InitProjectRequest) (application.InitProjectResult, error)
}

func newInitCommand(service initProjectUseCase) *cobra.Command {
	var id string
	var name string
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Create a minimal Pivot project manifest",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			result, err := service.InitProject(cmd.Context(), application.InitProjectRequest{Path: path, ID: id, Name: name})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created manifest: %s\nNext: pivot add %s\n", result.ManifestPath, shellQuote(result.ProjectPath))
			return err
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "override the generated project ID")
	cmd.Flags().StringVar(&name, "name", "", "override the project name")
	return cmd
}
