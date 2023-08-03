package main

import (
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/koho/udpp/config"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate new config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		var example = &config.Config{
			ID:     uuid.NewString(),
			Server: "redis://127.0.0.1:6379",
			Local:  "127.0.0.1:9999",
		}
		return example.Save("config.yml", os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	},
}
