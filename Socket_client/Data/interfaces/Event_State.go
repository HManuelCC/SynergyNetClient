package interfaces

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type State struct {
	Status  bool        `json:"status"`
	Message string      `json:"state"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

func (state State) SendData(client net.Conn) {

	data, err := json.Marshal(state)

	if err != nil {
		log.Println("Error al convertir el mensaje: ", err)
		return
	}

	messageBytes := []byte(string(data))

	messageSize := uint32(len(messageBytes))

	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, messageSize)

	_, err = client.Write(sizeBuffer)
	if err != nil {
		fmt.Println("Error al enviar el tamaño:", err)
		return
	}

	fmt.Println("Enviando mensaje a cliente: ", string(data))

	client.Write([]byte(string(data)))
}

func ReadData(conn net.Conn, clientName string, eventSlice *EventSlice) {

	for {
		sizeBuffer := make([]byte, 4)
		_, err := io.ReadFull(conn, sizeBuffer)
		if err != nil {
			if err == io.EOF {
				log.Println("El cliente cerro la conexión")
				return
			} else {
				fmt.Println("Error al leer el tamaño del mensaje:", err)
				break
			}

		}

		messageSize := binary.BigEndian.Uint32(sizeBuffer)

		buf := make([]byte, messageSize)
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		req := string(buf[:reqLen])
		var event Event = Event{}
		err = json.Unmarshal([]byte(req), &event)

		if err != nil {
			log.Println("Error al convertir el json")
			conn.Write([]byte("Error al convertir el json"))
		}
		fmt.Println("Evento recibido: ", event.Event)
		HandleEvents(event, conn, clientName, eventSlice)
	}
}
