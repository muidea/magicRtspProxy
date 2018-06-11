package main

import (
	"log"
	"net/http"
	"supconcloud/rtspproxy/rtspserver"
)

func main() {
	rtspServer := rtspserver.NewRtspServer()

	http.HandleFunc("/home", home2)
	http.HandleFunc("/", rtspServer.WebSocketEntry)
	log.Fatal(http.ListenAndServe("0.0.0.0:8010", nil))
}

func home2(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}
