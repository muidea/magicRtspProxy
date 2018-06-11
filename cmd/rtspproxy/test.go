package main

import (
	"log"
	"net"
	"net/http"
	"os/exec"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type transferData struct {
	conn net.Conn
}

func (s *transferData) Write(p []byte) (n int, err error) {
	err = wsutil.WriteServerMessage(s.conn, ws.OpBinary, p)
	if err != nil {
		log.Printf("WriteServerMessage failed, err:%s", err.Error())
	}

	return len(p), nil
}

func _main() {
	http.HandleFunc("/home", echo)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe("0.0.0.0:8010", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}

func echo(w http.ResponseWriter, r *http.Request) {
	responseHeader := http.Header{}
	responseHeader.Add("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))
	conn, _, _, err := ws.UpgradeHTTP(r, w, responseHeader)
	if err != nil {
		log.Printf("ws.UpgradeHTTP failed, err:%s", err.Error())
		return
	}

	log.Print(r.URL.RawQuery)
	log.Printf("new client, address:%s", r.RemoteAddr)

	trace := transferData{conn: conn}
	argList := []string{
		"-i",
		"rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast",
		"-f",
		"mpegts",
		"-codec:v",
		"mpeg1video",
		"-an",
		"-",
	}
	ffmpegCmd := exec.Command("ffmpeg", argList...)
	ffmpegCmd.Stdout = &trace

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				ffmpegCmd.Process.Kill()
			}
		}()
		defer conn.Close()

		err = ffmpegCmd.Start()
		if err != nil {
			log.Printf("ffmpegCmd.Start failed, err:%s", err.Error())
			return
		}

		ffmpegCmd.Wait()
		if err != nil {
			log.Printf("ffmpegCmd.Wait failed, err:%s", err.Error())
			return
		}

		log.Printf("client disconnect...")
	}()
}
