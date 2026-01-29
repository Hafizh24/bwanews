package cmd

import (
	"bwanews/internal/app"

	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Bwanews server",
	Long:  `Starts the Bwanews server with the necessary configurations and database connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(StartCmd)
}
