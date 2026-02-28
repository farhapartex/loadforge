package cli

import (
	"fmt"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/engine"
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

	cfg, err := config.Load(scenarioFile)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	fmt.Printf("Loaded scenario: %s\n", cfg.Name)
	fmt.Printf("Scenarios: %d\n\n", len(cfg.Scenarios))

	eng := engine.New(cfg)

	for _, scenario := range cfg.Scenarios {
		fmt.Printf(" --- Scenario: %s ---\n", scenario.Name)

		for _, step := range scenario.Steps {
			result := eng.ExecuteStep(step)
			printResult(result)
		}
	}

	return nil
}

func printResult(r *engine.Result) {
	if r.Error != nil {
		fmt.Printf("  [FAIL] %s %s\n", r.Method, r.URL)
		fmt.Printf("         Error: %v\n", r.Error)
		return
	}

	statusLabel := "OK"

	if r.StatusCode >= 400 {
		statusLabel = "ERR"
	}

	fmt.Printf("  [%s] %s %s\n", statusLabel, r.Method, r.URL)
	fmt.Printf("       Status: %d | Duration: %v | Bytes: %d\n",
		r.StatusCode, r.Duration.Round(time.Millisecond), r.BytesRead)

}
