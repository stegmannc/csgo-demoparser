package demoinfo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/segmentio/go-camelcase"
	"github.com/stegmannc/csgo-demoparser/protom"
)

const (
	MaxOSPath      = 260
	PacketOffset   = 160
	DemoHeaderID   = "HL2DEMO"
	DemSignon      = 1
	DemPacket      = 2
	DemSynctick    = 3
	DemConsoleCMD  = 4
	DemUserCMD     = 5
	DemDatatables  = 6
	DemStop        = 7
	DemCustomdata  = 8
	DemSringTables = 9
)

type DemoHeader struct {
	Demofilestamp   [8]byte
	Demoprotocol    int32
	Networkprotocol int32
	Servername      [MaxOSPath]byte
	Clientname      [MaxOSPath]byte
	Mapname         [MaxOSPath]byte
	Gamedirectory   [MaxOSPath]byte
	PlaybackTime    float32
	PlaybackTicks   int32
	PlaybackFrames  int32
	Signonlength    int32
}

type DemoCmdHeader struct {
	Cmd        uint8
	Tick       int32
	Playerslot uint8
}

type DemoParser struct {
	Header  *DemoHeader
	Context *DemoContext
	stream  *DemoStream
}

type ServerClass struct {
	ClassID     int16
	DataTableID int
	Name        string
	DTName      string
}

func (dp *DemoParser) PrintHeader() {
	fmt.Println("----HEADER START----")
	fmt.Printf("demofilestamp: %x\n", dp.Header.Demofilestamp)
	fmt.Printf("demoprotocol: %d\n", dp.Header.Demoprotocol)
	fmt.Printf("networkprotocol: %d\n", dp.Header.Networkprotocol)
	fmt.Printf("Server Name: %s\n", dp.Header.Servername)
	fmt.Printf("Client name: %s\n", dp.Header.Clientname)
	fmt.Printf("Mapname: %s\n", dp.Header.Mapname)
	fmt.Printf("Game Directory: %s\n", dp.Header.Gamedirectory)
	fmt.Printf("Playback time: %f seconds\n", dp.Header.PlaybackTime)
	fmt.Printf("Signon Length: %d\n", dp.Header.Signonlength)
	fmt.Printf("Playback Frames: %d\n", dp.Header.PlaybackFrames)
	fmt.Printf("Playback Ticks: %d\n", dp.Header.PlaybackTicks)
	fmt.Println("----HEADER END----")
}

func (dp *DemoParser) parseMessage(stream *DemoStream, tick int32) {
	messageType := stream.GetVarInt()
	length := stream.GetVarInt()

	switch protom.SVC_Messages(messageType) {
	//case protom.SVC_Messages_svc_CreateStringTable:
	//	msg := new(protom.CSVCMsg_CreateStringTable)
	//	stream.ParseToStruct(msg, length)
	//	printJSON(msg)
	//case protom.SVC_Messages_svc_UpdateStringTable:
	//	msg := new(protom.CSVCMsg_UpdateStringTable)
	//	stream.ParseToStruct(msg, length)
	//	printJSON(msg)
	//case protom.SVC_Messages_svc_UserMessage:
	//	msg := new(protom.CSVCMsg_UserMessage)
	//	stream.ParseToStruct(msg, length)
	//	printJSON(msg)
	case protom.SVC_Messages_svc_GameEvent:
		msg := &protom.CSVCMsg_GameEvent{}
		stream.ParseToStruct(msg, length)
		descriptor := dp.Context.GetGameEventDescriptor(msg.GetEventid())
		event := NewDemoGameEvent(msg.GetEventid(), descriptor.GetName(), tick)
		descriptorKeys := descriptor.GetKeys()
		eventKeys := msg.GetKeys()

		for i, eventKey := range eventKeys {
			descriptorKey := descriptorKeys[i]
			name := camelcase.Camelcase(descriptorKey.GetName())
			mappedValue := mapGameEventKeyValue(descriptorKey.GetType(), eventKey)
			event.addData(name, mappedValue)
		}
		dp.Context.GameEventChan <- event
	//case protom.SVC_Messages_svc_PacketEntities:
	//	msg := new(protom.CSVCMsg_PacketEntities)
	//	stream.ParseToStruct(msg, length)
	//	printJSON(msg)
	case protom.SVC_Messages_svc_GameEventList:
		msg := &protom.CSVCMsg_GameEventList{}
		stream.ParseToStruct(msg, length)
		dp.Context.GameEventList = msg
	default:
		stream.Skip(int64(length))
	}
}

func mapGameEventKeyValue(valueType int32, key *protom.CSVCMsg_GameEventKeyT) interface{} {
	switch valueType {
	case 1:
		return key.GetValString()
	case 2:
		return key.GetValFloat()
	case 3:
		return key.GetValLong()
	case 4:
		return key.GetValShort()
	case 5:
		return key.GetValByte()
	case 6:
		return key.GetValBool()
	default:
		return nil
	}
}

func (dp *DemoParser) parseDemoPacket(stream *DemoStream, context *DemoContext, tick int32) {
	stream.Skip(PacketOffset)
	packetStream := stream.CreatePacketStream()
	for !packetStream.IsProcessed() {
		dp.parseMessage(packetStream, tick)
	}
}

func (dp *DemoParser) parseDatatables() {
	dataTablesStream := dp.stream.CreatePacketStream()
	dataTables := make([]*protom.CSVCMsg_SendTable, 0)

	for {
		messageType := dataTablesStream.GetVarInt()
		messageLength := dataTablesStream.GetVarInt()
		svcMessage := protom.SVC_Messages(messageType)

		if svcMessage.String() != protom.SVC_Messages_svc_SendTable.String() {
			panic("unexpected message")
		}

		sendTable := &protom.CSVCMsg_SendTable{}
		if err := dataTablesStream.ParseToStruct(sendTable, messageLength); err != nil {
			panic(err)
		}

		if sendTable.GetIsEnd() {
			break
		}

		dataTables = append(dataTables, sendTable)
	}
	fmt.Println("dataTables lenght: ", len(dataTables))

	serverClassCount := int(dataTablesStream.GetInt16())
	serverClasses := make([]*ServerClass, serverClassCount)

	for i := 0; i < serverClassCount; i++ {
		serverClass := &ServerClass{
			ClassID: dataTablesStream.GetInt16(),
			Name:    dataTablesStream.GetDataTableString(),
			DTName:  dataTablesStream.GetDataTableString(),
		}
		serverClass.DataTableID = findDataTableID(dataTables, serverClass.DTName)
		//fmt.Println(serverClass)
		serverClasses[i] = serverClass
	}

	fmt.Println("serverClasses lenght: ", len(serverClasses))

}

func findDataTableID(sendTables []*protom.CSVCMsg_SendTable, name string) int {
	for index, sendTable := range sendTables {
		if sendTable.GetNetTableName() == name {
			return index
		}
	}
	return -1
}

func (d *DemoParser) parseStringTables() {
	stream := d.stream.CreatePacketStream()
	numberOfTables := stream.GetByte()
	fmt.Printf("stringTables size: %d\n", numberOfTables)

}

func (dp *DemoParser) ParseTicks() {
	stream := dp.stream
	for {
		cmdHeader := stream.readCommandHeader()
		switch cmdHeader.Cmd {
		case DemSignon, DemPacket:
			dp.parseDemoPacket(stream, dp.Context, cmdHeader.Tick)
		case DemSynctick:
			fmt.Println("skip synctick")
		case DemConsoleCMD:
			fmt.Println("consolecmd")
		case DemUserCMD:
			fmt.Println("usercmd")
		case DemDatatables:
			dp.parseDatatables()
		case DemStop:
			fmt.Println("STOP")
			dp.Context.StopChan <- true
			close(dp.Context.StopChan)
			close(dp.Context.GameEventChan)
			return
		case DemCustomdata:
			fmt.Println("customdata")
		case DemSringTables:
			dp.parseStringTables()
		}
	}
}

func NewDemoParser(path string) (*DemoParser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	header := &DemoHeader{}
	stream := NewDemoStream(f, -1)

	if err = binary.Read(stream, binary.LittleEndian, header); err != nil {
		return nil, err
	}
	if string(header.Demofilestamp[:7]) != DemoHeaderID {
		return nil, errors.New("Invalid demo header, are you sure this is a .dem?")
	}

	parser := &DemoParser{
		Header:  header,
		Context: NewDemoContext(header),
		stream:  stream,
	}
	return parser, nil
}
