package cmd

import (
	"fmt"
	"os"

	"github.com/kiatkoding/vpsctl/internal/lxd"
	"github.com/kiatkoding/vpsctl/pkg/utils"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new VPS instance",
	Long: `Create a new VPS instance (container or VM) using LXD.

Examples:
  vpsctl create myserver
  vpsctl create webserver --image ubuntu:22.04 --cpu 2 --memory 2GB --disk 20GB
  vpsctl create dbserver --type vm --cpu 4 --memory 4GB --ssh-key ~/.ssh/id_rsa.pub`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Validate instance name
		if err := utils.ValidateName(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		image, _ := cmd.Flags().GetString("image")
		cpu, _ := cmd.Flags().GetInt("cpu")
		memory, _ := cmd.Flags().GetString("memory")
		disk, _ := cmd.Flags().GetString("disk")
		instanceType, _ := cmd.Flags().GetString("type")
		sshKey, _ := cmd.Flags().GetString("ssh-key")
		password, _ := cmd.Flags().GetString("password")

		// Validate image
		if err := utils.ValidateImage(image); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Read SSH key file if provided
		if sshKey != "" && len(sshKey) < 50 {
			// It's a file path, read it
			data, err := os.ReadFile(sshKey)
			if err == nil {
				sshKey = string(data)
			}
		}

		// Create instance options
		opts := lxd.CreateInstanceOptions{
			Name:     name,
			Image:    image,
			CPU:      cpu,
			Memory:   memory,
			Disk:     disk,
			Type:     instanceType,
			SSHKey:   sshKey,
			Password: password,
		}

		fmt.Printf("Creating instance '%s'...\n", name)
		fmt.Printf("  Image: %s\n", image)
		fmt.Printf("  CPU: %d cores\n", cpu)
		fmt.Printf("  Memory: %s\n", memory)
		fmt.Printf("  Disk: %s\n", disk)
		fmt.Printf("  Type: %s\n", instanceType)

		client, err := lxd.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to LXD: %v\n", err)
			os.Exit(1)
		}

		info, err := client.CreateInstance(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nInstance '%s' created successfully!\n", name)

		// Get IP address
		if info.IP != "" {
			fmt.Printf("IP Address: %s\n", info.IP)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("image", "i", "ubuntu:24.04", "Image to use (e.g., ubuntu:24.04, alpine:3.19)")
	createCmd.Flags().IntP("cpu", "c", 1, "CPU cores limit")
	createCmd.Flags().StringP("memory", "m", "512MB", "Memory limit (e.g., 512MB, 2GB)")
	createCmd.Flags().StringP("disk", "d", "10GB", "Disk size (e.g., 10GB, 100GB)")
	createCmd.Flags().StringP("type", "t", "container", "Instance type: container or vm")
	createCmd.Flags().String("ssh-key", "", "SSH public key file path or content")
	createCmd.Flags().String("password", "", "Root password (will be generated if not provided)")
}
