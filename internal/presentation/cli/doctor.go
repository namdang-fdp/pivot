package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/namdang-fdp/pivot/internal/application"
	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/spf13/cobra"
)

type doctorProjectUseCase interface {
	DoctorProject(context.Context, application.DoctorProjectRequest) (application.DoctorProjectResult, error)
}

type doctorJSON struct {
	Project doctorProjectJSON `json:"project"`
	Status  string            `json:"status"`
	Summary doctorSummaryJSON `json:"summary"`
	Checks  []doctorCheckJSON `json:"checks"`
}

type doctorProjectJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type doctorSummaryJSON struct {
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
}

type doctorCheckJSON struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

func newDoctorCommand(service doctorProjectUseCase) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor [project]",
		Short: "Inspect whether a project is locally prepared",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := ""
			if len(args) == 1 {
				project = args[0]
			}
			result, err := service.DoctorProject(cmd.Context(), application.DoctorProjectRequest{Project: project})
			var failed *application.DoctorFailedError
			if err != nil && !errors.As(err, &failed) {
				return err
			}
			if asJSON {
				if renderErr := writeJSON(cmd, doctorJSONResult(result)); renderErr != nil {
					return renderErr
				}
			} else if renderErr := renderDoctorHuman(cmd, result); renderErr != nil {
				return renderErr
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print diagnostics as JSON")
	return cmd
}

func doctorJSONResult(result application.DoctorProjectResult) doctorJSON {
	output := doctorJSON{
		Project: doctorProjectJSON{ID: string(result.Project.ID), Name: result.Project.Name, Path: result.Project.Path},
		Status:  string(result.Status),
		Summary: doctorSummaryJSON{Passed: result.Summary.Passed, Warnings: result.Summary.Warnings, Failed: result.Summary.Failed},
		Checks:  make([]doctorCheckJSON, 0, len(result.Checks)),
	}
	for _, check := range result.Checks {
		output.Checks = append(output.Checks, doctorCheckJSON{
			Category: check.Category, Name: check.Name, Status: string(check.Status), Message: check.Message,
		})
	}
	return output
}

func renderDoctorHuman(cmd *cobra.Command, result application.DoctorProjectResult) error {
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nPath: %s\n", result.Project.ID, result.Project.Name, result.Project.Path); err != nil {
		return err
	}
	category := ""
	for _, check := range result.Checks {
		if check.Category != category {
			category = check.Category
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "\n%s:\n", strings.ToUpper(category)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %-7s %-24s %s\n", doctorStatusLabel(check.Status), check.Name, check.Message); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "\nSummary: %d passed, %d warnings, %d failed — %s\n",
		result.Summary.Passed, result.Summary.Warnings, result.Summary.Failed, result.Status)
	return err
}

func doctorStatusLabel(status core.DiagnosticStatus) string {
	return "[" + strings.ToUpper(string(status)) + "]"
}
