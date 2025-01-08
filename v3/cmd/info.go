package cmd

import (
	"gdcl/v3/protocol/modules/info"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial Port")
	infoCmd.Flags().IntVarP(&speed, "speed", "s", 115200, "Serial Speed")
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info",
	Run: func(cmd *cobra.Command, args []string) {
		eventLoop(port, speed, info.Process)
	},
}
