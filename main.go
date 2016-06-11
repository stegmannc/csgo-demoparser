package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"encoding/json"
	"bytes"
	"time"
	"github.com/stegmannc/csgo-demoparser/protom"
)

const (
	MAX_OSPATH = 260
	PACKET_OFFSET = 160
	DEMO_HEADER_ID = "HL2DEMO"
	DEM_SIGNON = 1
	DEM_PACKET = 2
	DEM_SYNCTICK = 3
	DEM_CONSOLECMD = 4
	DEM_USERCMD = 5
	DEM_DATATABLES = 6
	DEM_STOP = 7
	DEM_CUSTOMDATA = 8
	DEM_STRINGTABLES = 9
)

type Demoheader struct {
	Demofilestamp   [8]byte
	Demoprotocol    int32
	Networkprotocol int32
	Servername      [MAX_OSPATH]byte
	Clientname      [MAX_OSPATH]byte
	Mapname         [MAX_OSPATH]byte
	Gamedirectory   [MAX_OSPATH]byte
	Playback_time   float32
	Playback_ticks  int32
	Playback_frames int32
	Signonlength    int32
}
type democmdheader struct {
	Cmd        byte
	Tick       int32
	Playerslot byte
}
type demofile struct {
	header Demoheader
	tick   int32
	frame  int32
	stream *Demostream
}

func (d *demofile) PrintInfo() {
	fmt.Println("----HEADER START----")
	fmt.Printf("demofilestamp: %x\n", d.header.Demofilestamp)
	fmt.Printf("demoprotocol: %d\n", d.header.Demoprotocol)
	fmt.Printf("networkprotocol: %d\n", d.header.Networkprotocol)
	fmt.Printf("Server Name: %s\n", d.header.Servername)
	fmt.Printf("Client name: %s\n", d.header.Clientname)
	fmt.Printf("Mapname: %s\n", d.header.Mapname)
	fmt.Printf("Ticks: %d\n", d.header.Playback_ticks)
	fmt.Printf("Game Directory: %s\n", d.header.Gamedirectory)
	fmt.Printf("Playback time: %f seconds\n", d.header.Playback_time)
	fmt.Printf("Signon Length: %d\n", d.header.Signonlength)
	fmt.Printf("Frames: %d\n", d.header.Playback_frames)
	fmt.Printf("Ticks: %d\n", d.header.Playback_ticks)
	fmt.Println("----HEADER END----")
}
func (d *demofile) readCommandHeader() democmdheader {
	return democmdheader{Cmd: d.stream.GetByte(),
		Tick:       d.stream.GetInt(),
		Playerslot: d.stream.GetByte()}
}
func processPacket(stream *Demostream) {
	messagetype := stream.GetVarInt()
	length := stream.GetVarInt()
	message := protom.SVC_Messages(messagetype)

	if messagetype < 5 {
		return //net messages ignored
	}

	//fmt.Printf("length: %d\n", length)
	//fmt.Printf("message: %v\n", message)
	switch message {
	case protom.SVC_Messages_svc_ServerInfo:
		msg := new(protom.CSVCMsg_ServerInfo)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_SendTable         :
		msg := new(protom.CSVCMsg_SendTable)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_ClassInfo         :
		msg := new(protom.CSVCMsg_ClassInfo)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_SetPause          :
		msg := new(protom.CSVCMsg_SetPause)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_CreateStringTable :
		msg := new(protom.CSVCMsg_CreateStringTable)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_UpdateStringTable :
		msg := new(protom.CSVCMsg_UpdateStringTable)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_VoiceInit         :
		msg := new(protom.CSVCMsg_VoiceInit)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_VoiceData         :
		msg := new(protom.CSVCMsg_VoiceData)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_Print             :
		msg := new(protom.CSVCMsg_Print)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_Sounds            :
		msg := new(protom.CSVCMsg_Sounds)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_SetView           :
		msg := new(protom.CSVCMsg_SetView)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_FixAngle          :
		msg := new(protom.CSVCMsg_FixAngle)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_CrosshairAngle    :
		msg := new(protom.CSVCMsg_CrosshairAngle)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_BSPDecal          :
		msg := new(protom.CSVCMsg_BSPDecal)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_UserMessage       :
		msg := new(protom.CSVCMsg_UserMessage)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_GameEvent         :
		msg := new(protom.CSVCMsg_GameEvent)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_PacketEntities    :
		msg := new(protom.CSVCMsg_PacketEntities)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_TempEntities      :
		msg := new(protom.CSVCMsg_TempEntities)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_Prefetch          :
		msg := new(protom.CSVCMsg_Prefetch)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_Menu              :
		msg := new(protom.CSVCMsg_Menu)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_GameEventList     :
		msg := new(protom.CSVCMsg_GameEventList)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	case protom.SVC_Messages_svc_GetCvarValue      :
		msg := new(protom.CSVCMsg_GetCvarValue)
		stream.ParseToStruct(msg, length)
		printJson(msg)
	default:
		fmt.Printf("unmapped messagetype: %d\n", messagetype)
	}
}

func printJson(msg proto.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func (d *demofile) readPacket() {
	d.stream.Skip(PACKET_OFFSET)
	blocksize := d.stream.GetInt()
	//fmt.Printf("CHUNK SIZE: %d\n", blocksize)
	buffer := make([]byte, blocksize)
	d.stream.Read(buffer)
	stream := NewDemoStream(bytes.NewReader(buffer))
	go processPacket(stream)
}

type ServerClass struct {
	ClassID     int16
	DataTableID int
	Name        string
	DTName      string
}

func (d *demofile) readDatatables() {
	blocksize := d.stream.GetInt()
	//fmt.Printf("Datatable size: %d\n", blocksize)
	buffer := make([]byte, blocksize)
	d.stream.Read(buffer)
	stream := NewDemoStream(bytes.NewReader(buffer))

	dataTables := make([]*protom.CSVCMsg_SendTable, 0)

	for {
		messageType := stream.GetVarInt()
		messageLength := stream.GetVarInt()
		svcMessage := protom.SVC_Messages(messageType)

		if svcMessage.String() != protom.SVC_Messages_svc_SendTable.String() {
			panic("unexpected message")
		}

		sendTable := &protom.CSVCMsg_SendTable{}
		err := stream.ParseToStruct(sendTable, messageLength)
		if err != nil {
			panic(err)
		}
		if sendTable.GetNetTableName() == "DT_CSPlayerResource" {
			printJson(sendTable)
		}

		if sendTable.GetIsEnd() {
			break
		}

		dataTables = append(dataTables, sendTable)
	}
	fmt.Println("dataTables lenght: ", len(dataTables))

	serverClassCount := int(stream.GetInt16())
	fmt.Println("server class count ", serverClassCount)
	serverClasses := make([]*ServerClass, serverClassCount)

	for i := 0; i < serverClassCount; i++ {
		serverClass := &ServerClass{
			ClassID: stream.GetInt16(),
			Name: stream.GetDataTableString(),
			DTName: stream.GetDataTableString(),
		}
		serverClass.DataTableID = findDataTableId(dataTables, serverClass.DTName)
		fmt.Println(serverClass)
		serverClasses[i] = serverClass
	}

	fmt.Println("serverClasses lenght: ", len(serverClasses))

}

func findDataTableId(sendTables []*protom.CSVCMsg_SendTable, name string) (int) {
	for index, sendTable := range sendTables {
		if sendTable.GetNetTableName() == name {
			return index
		}
	}
	return -1
}

func (d *demofile) readStringTables() {
	blocksize := d.stream.GetInt()
	fmt.Printf("StringTables size: %d\n", blocksize)
	buffer := make([]byte, blocksize)
	d.stream.Read(buffer)
	stream := NewDemoStream(bytes.NewReader(buffer))

	numberOfTables := stream.GetByte()
	fmt.Printf("stringTables size: %d\n", numberOfTables)

}

func (d *demofile) LoopFrames() {
	for {
		cmdHeader := d.readCommandHeader()
		switch cmdHeader.Cmd {
		case DEM_SIGNON, DEM_PACKET:
			d.readPacket()
		case DEM_SYNCTICK:
			fmt.Println("skip synctick")
		case DEM_CONSOLECMD:
			fmt.Println("consolecmd")
		case DEM_USERCMD:
			fmt.Println("usercmd")
		case DEM_DATATABLES:
			d.readDatatables()
		case DEM_STOP:
			fmt.Println("STOP")
			return
		case DEM_CUSTOMDATA:
			fmt.Println("customdata")
		case DEM_STRINGTABLES:
			d.readStringTables()
		}
		d.frame++
	}
}
func (d *demofile) Open(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	d.stream = NewDemoStream(f)
	d.header = Demoheader{}
	err = binary.Read(d.stream, binary.LittleEndian, &d.header)
	if err != nil {
		panic(err)
	}
	if string(d.header.Demofilestamp[:7]) != DEMO_HEADER_ID {
		log.Fatal("Invalid demo header, are you sure this is a .dem?\n")
	}
	d.tick = 0
	d.frame = 0
}
func usage() {
	fmt.Printf("Usage: %s [demo.dem]\n", os.Args[0])
	os.Exit(2)
}
func main() {
	if len(os.Args) != 2 {
		usage()
	}
	start := time.Now()

	d := demofile{}
	d.Open(os.Args[1])
	d.PrintInfo()
	d.LoopFrames()
	elapsed := time.Since(start)
	log.Printf("Demo parsing took %s", elapsed)
}
