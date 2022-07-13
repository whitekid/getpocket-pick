package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	pocket "pocket-pick"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run:   func(cmd *cobra.Command, args []string) { Version() },
	})
}

// Version print version informations
func Version() {
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Compiler: %s\n", runtime.Compiler)
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Git commit: %s\n", pocket.GitCommit)
	fmt.Printf("Git branch: %s\n", pocket.GitBranch)
	fmt.Printf("Git tag: %s\n", pocket.GitTag)
	fmt.Printf("Git tree state: %s\n", pocket.GitDirty)
	fmt.Printf("Built %s\n", pocket.BuildTime)
}
