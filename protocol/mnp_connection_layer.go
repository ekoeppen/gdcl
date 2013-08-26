package protocol

import (
	"fmt"
	"gdcl/fsm"
	"os"
	"time"
)

const (
	IDLE = iota
	LINK_REQUEST
	DATA_PHASE
	DISCONNECTING
)

const (
	LR byte = 1
	LD byte = 2
	LT byte = 4
	LA byte = 5
)

type PacketType struct {
	packetType byte
}

type OutstandingPacket struct {
	data               []byte
	sendSequenceNumber byte
}

type AckInfo struct {
	credit                byte
	receiveSequenceNumber byte
	lastAckSequenceNumber byte
}

type MNPConnectionLayer struct {
	state                      int
	stateTable                 map[int][]fsm.State
	maxInfoLength              uint16
	lastAckSequenceNumber      byte
	FromDockLink               chan []byte
	ToDockLink                 chan []byte
	FromPacketLayer            chan Packet
	ToPacketLayer              chan []byte
	writeQueue                 chan []byte
	ackChannel                 chan AckInfo
	outstandingPackets         []OutstandingPacket
	sendCreditStateVariable    byte
	receiveCreditStateVariable byte
	sendStateVariable          byte
	receiveStateVariable       byte
}

var connectionLayer MNPConnectionLayer

func (packetType PacketType) Matches(input interface{}) bool {
	return input.(*Packet).packetType == packetType.packetType
}

func sendLinkRequestResponse(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPConnectionLayer)
	packet := input.(*Packet)
	framingMode := packet.data[11]
	maxOutstanding := packet.data[14]
	dataPhaseOpt := packet.data[21]
	if dataPhaseOpt & 0x1 == 0x1 {
		layer.maxInfoLength = 256
	} else {
		layer.maxInfoLength = 64
	}
	buf := []byte{23, LR, 2,
		1, 6, 1, 0, 0, 0, 0, 255,
		2, 1, framingMode,
		3, 1, maxOutstanding,
		4, 2, 64, 0,
		8, 1, dataPhaseOpt}
	layer.outstandingPackets = make([]OutstandingPacket, 0, maxOutstanding)
	layer.ToPacketLayer <- buf
}

func handleLinkAcknowledgement(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPConnectionLayer)
	packet := input.(*Packet)
	receiveStateVariable := packet.data[0]
	adjustedCredit := int(packet.data[1]) - len(layer.outstandingPackets)
	if adjustedCredit > 0 && adjustedCredit <= 8 {
		layer.receiveCreditStateVariable = byte(adjustedCredit)
	} else {
		layer.receiveCreditStateVariable = 0
	}
	layer.ackChannel <- AckInfo{layer.receiveCreditStateVariable, receiveStateVariable, layer.lastAckSequenceNumber}
	layer.lastAckSequenceNumber = receiveStateVariable
}

func handleLinkTransfer(state int, input interface{}, output interface{}, data interface{}) {
	layer := data.(*MNPConnectionLayer)
	packet := input.(*Packet)
	layer.receiveStateVariable = packet.data[0]
	packet.data = packet.data[1:]
	layer.sendLinkAcknowledgement(layer.receiveStateVariable, 8)
	layer.ToDockLink <- packet.data
}

func closeConnection(state int, input interface{}, output interface{}, data interface{}) {
	fmt.Println("Disconnected")
	os.Exit(0)
}

func (layer *MNPConnectionLayer) sendLinkAcknowledgement(receiveSequenceNumber byte, credit byte) {
	buf := []byte{3, LA, receiveSequenceNumber, credit}
	layer.ToPacketLayer <- buf
}

func (layer *MNPConnectionLayer) sendLinkTransfer(data []byte) {
	layer.sendStateVariable++
	buf := make([]byte, len(data)+3)
	buf[0] = 2
	buf[1] = LT
	buf[2] = layer.sendStateVariable
	copy(buf[3:], data)
	layer.writeQueue <- buf
}

func (layer *MNPConnectionLayer) clearAcknowledgedPackets(ack AckInfo) {
	n := len(layer.outstandingPackets) - 1
	for n >= 0 && layer.outstandingPackets[n].sendSequenceNumber != ack.receiveSequenceNumber {
		n--
	}
	layer.outstandingPackets = layer.outstandingPackets[n+1:]
	if ack.receiveSequenceNumber == ack.lastAckSequenceNumber {
		for i := 0; i < int(layer.receiveCreditStateVariable) && i < len(layer.outstandingPackets); i++ {
			layer.ToPacketLayer <- layer.outstandingPackets[i].data
		}
	}
}

func (layer *MNPConnectionLayer) transition(packet *Packet) {
	layer.state = fsm.Transition(layer.stateTable, layer.state, packet, nil, layer)
}

func (layer *MNPConnectionLayer) reader() {
	go func() {
		for {
			packet := <-layer.FromPacketLayer
			layer.transition(&packet)
		}
	}()
}

func (layer *MNPConnectionLayer) writer() {
	go func() {
		for {
			buf := <-layer.FromDockLink
			for len(buf) > 0 {
				n := len(buf)
				if n > int(layer.maxInfoLength) {
					n = int(layer.maxInfoLength)
				}
				layer.sendLinkTransfer(buf[:n])
				buf = buf[n:]
			}
		}
	}()
}

func (layer *MNPConnectionLayer) writeQueueHandler() {
	go func() {
		for {
			if layer.receiveCreditStateVariable > 0 {
				select {
				case buf := <-layer.writeQueue:
					layer.outstandingPackets = append(layer.outstandingPackets, OutstandingPacket{buf, buf[2]})
					layer.ToPacketLayer <- buf
					layer.receiveCreditStateVariable--
				case ack := <-layer.ackChannel:
					layer.clearAcknowledgedPackets(ack)
				}
			} else {
				ack := <-layer.ackChannel
				layer.clearAcknowledgedPackets(ack)
			}
		}
	}()
}

func (layer *MNPConnectionLayer) keepAlive() {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			if layer.state == DATA_PHASE {
				layer.sendLinkAcknowledgement(layer.receiveStateVariable, 8)
			}
		}
	}()
}

func MNPConnectionLayerNew(fromPacketLayer chan Packet, toPacketLayer chan []byte) *MNPConnectionLayer {
	connectionLayer.receiveCreditStateVariable = 8
	connectionLayer.FromPacketLayer = fromPacketLayer
	connectionLayer.ToPacketLayer = toPacketLayer
	connectionLayer.FromDockLink = make(chan []byte)
	connectionLayer.ToDockLink = make(chan []byte)
	connectionLayer.writeQueue = make(chan []byte)
	connectionLayer.ackChannel = make(chan AckInfo)
	connectionLayer.stateTable = map[int][]fsm.State{
		IDLE: {{Input: PacketType{LR}, NewState: LINK_REQUEST, Action: sendLinkRequestResponse}},
		LINK_REQUEST: {
			{Input: PacketType{LR}, NewState: LINK_REQUEST},
			{Input: PacketType{LA}, NewState: DATA_PHASE},
			{Input: PacketType{LD}, NewState: IDLE},
			{Input: PacketType{LT}, NewState: IDLE},
		},
		DATA_PHASE: {
			{Input: PacketType{LR}, NewState: IDLE},
			{Input: PacketType{LA}, NewState: DATA_PHASE, Action: handleLinkAcknowledgement},
			{Input: PacketType{LD}, NewState: IDLE, Action: closeConnection},
			{Input: PacketType{LT}, NewState: DATA_PHASE, Action: handleLinkTransfer},
		},
	}
	connectionLayer.writeQueueHandler()
	connectionLayer.reader()
	connectionLayer.writer()
	connectionLayer.keepAlive()
	return &connectionLayer
}
