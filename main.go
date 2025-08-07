// @title Simple ICMP Ping API
// @version 1.0
// @description This is a demo API to ICMP-ping an IP address.
// @host localhost:8080
// @BasePath /

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "example.com/hello/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/tatsushid/go-fastping"
)

func main() {
	http.HandleFunc("/ping", handlePing)
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handlePing godoc
// @Summary Ping an IP address via ICMP
// @Description Sends an ICMP Echo Request to the given IP address
// @Produce json
// @Param ip query string true "IP Address"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Missing IP"
// @Router /ping [get]
func handlePing(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing 'ip' query parameter", http.StatusBadRequest)
		return
	}

	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	p := fastping.NewPinger()
	p.AddIPAddr(ra)

	result := map[string]interface{}{
		"reachable": false,
		"ip":        ip,
	}

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) { //cevap geldiyse
		result["reachable"] = true
		result["rtt_ms"] = rtt.Milliseconds() //round-trip time
	}

	p.OnIdle = func() {
		w.Header().Set("Content-Type", "application/json")
		if result["reachable"].(bool) {
			fmt.Fprintf(w, `{"reachable": true, "ip": "%s", "rtt_ms": %d}`, ip, result["rtt_ms"])
		} else {
			fmt.Fprintf(w, `{"reachable": false, "ip": "%s"}`, ip)
		}
	}

	err = p.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ping failed: %v", err), http.StatusInternalServerError)
	}
}
