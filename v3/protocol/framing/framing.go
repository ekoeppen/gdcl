package framing

import (
	"gitlab.com/40hz/newton/gdcl/v3/crc16"
	"gitlab.com/40hz/newton/gdcl/v3/fsm"
	"gitlab.com/40hz/newton/gdcl/v3/protocol"
)

const (
	outsidePacket = iota
	startSyn
	startDle
	insidePacket
	dleInPacket
	endEtx
	endCrc1
	packetEnd
)

const (
	none = iota
	startPacket
	addChar
	addDle
	updateCalculatedCrc
	resetReceivedCrc
	packetReceived
)

const (
	syn byte = 22
	dle byte = 16
	stx byte = 2
	etx byte = 3
)

var (
	state         = outsidePacket
	data          []byte
	receivedCrc   uint16
	calculatedCrc uint16
)

var transitions = []fsm.Transition[int, byte, int]{
	{State: outsidePacket, Event: syn, NewState: startSyn},
	{State: outsidePacket, Fallback: true, NewState: outsidePacket},
	{State: startSyn, Event: dle, NewState: startDle},
	{State: startSyn, Fallback: true, NewState: outsidePacket},
	{State: startDle, Event: stx, NewState: insidePacket, Action: startPacket},
	{State: startDle, Fallback: true, NewState: outsidePacket},
	{State: insidePacket, Event: dle, NewState: dleInPacket},
	{State: insidePacket, Fallback: true, NewState: insidePacket, Action: addChar},
	{State: dleInPacket, Event: dle, NewState: insidePacket, Action: addDle},
	{State: dleInPacket, Event: etx, NewState: endEtx, Action: updateCalculatedCrc},
	{State: dleInPacket, Fallback: true, NewState: outsidePacket},
	{State: endEtx, Fallback: true, NewState: endCrc1, Action: resetReceivedCrc},
	{State: endCrc1, Fallback: true, NewState: packetEnd, Action: packetReceived},
	{State: packetEnd, Event: syn, NewState: startSyn},
	{State: packetEnd, Fallback: true, NewState: outsidePacket},
}

func processIn(event *protocol.SerialEvent) {
	var action int
	for _, input := range event.Data {
		action, state = fsm.Input(input, state, transitions)
		switch action {
		case startPacket:
			data = make([]byte, 0, 128)
			receivedCrc = 0
			calculatedCrc = 0
		case addChar:
			data = append(data, input)
			calculatedCrc = crc16.Crc16(input, calculatedCrc)
		case addDle:
			data = append(data, input)
		case updateCalculatedCrc:
			calculatedCrc = crc16.Crc16(input, calculatedCrc)
		case resetReceivedCrc:
			receivedCrc = uint16(input)
		case packetReceived:
			receivedCrc = receivedCrc + uint16(input)<<8
			protocol.Events <- &protocol.MnpEvent{
				Direction: protocol.In,
				Data:      data,
			}
		}
	}
}

func processOut(event *protocol.MnpEvent) {
	outBuf := make([]byte, 0, len(event.Data)*2+7)
	crc := uint16(0)
	outBuf = append(outBuf, syn, dle, stx)
	for i := 0; i < len(event.Data); i++ {
		outBuf = append(outBuf, event.Data[i])
		crc = crc16.Crc16(event.Data[i], crc)
		if event.Data[i] == dle {
			outBuf = append(outBuf, dle)
		}
	}
	outBuf = append(outBuf, dle, etx)
	crc = crc16.Crc16(etx, crc)
	outBuf = append(outBuf, byte(crc&0xff), byte(crc>>8))
	protocol.Events <- &protocol.SerialEvent{
		Direction: protocol.Out,
		Data: outBuf,
	}
}

func Process(event protocol.Event) {
	switch event.(type) {
	case *protocol.SerialEvent:
		if event.(*protocol.SerialEvent).Direction == protocol.In {
			processIn(event.(*protocol.SerialEvent))
		}
	case *protocol.MnpEvent:
		if event.(*protocol.MnpEvent).Direction == protocol.Out {
			processOut(event.(*protocol.MnpEvent))
		}
	}
}
