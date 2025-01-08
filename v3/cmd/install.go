package cmd

import (
	"gdcl/v3/protocol/modules/info"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial Port")
	installCmd.Flags().IntVarP(&speed, "speed", "s", 115200, "Serial Speed")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install package",
	Run: func(cmd *cobra.Command, args []string) {
		eventLoop(port, speed, info.Process)
	},
}
