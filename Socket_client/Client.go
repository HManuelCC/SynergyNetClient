package SynergyNetClient

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

var EventSlice *interfaces.EventSlice = &interfaces.EventSlice{}

func NewClient(host string, port string, clientName string) {
	var Attempt *int = new(int)
	*Attempt = 1
	ConnectToServer(host, port, clientName, Attempt)
}

func ConnectToServer(host string, port string, clientName string, attempt *int) {
	var timeout int = 5
	var serverStatus chan bool = make(chan bool)
	var connection bool = true
	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		log.Println("Error al conectar al servidor: ", err)
		connection = false
	} else {
		log.Println("Connected to server")
		defer conn.Close()

		go interfaces.ReadData(conn, clientName, EventSlice, serverStatus)
	}

	if !connection {
		fmt.Println("Reconectando al servidor en ", timeout, " segundos, intento: ", *attempt)
		*attempt = *attempt + 1
		time.Sleep(time.Duration(timeout) * time.Second)
		NewClient(host, port, clientName)
	}

	select {
	case status := <-serverStatus:
		if !status {
			fmt.Println("Reconectando al servidor en ", timeout, " segundos, intento: ", *attempt)
			*attempt = *attempt + 1
			time.Sleep(time.Duration(timeout) * time.Second)
			NewClient(host, port, clientName)
		}

	case <-time.After(time.Duration(timeout) * time.Hour):
		log.Println("No se pudo reconectar al servidor")
		return
	}
}
