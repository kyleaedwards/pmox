package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kyleaedwards/pmox/api"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pmox",
	Short: "[pmox] is a set of utilities for interacting with Proxmox VE",
	Long:  "[pmox] is a set of utilities for interacting with Proxmox VE",
}

var cmdVmip4 = &cobra.Command{
	Use:   "ipv4 [vm name]",
	Short: "Fetch the local network IPv4 for a VM",
	Long: `Attempts to detect a virtual machine's local network IP.
Will not return an IP if there is no working network
device or if the Proxmox agent is unable to find one.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		proxmox, _ := api.FromContext(cmd.Context())
		ip, err := proxmox.FindIpAddress(name, "ipv4")
		if err != nil {
			Fatal(err)
		}
		fmt.Println(ip)
	},
}

var cmdVmip6 = &cobra.Command{
	Use:   "ipv6 [vm name]",
	Short: "Fetch the local network IPv4 for a VM",
	Long: `Attempts to detect a virtual machine's local network IP.
Will not return an IP if there is no working network
device or if the Proxmox agent is unable to find one.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		proxmox, _ := api.FromContext(cmd.Context())
		ip, err := proxmox.FindIpAddress(name, "ipv6")
		if err != nil {
			Fatal(err)
		}
		fmt.Println(ip)
	},
}

var ssh = &cobra.Command{
	Use:   "ssh [user]@[vm name]",
	Short: "Opens a secure shell to a VM",
	Long:  "Opens a secure shell to a VM.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parts := strings.Split(args[0], "@")
		if len(parts) < 2 {
			Fatal("Connection string must be in the format [user]@[vm name].")
		}
		user := parts[0]
		vm := parts[1]

		proxmox, _ := api.FromContext(cmd.Context())

		ip, err := proxmox.FindIpAddress(vm, "ipv4")
		if err != nil {
			ip, err = proxmox.FindIpAddress(vm, "ipv6")
		}
		if err != nil {
			Fatalf("Cannot determine IP for VM \"%s\".", vm)
		}

		shell := exec.Command("ssh", fmt.Sprintf("%s@%s", user, ip))
		shell.Stdout = os.Stdout
		shell.Stdin = os.Stdin
		shell.Stderr = os.Stderr
		shell.Run()
	},
}

func init() {
	rootCmd.AddCommand(cmdVmip4)
	rootCmd.AddCommand(cmdVmip6)
	rootCmd.AddCommand(ssh)
}

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		Fatal(err)
	}
}

func Fatal(v ...any) {
	for _, entry := range v {
		fmt.Fprintf(os.Stderr, "[error] %s\n", entry)
		os.Exit(1)
	}
}

func Fatalf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("[error] %s\n", format), v...)
	os.Exit(1)
}
