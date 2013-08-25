package main

import (
	"gdcl/nsof"
	"gdcl/protocol"
	"io/ioutil"
	"os"
	"log"
	"encoding/hex"
)

func nsofTest() {
	bytes, _ := ioutil.ReadFile(os.Args[1])
	var in nsof.Data
	var out nsof.Data
	in = bytes[1:]
	objects := in.Factory()
	out = append(out, 2)
	objects[0].WriteNSOF(&out)
	log.Printf("\n%s\n", hex.Dump(out))
	ioutil.WriteFile("/tmp/g", out, 0644)
}

func dataHandler(receivedData <-chan protocol.DantePacket, sentData chan<- protocol.DantePacket, commands <-chan byte) {
	for {
		select {
		case packet := <-receivedData:
			log.Printf("%x\n", packet)
		case command := <-commands:
			log.Printf("Command: %d\n", command)
		}
	}
}

func commandReader() chan byte {
	out := make(chan byte)
	go func() {
		for {
			log.Println("Waiting for command")
			command := make([]byte, 1)
			for command[0] < 32 {
				os.Stdin.Read(command)
			}
			out <- command[0]
		}
	}()
	return out
}

func serialTest(session byte) {
	log.SetOutput(os.Stdout)
	commandChannel := commandReader()
	packetLayer := protocol.MNPPacketLayerNew(os.Args[1], 115200)
	connectionLayer := protocol.MNPConnectionLayerNew(packetLayer.ToConnection, packetLayer.FromConnection)
	dockLinkLayer := protocol.DockLinkLayerNew(connectionLayer.ToDockLink, connectionLayer.FromDockLink)
	connectModule := protocol.ConnectModuleNew(dockLinkLayer.FromApplication, session)
	dockLinkLayer.AddModule(connectModule)
	if session == protocol.SESSION_LOAD_PACKAGE {
		buf, _ := ioutil.ReadFile(os.Args[2])
		loadPackageModule := protocol.LoadPackageModuleNew(dockLinkLayer.FromApplication, buf)
		dockLinkLayer.AddModule(loadPackageModule)
	}
	dataHandler(dockLinkLayer.ToApplication, dockLinkLayer.FromApplication, commandChannel)
}

func main() {
	//nsofTest()
	serialTest(protocol.SESSION_LOAD_PACKAGE)
}
