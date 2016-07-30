package main

import (
	"bytes"
	"fmt"
)

func parseStringTableFrame(demofile *demofile) {
	blocksize := demofile.stream.GetInt()
	buffer := make([]byte, blocksize)

	demofile.stream.Read(buffer)
	stream := NewDemoStream(bytes.NewReader(buffer))

	numberOfTables := int(stream.GetByte())
	fmt.Println(numberOfTables)
	for i := 0; i < numberOfTables; i++ {
		tableName := stream.GetString()
		fmt.Println("###tablename: ", tableName)
		//parseSringTable(stream, tableName)
		break
	}
}

func parseSringTable(stream *Demostream, tableName string) {
	numOfStrings := stream.GetInt16()

	for i := 0; i < int(numOfStrings); i++ {
		stringName := stream.GetString()
		fmt.Println("stringname: ", stringName)

		//		test := stream.GetBit()
		//	fmt.Println(test)
	}
}
