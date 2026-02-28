package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/loader"
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
		fmt.Println("\n\nInterrupted - stopping worker gracefully ...")
		cancel()
	}()

	onTick := func(activeWorkers int) {
		fmt.Println("Workers : %d active\n", activeWorkers)
	}

	result, err := loader.Run(ctx, cfg, onTick, nil)
	if err != nil {
		return fmt.Errorf("load test failed: %w", err)
	}
	printResult(result.Metrics)

	// eng := engine.New(cfg)

	// for _, scenario := range cfg.Scenarios {
	// 	fmt.Printf(" --- Scenario: %s ---\n", scenario.Name)

	// 	for _, step := range scenario.Steps {
	// 		result := eng.ExecuteStep(step)
	// 		printResult(result)
	// 	}
	// }

	return nil
}

func printResult(m *loader.Metrics) {
	snap := m.Snapshot()
	fmt.Println("\n--------- RESULTS ---------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Total Requests\t%d\n", snap.TotalRequests)
	fmt.Fprintf(w, "Successful\t%d\n", snap.TotalSuccesses)
	fmt.Fprintf(w, "Failed\t%d\n", snap.TotalFailures)
	fmt.Fprintf(w, "Error Rate\t%.2f%%\n", snap.ErrorRate())
	fmt.Fprintf(w, "Total Data\t%s\n", formatBytes(snap.TotalBytes))
	fmt.Fprintf(w, "Duration\t%s\n", snap.Elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, "Avg RPS\t%.2f\n", snap.RPS)
	w.Flush()

	if len(snap.Latencies) > 0 {
		fmt.Println("\n--- Latency Percentiles ---")
		lw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(lw, "p50\t%v\n", percentile(snap.Latencies, 50).Round(time.Millisecond))
		fmt.Fprintf(lw, "p90\t%v\n", percentile(snap.Latencies, 90).Round(time.Millisecond))
		fmt.Fprintf(lw, "p95\t%v\n", percentile(snap.Latencies, 95).Round(time.Millisecond))
		fmt.Fprintf(lw, "p99\t%v\n", percentile(snap.Latencies, 99).Round(time.Millisecond))
		lw.Flush()
	}

	if len(snap.StatusCodes) > 0 {
		fmt.Println("\n--- Status Codes ---")
		sw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for code, count := range snap.StatusCodes {
			fmt.Fprintf(sw, "HTTP %d\t%d\n", code, count)
		}
		sw.Flush()
	}

	if len(snap.Errors) > 0 {
		fmt.Println("\n--- Errors ---")
		for errMsg, count := range snap.Errors {
			fmt.Printf("  [%d] %s\n", count, errMsg)
		}
	}

	fmt.Println("-------------------------")

}

func percentile(latencies []time.Duration, n int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted)-1) * float64(n) / 100)
	return sorted[idx]
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
