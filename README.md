# SynergyNet Go Client (Librería)

Cliente Go listo para usarse como librería en otros proyectos. Expone tipos y funciones de forma conveniente para que puedas escribir código como:

```go
import (
		"fmt"
		SynergyNetClient "github.com/HManuelCC/SynergyNetClient/Socket_client"
)

func main() {
		c := SynergyNetClient.NewClient("localhost", "443", "my_client", nil, false)
		defer c.Close()

		SynergyNetClient.EventSlice.AddEvent("registro", func(e SynergyNetClient.Event, conn *SynergyNetClient.Client, pid int, dest string) {
				var s SynergyNetClient.State = SynergyNetClient.State{
						Status:  true,
						Message: "Hola amigo",
						Error:   "",
						Data:    nil,
						PID:     e.PID,
				}
				s.SendData(conn, pid, dest)
		})

		evt := SynergyNetClient.Event{Event: "login", Data: map[string]string{"username": "alice", "password": "pw"}}
		c.Send(evt, nil, func(resp SynergyNetClient.State) {
				fmt.Println("Respuesta:", resp.ToString())
		})
}
```

## Instalación

- Requiere Go 1.20+
- Añade el módulo a tu proyecto (si lo publicas o lo usas localmente, ajusta el path):

```sh
go get github.com/HManuelCC/SynergyNetClient@latest
```

Si estás trabajando con el repo local, usa `replace` en tu `go.mod`:

```go
replace github.com/HManuelCC/SynergyNetClient => ../ruta/local/Clients/Go
```

## API expuesta

- `NewClient(host, port, clientName string, apiKey *string, useTLS bool) *Client`
  - Crea y arranca el bucle de conexión/reconexión.
- `EventSlice`
  - Registro y eliminación de eventos: `AddEvent(name, func)`, `RemoveEvent(name)`.
- Tipos accesibles en el paquete superior:
  - `Event`, `State`, `MessageState`, `Client`, `EventsSubscribed`, `Process`, `ResponseCallback`.
  - Son alias directos de `Data/interfaces`, no se modificó funcionalidad.

## Protocolo

- Cliente→Servidor:
  - Event/State/MessageState: `[tipo(1)][tamaño(4)][JSON]` (PID dentro del JSON)
- Servidor→Cliente:
  - `[tipo(1)][PID(4)][tamaño(4)][JSON]`
- Orden de ack:
  - Al recibir `Event`: se envía `MessageState(process_status=1)`.
  - Al responder `State`: primero `MessageState(process_status=2)`, luego `State`.

## Ejemplo HTTP (como `cmd/demo.go`)

Consulta `cmd/demo.go` para un ejemplo de servidor HTTP que envía un evento `login` y registra el handler `registro`.

Arranque rápido:

```sh
cd Clients/Go
go run cmd/demo.go
```

## Notas

- No se modificó ninguna funcionalidad existente; solo se añadió `exports.go` con alias para consumir la API desde el paquete superior.
- TLS: el cliente soporta TLS (equivalente a `InsecureSkipVerify`), ajusta `useTLS` según tus certificados.
