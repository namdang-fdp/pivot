package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/namdang-fdp/pivot/internal/application"
	"github.com/spf13/cobra"
)

type listProjectsUseCase interface {
	ListProjects(context.Context) ([]application.ListedProject, error)
}

type listJSON struct {
	Projects []listedProjectJSON `json:"projects"`
}

type listedProjectJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Manifest  string `json:"manifest"`
	Available bool   `json:"available"`
}

func newListCommand(service listProjectsUseCase) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered Pivot projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			projects, err := service.ListProjects(cmd.Context())
			if err != nil {
				return err
			}
			if asJSON {
				output := listJSON{Projects: make([]listedProjectJSON, 0, len(projects))}
				for _, project := range projects {
					output.Projects = append(output.Projects, listedProjectJSON{
						ID: string(project.ID), Name: project.Name, Path: project.Path,
						Manifest: project.Manifest, Available: project.Available,
					})
				}
				return writeJSON(cmd, output)
			}
			if len(projects) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No projects registered. Use `pivot add <path>` to register one.")
				return err
			}
			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			if _, err := fmt.Fprintln(writer, "PROJECT\tNAME\tPATH\tAVAILABLE"); err != nil {
				return err
			}
			for _, project := range projects {
				if _, err := fmt.Fprintf(writer, "%s\t%s\t%s\t%t\n", project.ID, project.Name, project.Path, project.Available); err != nil {
					return err
				}
			}
			return writer.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print registered projects as JSON")
	return cmd
}

func writeJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
