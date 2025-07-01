package cmd

import (
	"fmt"
	"log"

	"github.com/callumalpass/handwrite/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration files",
	Long:  `Manage configuration files for the handwrite tool.`,
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create the default configuration file",
	Long:  `Create the default configuration file at ~/.config/handwrite/config.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.SetupDefaultConfig(); err != nil {
			log.Fatalf("Failed to setup config: %v", err)
		}

		configPath := config.GetDefaultConfigPath()
		fmt.Printf("Created default configuration at: %s\n", configPath)
	},
}

func init() {
	configCmd.AddCommand(setupCmd)
}

