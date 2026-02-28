package cli

import (
	"fmt"

	"github.com/farhapartex/loadforge/internal/config"
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

	cfg, err := config.Load(scenarioFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("  Name     : %s\n", cfg.Name)
	fmt.Printf("  Base URL : %s\n", cfg.BaseURL)
	fmt.Printf("  Scenarios: %d\n\n", len(cfg.Scenarios))

	for i, scenario := range cfg.Scenarios {
		fmt.Printf("  Scenario[%d]: %s\n", i+1, scenario.Name)
		fmt.Printf("    Weight : %d\n", scenario.Weight)
		fmt.Printf("    Steps  : %d\n", len(scenario.Steps))

		for j, step := range scenario.Steps {
			fmt.Printf("    Step[%d]: %-30s  %s %s\n",
				j+1, step.Name, step.Method, step.URL)
		}
		fmt.Println()
	}

	fmt.Println("Config is valid.")

	return nil
}
