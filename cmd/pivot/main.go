// Command pivot is the command-line entry point for Pivot.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/namdang-fdp/pivot/internal/presentation/cli"
)

func main() {
	if err := cli.NewRootCommand().ExecuteContext(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "pivot: %v\n", err)
		os.Exit(1)
	}
}
