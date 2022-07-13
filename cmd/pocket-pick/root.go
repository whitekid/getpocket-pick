package main

import (
	"github.com/spf13/cobra"

	pocket "pocket-pick"
	"pocket-pick/config"
)

var rootCmd = &cobra.Command{
	Use:  "pocket-pick",
	RunE: func(cmd *cobra.Command, args []string) error { return pocket.New().Serve(cmd.Context()) },
}

func init() {
	config.InitFlagSet(rootCmd.Use, rootCmd.Flags())
}
