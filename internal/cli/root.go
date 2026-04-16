package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const installPath = "/usr/local/bin/loadforge"
const installPathWeb = "/usr/local/bin/loadforge-web"

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

	stopService()

	for _, path := range []string{installPath, installPathWeb} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w (try running with sudo)", path, err)
		}
		fmt.Printf("Removed %s\n", path)
	}

	appDir, err := appDataDir()
	if err == nil {
		if err := os.RemoveAll(appDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove %s: %v\n", appDir, err)
		} else {
			fmt.Printf("Removed %s\n", appDir)
		}
	}

	fmt.Println("LoadForge has been fully uninstalled.")
	return nil
}

func stopService() {
	switch runtime.GOOS {
	case "linux":
		run("systemctl", "stop", "loadforge-web")
		run("systemctl", "disable", "loadforge-web")
		os.Remove("/etc/systemd/system/loadforge-web.service")
		run("systemctl", "daemon-reload")
		fmt.Println("Removed systemd service")
	case "darwin":
		plist := "/Library/LaunchDaemons/com.loadforge.web.plist"
		run("launchctl", "unload", "-w", plist)
		os.Remove(plist)
		fmt.Println("Removed launchd service")
	}
}

func run(name string, args ...string) {
	_ = exec.Command(name, args...).Run()
}

func appDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".loadforge"), nil
}
