package SynergyNetClient

import (
	"crypto/tls"
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HManuelCC/SynergyNetClient/Socket_client/Data/interfaces"
)

// Reutiliza tu EventSlice global
var EventSlice *interfaces.EventSlice = &interfaces.EventSlice{}

// Client maneja la vida de la conexión, reconexiones y envío seguro.
type Client struct {
	host   string
	port   string
	name   string
	apiKey *string

	// Conexión activa protegida por mutex de lectura/escritura.
	mu   sync.RWMutex
	conn *tls.Conn

	// Serialize writes para evitar interleaving de paquetes.
	writeMu sync.Mutex

	// Ciclo de vida
	closed  atomic.Bool
	done    chan struct{}
	attempt int

	// Estrategia de reconexión
	minBackoff time.Duration
	maxBackoff time.Duration

	// Config TLS (puedes exponer setter si luego quieres personalizarla)
	tlsConf *tls.Config
}

// NewClient crea el cliente y arranca el bucle de conexión/reconexión.
func NewClient(host, port, clientName string, apiKey *string) *Client {
	c := &Client{
		host:       host,
		port:       port,
		name:       clientName,
		apiKey:     apiKey,
		done:       make(chan struct{}),
		attempt:    1,
		minBackoff: 1 * time.Second,
		maxBackoff: 30 * time.Second,
		tlsConf: &tls.Config{
			InsecureSkipVerify: true, // ⚠️ solo para pruebas
		},
	}
	go c.run()
	return c
}

// Send envía un Event usando la conexión actual y maneja el callback/timeout
// exactamente como tu interfaces.Event.SendData, pero asegurando write serializado.
func (c *Client) Send(event interfaces.Event, timeout *time.Duration, cb ...interfaces.ResponseCallback) error {
	if c.closed.Load() {
		return errors.New("client cerrado")
	}

	conn := c.getConn()
	if conn == nil {
		return errors.New("no hay conexión activa")
	}

	// Serialize writes para evitar que dos goroutines mezclen paquetes.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return event.SendData(conn, timeout, cb...)
}

// Close cierra el cliente y la conexión actual.
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Connected indica si hay una conexión activa.
func (c *Client) Connected() bool {
	return c.getConn() != nil
}

// -------------------- Internos --------------------

func (c *Client) setConn(conn *tls.Conn) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
}

func (c *Client) clearConn() {
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = nil
	c.mu.Unlock()
}

func (c *Client) getConn() *tls.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *Client) run() {
	for {
		if c.closed.Load() {
			return
		}

		start := time.Now()
		conn, err := tls.Dial("tcp", net.JoinHostPort(c.host, c.port), c.tlsConf)
		if err != nil {
			// Falló conectar: backoff y reintentar
			delay := c.backoff()
			log.Printf("[client %s] error al conectar: %v (reintento en %s, intento %d)", c.name, err, delay, c.attempt)
			select {
			case <-time.After(delay):
				c.attempt++
				continue
			case <-c.done:
				return
			}
		}

		// Reset de intentos al conectar
		c.attempt = 1

		handshakeLatency := time.Since(start).Milliseconds()
		log.Printf("[client %s] conectado (latencia handshake: %d ms)", c.name, handshakeLatency)

		// Publicar conexión
		c.setConn(conn)

		// Lanzamos lector con el canal de estado del servidor
		serverStatus := make(chan bool, 1)
		go interfaces.ReadData(conn, c.name, EventSlice, serverStatus, float64(handshakeLatency))

		// Espera a desconexión o cierre del cliente
		select {
		case status := <-serverStatus:
			// ReadData manda false si se cae o hay error permanente
			if !status {
				log.Printf("[client %s] desconectado por el servidor, intentando reconectar...", c.name)
				c.clearConn()
				// Loop sigue: vuelve a intentar conectar
				continue
			}
		case <-c.done:
			// Cierre explícito del cliente
			c.clearConn()
			return
		}
	}
}

func (c *Client) backoff() time.Duration {
	// Exponencial con jitter
	// intento 1 => minBackoff, 2 => 2*min, 3 => 4*min, capped en maxBackoff
	base := c.minBackoff << (c.attempt - 1)
	if base > c.maxBackoff {
		base = c.maxBackoff
	}
	// Jitter +/- 20%
	jit := time.Duration(rand.Int63n(int64(base) / 5))
	if rand.Intn(2) == 0 {
		return base - jit
	}
	return base + jit
}
