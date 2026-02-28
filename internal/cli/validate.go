package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = cobra.Command{
	Use:   "validate [scenario file]",
	Short: "Validate a scenario YAML file without runnint it",
	Args:  cobra.ExactArgs(1),
	RunE:  validateScenario,
}

func init() {
	rootCmd.AddCommand(&validateCmd)
}

func validateScenario(cmd *cobra.Command, args []string) error {
	scenarioFile := args[0]
	fmt.Printf("Validating: %s\n", scenarioFile)

	return nil
}
