package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	http.HandleFunc("/home", echo)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe("127.0.0.1:8010", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}
func echo(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))

	responseHeader := http.Header{}
	responseHeader.Add("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))
	conn, _, _, err := ws.UpgradeHTTP(r, w, responseHeader)
	if err != nil {
		// handle error
	}

	log.Printf("new collect, address:%s", r.RemoteAddr)

	go func() {
		defer conn.Close()

		// ffmpeg -rtsp_transport tcp -i rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast -f mpegts -codec:v mpeg1video -an -
		// ffmpeg -rtsp_transport tcp -i 'rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast' -an -f mpegts -codec:v mpeg1video -
		ffmpegCmd := exec.Command("ffmpeg", "-rtsp_transport tcp -i rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast -f mpegts -codec:v mpeg1video -an -")
		outStream, err := ffmpegCmd.StdoutPipe()
		if err != nil {
			log.Printf("Get stdout pipe failed, err:%s", err.Error())
			return
		}

		err = ffmpegCmd.Start()
		if err != nil {
			log.Printf("ffmpegCmd.Start failed, err:%s", err.Error())
			return
		}

		for {
			//msg, op, err := wsutil.ReadClientData(conn)
			//if err != nil {
			//	log.Printf("ReadClientData failed, err:%s", err.Error())
			//	break
			//}

			grepBytes, err := ioutil.ReadAll(outStream)
			if err != nil {
				log.Printf("ioutil.ReadAll failed, err:%s", err.Error())
				break
			}

			log.Printf("read data size:%d", len(grepBytes))

			err = wsutil.WriteServerMessage(conn, ws.OpBinary, grepBytes)
			if err != nil {
				log.Printf("WriteServerMessage failed, err:%s", err.Error())
				break
			}
		}
	}()
}
