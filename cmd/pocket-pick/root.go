package main

import (
	pocket "pocket-pick"
	"pocket-pick/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "pocket-pick",
	RunE: func(cmd *cobra.Command, args []string) error { return pocket.New().Serve(cmd.Context()) },
}

func init() {
	config.InitFlagSet(rootCmd.Use, rootCmd.Flags())
}
