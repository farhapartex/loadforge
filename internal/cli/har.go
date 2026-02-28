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
	Use:   "reply [har file]",
	Short: "Reply a HAR file as a load test",
	Args:  cobra.ExactArgs(1),
	RunE:  harReply,
}

var harConvertedCmd = &cobra.Command{
	Use:   "conver [har file]",
	Short: "Convert a HAR file to a loadforge scanario YAML",
	Args:  cobra.ExactArgs(1),
	RunE:  harConvert,
}

func init() {
	rootCmd.AddCommand(harCmd)
	rootCmd.AddCommand(harReplayCmd)
	rootCmd.AddCommand(harConvertedCmd)
}

func harReply(cmd *cobra.Command, args []string) error {
	fmt.Printf("Replaying HAR: %s\n", args[0])
	return nil
}

func harConvert(cmd *cobra.Command, args []string) error {
	fmt.Printf("Converting HAR: %s to %s\n", args[0], output)
	return nil
}
