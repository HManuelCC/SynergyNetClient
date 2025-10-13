package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	SynergyNetClient "github.com/HManuelCC/SynergyNetClient/Socket_client"
	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

func main() {
	testConn := SynergyNetClient.NewClient("localhost", "443", "test", nil)
	mux := http.NewServeMux()

	handleTestEvents()

	authConn := SynergyNetClient.NewClient("localhost", "443", "manuel", nil)

	defer testConn.Close()
	defer authConn.Close()

	go createRoutes(mux, authConn, testConn)

	http.ListenAndServe(":8080", mux)

	select {}
}

func createRoutes(mux *http.ServeMux, client *SynergyNetClient.Client, testConn *SynergyNetClient.Client) {
	mux.HandleFunc("/login_prueba", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		event := interfaces.Event{
			Event:       "login",
			Destination: "test_js",
			Origen:      "test",
			Data: map[string]string{
				"username": username,
				"password": "test_password",
			},
		}
		fmt.Print("Enviando evento de login: ", event.Origen)

		err := testConn.Send(event, nil, func(response interfaces.State) {

			fmt.Println(response.ToString())
			// Manejar la respuesta del evento aqu
			fmt.Println("Respuesta del evento de login:", username)
			json.NewEncoder(w).Encode(username)
		})

		if err != nil {
			fmt.Println("Error sending login event:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}

func handleTestEvents() {
	SynergyNetClient.EventSlice.AddEvent("login", func(event interfaces.Event, conn net.Conn, messagePid int) {
		fmt.Println("Handling login event:", event.Data, " PIDC:", event.PID)
		var state interfaces.State = interfaces.State{
			Status:      true,
			Message:     "Login successful",
			PID:         event.PID,
			Origen:      event.Destination,
			Destination: event.Origen,
			Error:       "",
			Data:        nil,
		}
		state.SendData(conn, messagePid)
	})
}
