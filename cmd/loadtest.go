package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	clientlib "github.com/HManuelCC/SynergyNetClient/Socket_client"
	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

// Simple escenario de carga para crear N clientes concurrentes que envían eventos al servidor
// Ejecución:
//   go run ./cmd/loadtest.go -host localhost -port 443 -clients 100 -events 50 -rate 20
// Significa: 100 clientes, cada uno envía 50 eventos "ping" distribuidos aprox. a 20 eventos/segundo global.
// Ajusta parámetros según la capacidad.

func main() {
	var (
		host        = flag.String("host", "localhost", "Host del servidor")
		port        = flag.String("port", "443", "Puerto del servidor")
		clients     = flag.Int("clients", 50, "Numero de clientes concurrentes")
		eventsPerCl = flag.Int("events", 20, "Eventos que enviara cada cliente")
		rate        = flag.Int("rate", 0, "Target global de eventos por segundo (0 = sin control)")
		namePrefix  = flag.String("name-prefix", "c", "Prefijo para el nombre de cada cliente")
		destination = flag.String("destination", "manuel", "Cliente destino para los eventos (debe estar conectado)")
		timeoutMs   = flag.Int("timeout", 3000, "Timeout de respuesta por evento en ms")
		warmup      = flag.Duration("warmup", 2*time.Second, "Tiempo de espera antes de iniciar el envio masivo")
	)
	flag.Parse()

	fmt.Printf("Iniciando load test: host=%s port=%s clients=%d eventsPerClient=%d rate=%d eps\n",
		*host, *port, *clients, *eventsPerCl, *rate)

	// Registrar un handler basico para el evento "ping" si el servidor lo reenvia a los clientes.
	// Si no se utiliza, no afecta.
	clientlib.EventSlice.AddEvent("ping", func(ev interfaces.Event, conn net.Conn, pid int) {
		state := interfaces.State{
			Status:      true,
			Message:     "pong",
			PID:         ev.PID,
			Origen:      ev.Destination,
			Destination: ev.Origen,
			Data:        nil,
		}
		state.SendData(conn, pid)
	})

	// Crear todos los clientes primero (fase warmup / handshake)
	clientsSlice := make([]*clientlib.Client, *clients)
	for i := 0; i < *clients; i++ {
		name := fmt.Sprintf("%s%d", *namePrefix, i)
		clientsSlice[i] = clientlib.NewClient(*host, *port, name, nil)
	}

	fmt.Println("Esperando warmup para que todas las conexiones esten estables ...")
	time.Sleep(*warmup)

	var totalSent int64
	var totalOk int64
	var totalErr int64
	var totalTimeout int64
	start := time.Now()

	wg := sync.WaitGroup{}

	var throttle <-chan time.Time
	if *rate > 0 {
		interval := time.Second / time.Duration(*rate)
		throttle = time.Tick(interval)
	}

	timeout := time.Duration(*timeoutMs) * time.Millisecond

	for idx, c := range clientsSlice {
		wg.Add(1)
		go func(id int, cl *clientlib.Client) {
			defer wg.Done()
			for n := 0; n < *eventsPerCl; n++ {
				if throttle != nil {
					<-throttle
				}
				if !cl.Connected() {
					atomic.AddInt64(&totalErr, 1)
					continue
				}
				event := interfaces.Event{
					Event:       "login",      // usar el evento ya manejado en main de ejemplo
					Destination: *destination, // destino configurable
					Origen:      fmt.Sprintf("%s%d", *namePrefix, id),
					Data: map[string]string{
						"username": fmt.Sprintf("user_%d_%d", id, n),
						"password": "pwd",
					},
				}

				atomic.AddInt64(&totalSent, 1)
				err := cl.Send(event, &timeout, func(st interfaces.State) {
					if st.Status {
						atomic.AddInt64(&totalOk, 1)
					} else {
						atomic.AddInt64(&totalErr, 1)
					}
				})
				if err != nil {
					if err.Error() == fmt.Sprintf("timeout esperando respuesta para PID %d", event.PID) {
						atomic.AddInt64(&totalTimeout, 1)
					} else {
						atomic.AddInt64(&totalErr, 1)
					}
				}

				// Pequeño jitter para no alinear todos los clientes
				if *rate == 0 {
					time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				}
			}
		}(idx, c)
	}

	wg.Wait()
	dur := time.Since(start)

	fmt.Println("================= RESULTADOS =================")
	fmt.Printf("Duracion total: %s\n", dur)
	fmt.Printf("Eventos enviados: %d\n", totalSent)
	fmt.Printf("Respuestas OK : %d\n", totalOk)
	fmt.Printf("Errores       : %d\n", totalErr)
	fmt.Printf("Timeouts      : %d\n", totalTimeout)
	fmt.Printf("Events/seg (aprox): %.2f\n", float64(totalSent)/dur.Seconds())
	fmt.Printf("Goroutines vivas: %d\n", runtime.NumGoroutine())
}
