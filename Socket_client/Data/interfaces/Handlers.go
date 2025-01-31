package interfaces

import (
	"net"
)

func HandleEvents(e Event, conn net.Conn, clientName string, eventSlice *EventSlice) {
	if e.Event == "connect" {
		var state *State = &State{Message: clientName, Status: true, Data: nil}
		state.SendData(conn)
		return
	}
	for _, v := range *eventSlice {
		if v.Name == e.Event {
			v.EventProccess(e, conn)
			return
		}

	}

	var state *State = &State{Message: "No se puede reconocer el evento", Status: false, Data: nil}
	state.SendData(conn)
}
