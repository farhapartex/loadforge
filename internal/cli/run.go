package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	workers  int
	duration string
	output   string
	varFlags []string
	envFile  string
)

var runCmd = &cobra.Command{
	Use:   "run [scenario file]",
	Short: "Run a load test from a scenario YAML file",
	Args:  cobra.ExactArgs(1),
	RunE:  runScenario,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(
		&workers, "workers", "w", 0, "Override number of concurrent workers",
	)
	runCmd.Flags().StringVarP(
		&duration, "duration", "d", "", "Override test duarion (ex: 30s, 2m, 1h)",
	)
	runCmd.Flags().StringVarP(
		&output, "output", "o", "", "Output report file",
	)
	runCmd.Flags().StringArrayVar(
		&varFlags, "var", []string{}, "Set a variable (ex: --var base_url=https://example.com)",
	)
	runCmd.Flags().StringVar(
		&envFile, "env-file", "", "Path to a .env file to load variables from",
	)
}

func runScenario(cmd *cobra.Command, args []string) error {
	scenarioFile := args[0]
	fmt.Printf("Loading scenario: %s\n", scenarioFile)

	if workers > 0 {
		fmt.Printf("Workers override: %d\n", workers)
	}

	if duration != "" {
		fmt.Printf("Durnation override: %s\n", duration)
	}

	if output != "" {
		fmt.Printf("Output report: %s\n", output)
	}

	return nil
}
