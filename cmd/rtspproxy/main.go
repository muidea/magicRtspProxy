package main

import (
	"flag"
	"log"
	"net/http"
	"supconcloud/rtspproxy/rtspserver"
)

var rtspServerURL = "192.168.11.124:8554"
var bindAddr = "0.0.0.0:8010"

func main() {
	flag.StringVar(&rtspServerURL, "RtspSvr", rtspServerURL, "rtsp stream server.")
	flag.StringVar(&bindAddr, "Address", bindAddr, "rtsp proxy listen address.")
	flag.Parse()

	if len(bindAddr) == 0 || len(rtspServerURL) == 0 {
		flag.PrintDefaults()
		return
	}

	rtspServer := rtspserver.NewRtspServer(rtspServerURL)

	http.HandleFunc("/", rtspServer.WebSocketEntry)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
