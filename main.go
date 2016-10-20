package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/coderplay/ipvlan/ipvlan"
)

func main() {

	var (
		debug            bool
		address          string
	)

	flag.StringVar(&debug, "debug", false, "enable debugging")
	flag.StringVar(&address, "socket", "/run/docker/plugins/ipvlan.sock", "socket on which to listen")

	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	d := ipvlan.NewDriver()
	h := ipvlan.NewHandler(d)
	h.ServeUnix("root", "ipvlan")
}