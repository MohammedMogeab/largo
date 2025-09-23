package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// Execute runs the root command.
func Execute(version, commit, date string) {
    root := newRootCmd(version, commit, date)
    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func newRootCmd(version, commit, date string) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "largo",
        Short: "LarGo CLI — scaffold and manage Go web apps",
        Long:  "LarGo is a Laravel-style Go web framework. Use this CLI to scaffold apps, generate code, run migrations, and more.",
        SilenceUsage:  true,
        SilenceErrors: true,
    }

    // Add subcommands
    cmd.AddCommand(
        newVersionCmd(version, commit, date),
        newNewCmd(),
        newServeCmd(),
        newMakeControllerCmd(),
        newMakeMigrationCmd(),
        newMigrateCmd(),
        newMigrateRollbackCmd(),
        newMigrateStatusCmd(),
    )

    return cmd
}
