package cli

import (
    "fmt"
    "runtime"
    "time"

    "github.com/spf13/cobra"
)

func newVersionCmd(version, commit, date string) *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print the largo version information",
        Run: func(cmd *cobra.Command, args []string) {
            // Normalize date if it looks like an epoch or is empty
            built := date
            if built == "" || built == "unknown" || built == "0" {
                built = time.Now().Format(time.RFC3339)
            }
            fmt.Printf("largo %s\ncommit: %s\nbuilt : %s\ngo    : %s %s/%s\n",
                nz(version, "dev"), nz(commit, "none"), built, runtime.Version(), runtime.GOOS, runtime.GOARCH,
            )
        },
    }
}

func nz(s, def string) string {
    if s == "" {
        return def
    }
    return s
}

