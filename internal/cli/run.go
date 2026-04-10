package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/loader"
	"github.com/farhapartex/loadforge/internal/ui"
	"github.com/spf13/cobra"
)

var (
	workers  int
	duration string
	output   string
	varFlags []string
	envFile  string
	noUI     bool
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
		&duration, "duration", "d", "", "Override test duration (ex: 30s, 2m, 1h)",
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
	runCmd.Flags().BoolVar(
		&noUI, "no-ui", false, "Disable the terminal UI, print plain text instead",
	)
}

func runScenario(cmd *cobra.Command, args []string) error {
	scenarioFile := args[0]

	cfg, err := config.LoadFromFile(scenarioFile)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	if workers > 0 {
		cfg.Load.Workers = workers
	}
	if duration != "" {
		cfg.Load.Duration = duration
	}

	fmt.Printf("Loaded scenario: %s\n", cfg.Name)
	fmt.Printf("Scenarios: %d\n\n", len(cfg.Scenarios))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nInterrupted - stopping workers gracefully ...")
		cancel()
	}()

	metricCh := make(chan loader.MetricsSnapshot, 10)
	doneCh := make(chan struct{})

	var runErr error
	go func() {
		_, runErr = loader.Run(ctx, cfg, nil, metricCh, doneCh)
	}()

	if noUI {
		<-doneCh
		for len(metricCh) > 0 {
			<-metricCh
		}
		fmt.Printf("\nScenario : %s\n", cfg.Name)
		fmt.Printf("Base URL : %s\n\n", cfg.BaseURL)
		return runErr
	}

	model := ui.NewModel(
		cfg.Name,
		cfg.BaseURL,
		cfg.Load.Profile,
		cfg.Load.Duration,
		metricCh,
		doneCh,
	)

	if err := ui.Run(model); err != nil {
		return fmt.Errorf("ui error: %w", err)
	}

	return runErr
}
