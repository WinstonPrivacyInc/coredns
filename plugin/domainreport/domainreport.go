package domainreport

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/gorilla/websocket"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("domainreport", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}


// If we use a global instance, set it here
var globalIPL *DomainReport

type DomainReport struct {
	Next   		plugin.Handler
	port   		string
	conn 		*websocket.Conn
	mu     		sync.Mutex
	lastreport 	string
}

func setup(c *caddy.Controller) error {

	fmt.Println("[DEBUG] domainreport:setup()")
	ipl := new(DomainReport)
	useGlobal := true

	for c.Next() {
		for c.NextBlock() {
			arg, val, err := getArgLine(c)
			if err != nil {
				return err
			}
			fmt.Println("[DEBUG] domainreport:setup() arg", arg)
			switch arg {
			case `port`:
				ipl.port = val
			default:
				return fmt.Errorf("Unknown domainreport configuration directive %s", arg)
			}
		}
	}

	// If we're using the global instance, update any configuration parameters
	if useGlobal {
		if globalIPL == nil {
			globalIPL = ipl
		}
		ipl = globalIPL
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		ipl.Next = next
		return ipl
	})

	if ipl.port != "" {
		go ipl.ConnectionManager()
	}

	return nil
}

// Name implements the Handler interface.
func (ipl *DomainReport) Name() string { return "domainreport" }

// Config parsing helper
func getArgLine(c *caddy.Controller) (name, value string, err error) {
	name = strings.ToLower(c.Val())
	if !c.NextArg() {
		err = fmt.Errorf("Missing argument to %s", name)
	}
	value = c.Val()
	if c.NextArg() {
		err = fmt.Errorf("%s only takes one argument", name)
	}
	return
}


// TODO: Background loop - send ping to endpoint every 15s. If it fails, open.
func (ir *DomainReport) ConnectionManager() {
	fmt.Println("[DEBUG] Starting DomainReporter starting ConnectionManager background loop")
	ticker := time.NewTicker(INTERVAL_PERIOD)
	for {
		ir.mu.Lock()
		fmt.Println("[DEBUG] ConnectionManager trying to reconnect...")
		if ir.conn == nil {
			fmt.Println("[DEBUG] ConnectionManager: ir.conn was nil...")
			err := ir.Connect()
			if err != nil {
				fmt.Println("[DEBUG] error while trying to reconnect to domain reporting endpoint", err)
				//ir.mu.Unlock()
				//return
			}
		} else {
			fmt.Println("[WARNING] ConnectionManager: ir.conn was not nil...")

			// Send dummy message
			msg := "{\"host\":\"\"}"

			err := ir.conn.WriteMessage(websocket.TextMessage, []byte(msg))

			if err != nil {
				// Couldn't write. Try again in 15 seconds.
				ir.conn = nil
				fmt.Println("[ERROR] while writing message to domain reporting endpoint", err)
			} else {
				fmt.Println("[OK] Connection OK")
			}
		}

		ir.mu.Unlock()
		<-ticker.C

	}
}

func (ir *DomainReport) Connect() error {
	// Ensure the old connection has been closed as a failsafe.
	if ir.conn != nil {
		ir.conn.Close()
	}
	if ir.port == "" {
		return fmt.Errorf("no port provided. Cannot report domains.")
	}
	addr := "localhost:" + ir.port

	u := url.URL{Scheme: "ws", Host: addr, Path: ""}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	ir.conn = c
	if err != nil {
		fmt.Printf("[ERROR] Couldn't connect to addr", addr, err)
		return fmt.Errorf("dial:", err)
	}
	fmt.Println("[OK] Successfully connected to", addr)

	return nil
}
