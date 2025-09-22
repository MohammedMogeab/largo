package cli

import (
    "fmt"

    "github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "new <app>",
        Short: "Scaffold a new LarGo application",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Placeholder: actual scaffolding will render embedded templates.
            name := args[0]
            fmt.Fprintf(cmd.OutOrStdout(), "Scaffolding for %q coming soon.\n", name)
            return nil
        },
    }
    return cmd
}

