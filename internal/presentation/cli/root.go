// Package cli implements Pivot's command-line presentation layer.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand constructs an isolated Pivot command tree.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pivot",
		Short:         "Safely orchestrate local project contexts",
		Long:          "Pivot safely orchestrates complete local development contexts.\n\nSwitch projects, keep your flow.",
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

	cmd.AddCommand(newVersionCommand())
	return cmd
}
