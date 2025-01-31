package main

import (
	"net"

	client_socket "github.com/HManuelCC/SynergyNetClient/Socket_client"
	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

func main() {
	client_socket.EventSlice.AddEvent("prueba", func(event interfaces.Event, conn net.Conn) {
		var state interfaces.State = interfaces.State{Status: true, Message: "Hello from client", Error: "", Data: nil}
		state.SendData(conn)
	})
	client_socket.NewClient("localhost", "4400", "Client1")

}
