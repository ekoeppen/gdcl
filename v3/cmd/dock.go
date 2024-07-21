package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/dock"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/framing"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/mnp"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/modules/info"
	"gitlab.com/40hz/newton/gdcl/v3/protocol/serial"
	"log"
)

const (
	noModule = iota
	infoModule
	loadPackageModule
)

var (
	port         string
	activeModule int
	logSerial    bool
	logMnp       bool
	logDock      bool
)

func init() {
	rootCmd.AddCommand(dockCmd)
	dockCmd.Flags().StringVarP(&port, "port", "p", "/dev/ttyUSB0", "Serial Port")
}

func logEvent(event protocol.Event) {
	switch event.(type) {
	case *protocol.SerialEvent:
		if logSerial {
			log.Println(event)
		}
	case *protocol.MnpEvent:
		if logMnp {
			log.Println(event)
		}
	case *protocol.DockEvent:
		if logDock {
			log.Println(event)
		}
	}
}

func eventLoop() {
	log.Println("Starting event loop")
	logDock = true
	for {
		event := <-protocol.Events
		logEvent(event)

		serial.Process(event)
		framing.Process(event)
		mnp.Process(event)
		dock.Process(event)

		switch activeModule {
		case infoModule:
			info.Process(event)
		case loadPackageModule:
			break
		}

		if protocol.IsQuitEvent(event) {
			break
		}
	}
	log.Println("Event loop complete")
}

var dockCmd = &cobra.Command{
	Use:   "dock",
	Short: "Docking commands",
	Run: func(cmd *cobra.Command, args []string) {
		activeModule = infoModule
		log.Printf("Connecting to %s\n", port)
		go serial.SerialLoop(port)
		eventLoop()
	},
}
