package domainreport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"

	"github.com/miekg/dns"
	"github.com/gorilla/websocket"

)

var INTERVAL_PERIOD = 15 * time.Second

type InterceptRecorder struct {
	dns.ResponseWriter
	ipl *DomainReport
}

func (ipl *DomainReport) NewInterceptRecorder(w dns.ResponseWriter) *InterceptRecorder {

	i := &InterceptRecorder{
		ResponseWriter: w,
		ipl:            ipl,
	}

	return i
}


// WriteMsg records the status code and calls the
// underlying ResponseWriter's WriteMsg method.
func (ir *InterceptRecorder) WriteMsg(res *dns.Msg) error {


	if len(res.Question) > 0 {
		host := res.Question[0].Name
		// Don't report reverse IP lookups
		if host != "" && !strings.Contains(host, "in-addr.arpa") {
			// The rest of this can be done in the background so we don't delay the response
			go func() {
				ir.ipl.mu.Lock()
				defer ir.ipl.mu.Unlock()

				// If no websocket, then open one
				if ir.ipl.conn == nil {
					fmt.Println("[ERROR] domainreport connection is not open")
					return
				}

				// Prevent spamming the firmware with bursts of the same request
				if host == ir.ipl.lastreport {
					return
				}
				ir.ipl.lastreport = host
				// Send message
				msg := "{\"host\":\"" + host + "\"}"

				err := ir.ipl.conn.WriteMessage(websocket.TextMessage, []byte(msg))

				if err != nil {
					fmt.Println("[ERROR] while writing message to domain reporting endpoint", err)
				} else {
					fmt.Println("[INFO] Reported host", host)
				}

			}()
		}
	}

	return ir.ResponseWriter.WriteMsg(res)
}

// ServeDNS implements the plugin.Handler interface.
func (ipl *DomainReport) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	return plugin.NextOrFailure(ipl.Name(), ipl.Next, ctx, ipl.NewInterceptRecorder(w), r)
}
