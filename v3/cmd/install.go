package cmd

import (
	"gdcl/v3/protocol/modules/install"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial Port")
	installCmd.Flags().IntVarP(&speed, "speed", "s", 115200, "Serial Speed")
	installCmd.Flags().StringVarP(&file, "file", "f", "", "Serial Port")
}

var file string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install package",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error installing %s: %s", file, err)
		}
		install.PackageData = data
		eventLoop(port, speed, install.Process)
	},
}
