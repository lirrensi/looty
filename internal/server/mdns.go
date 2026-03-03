package server

import (
	"log"

	"github.com/grandcat/zeroconf"
)

var mdnsServer *zeroconf.Server

// StartMDNS announces the looty service on the local network as "looty.local"
func StartMDNS(port int) error {
	var err error
	mdnsServer, err = zeroconf.Register(
		"looty",      // instance name
		"_http._tcp", // service type
		"local.",     // domain
		port,         // port
		nil,          // TXT records
		nil,          // interfaces (nil = all)
	)
	if err != nil {
		return err
	}
	log.Println("mDNS: Announcing as looty.local")
	return nil
}

// StopMDNS shuts down the mDNS server
func StopMDNS() {
	if mdnsServer != nil {
		mdnsServer.Shutdown()
		mdnsServer = nil
	}
}
