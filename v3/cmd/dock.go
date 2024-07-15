package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/dock"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/framing"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/mnp"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/serial"
)

var port string

func init() {
	rootCmd.AddCommand(dockCmd)
	dockCmd.Flags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial Port")
}

func eventLoop() {
	fmt.Println("Starting event loop")
	for {
		event := <-protocol.Events
		fmt.Printf("Event: %s\n", event)

		serial.Process(event)
		framing.Process(event)
		mnp.Process(event)
		dock.Process(event)

		if protocol.IsQuitEvent(event) {
			break
		}
	}
	fmt.Println("Event loop complete")
}

var dockCmd = &cobra.Command{
	Use:   "dock",
	Short: "Docking commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Connecting to %s\n", port)
		go serial.SerialLoop(port)
		eventLoop()
	},
}
