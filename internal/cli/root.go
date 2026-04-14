package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const installPath = "/usr/local/bin/loadforge"

var uninstall bool

var rootCmd = &cobra.Command{
	Use:   "loadforge",
	Short: "A powerful HTTP load testing tool",
	Long:  "loadforge is a developer first HTTP load testing tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		if uninstall {
			return runUninstall()
		}
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Uninstall loadforge from the system")
}

func runUninstall() error {
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("loadforge is not installed at %s", installPath)
	}

	if err := os.Remove(installPath); err != nil {
		return fmt.Errorf("failed to remove %s: %w (try running with sudo)", installPath, err)
	}
	fmt.Printf("Removed %s\n", installPath)

	appDir, err := appDataDir()
	if err == nil {
		if err := os.RemoveAll(appDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove %s: %v\n", appDir, err)
		} else {
			fmt.Printf("Removed %s\n", appDir)
		}
	}

	fmt.Println("loadforge has been uninstalled.")
	return nil
}

func appDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".loadforge"), nil
}
