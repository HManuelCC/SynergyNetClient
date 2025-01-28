package SynergyNetClient

import (
	"log"
	"net"

	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

var EventSlice *interfaces.EventSlice = &interfaces.EventSlice{}

func NewClient(host string, port string, clientName string) {
	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		panic(err)
	} else {
		log.Println("Connected to server")
	}

	defer conn.Close()

	go interfaces.ReadData(conn, clientName, EventSlice)

	select {}

}
