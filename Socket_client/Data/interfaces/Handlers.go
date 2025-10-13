package interfaces

import (
	"encoding/json"
	"log"
	"net"
)

func HandleEvents(e Event, conn net.Conn, clientName string, eventSlice *EventSlice, latency int, messagePid int) {

	//fmt.Println("Procesando evento:", e.Event)
	if e.Event == "connect" {

		clientHardwareStats := &ClientHardwareResourcesStatistics{}
		err := clientHardwareStats.GetSystemStats()
		if err != nil {
			log.Fatalln("Error al obtener las estadísticas del sistema:", err)
		}

		var clientInfo *ClientInformation = &ClientInformation{
			ClientName: clientName,
			Latency:    float64(latency),
			Resources:  *clientHardwareStats,
		}

		jsonData, err := json.Marshal(clientInfo)
		//fmt.Println("Información del cliente:", string(jsonData))

		if err != nil {
			log.Println("Error al convertir la información del cliente a JSON:", err)
			var state *State = &State{Origen: clientName, Destination: "127.0.0.1", Message: "Error al convertir la información del cliente a JSON", Status: false, Data: nil}
			state.SendData(conn, messagePid)
			return
		}

		var state *State = &State{Origen: clientName, Destination: "127.0.0.1", Message: "Cliente conectado con exito.", Status: true, Data: string(jsonData)}
		state.SendData(conn, messagePid)

		return
	}
	for _, v := range *eventSlice {
		if v.Name == e.Event {
			v.EventProccess(e, conn, messagePid)
			return
		}
	}
	log.Println("Evento no reconocido:", messagePid)
	var state *State = &State{Message: "No se puede reconocer el evento", Status: false, Data: nil, Origen: e.Destination, Destination: e.Origen, PID: e.PID, Error: "Evento no reconocido"}
	state.SendData(conn, messagePid)
}
