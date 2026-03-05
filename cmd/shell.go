package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell <name>",
	Short: "Open a shell in a VPS instance",
	Long: `Open an interactive shell session in a VPS instance.
This uses 'lxc exec' under the hood to provide shell access.

Examples:
  vpsctl shell myserver
  vpsctl shell myserver --user root
  vpsctl shell myserver -u ubuntu`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		user, _ := cmd.Flags().GetString("user")

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		// Check if instance is running
		info, err := client.GetInstance(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Instance '%s' not found: %v\n", name, err)
			os.Exit(1)
		}

		if info.Status != "Running" {
			fmt.Fprintf(os.Stderr, "Instance '%s' is not running (status: %s)\n", name, info.Status)
			fmt.Println("Start the instance first with: vpsctl start", name)
			os.Exit(1)
		}

		// Execute shell using lxc command
		shellArgs := []string{"exec", name}
		if user != "" {
			shellArgs = append(shellArgs, "--user", user)
		}
		shellArgs = append(shellArgs, "--", "/bin/bash")

		// Check if bash exists, fallback to sh
		lxcCmd := exec.Command("lxc", shellArgs...)
		lxcCmd.Stdin = os.Stdin
		lxcCmd.Stdout = os.Stdout
		lxcCmd.Stderr = os.Stderr

		if err := lxcCmd.Run(); err != nil {
			// Try with sh if bash fails
			shellArgs[len(shellArgs)-1] = "/bin/sh"
			lxcCmd = exec.Command("lxc", shellArgs...)
			lxcCmd.Stdin = os.Stdin
			lxcCmd.Stdout = os.Stdout
			lxcCmd.Stderr = os.Stderr

			if err := lxcCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open shell: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)

	shellCmd.Flags().StringP("user", "u", "", "User to connect as (default: root)")
}
