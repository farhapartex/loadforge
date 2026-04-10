package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var harCmd = &cobra.Command{
	Use:   "har",
	Short: "HAR file utilities",
}

var harReplayCmd = &cobra.Command{
	Use:   "replay [har file]",
	Short: "Replay a HAR file as a load test",
	Args:  cobra.ExactArgs(1),
	RunE:  harReplay,
}

var harConvertCmd = &cobra.Command{
	Use:   "convert [har file]",
	Short: "Convert a HAR file to a loadforge scenario YAML",
	Args:  cobra.ExactArgs(1),
	RunE:  harConvert,
}

func init() {
	harCmd.AddCommand(harReplayCmd)
	harCmd.AddCommand(harConvertCmd)
	rootCmd.AddCommand(harCmd)
}

func harReplay(cmd *cobra.Command, args []string) error {
	fmt.Printf("HAR replay is not yet implemented: %s\n", args[0])
	return nil
}

func harConvert(cmd *cobra.Command, args []string) error {
	fmt.Printf("HAR convert is not yet implemented: %s\n", args[0])
	return nil
}
