package rtspserver

import (
	"log"
	"net/http"
	"supconcloud/rtspproxy/videostream"
	"supconcloud/rtspproxy/websocket"
	"sync"

	"github.com/gobwas/ws"
)

// NewRtspServer create new rtspServer
func NewRtspServer() *RtspServer {
	return &RtspServer{videoServer: videostream.NewServer(), endpoint: make(map[string]*websocket.Endpoint)}
}

// RtspServer rtsp server
type RtspServer struct {
	videoServer *videostream.VideoServer
	endpoint    map[string]*websocket.Endpoint
	routesLock  sync.RWMutex
}

// Close endpoint close notify
func (s *RtspServer) Close(endpoint *websocket.Endpoint) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	delete(s.endpoint, endpoint.RemoteAddr())
}

// WebSocketEntry websocket entry
func (s *RtspServer) WebSocketEntry(w http.ResponseWriter, r *http.Request) {
	responseHeader := http.Header{}
	responseHeader.Add("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))
	conn, _, _, err := ws.UpgradeHTTP(r, w, responseHeader)
	if err != nil {
		log.Printf("ws.UpgradeHTTP failed, err:%s", err.Error())
		return
	}

	videoURL := "rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast"

	log.Printf("new client, address:%s", r.RemoteAddr)
	endpoint := websocket.NewEndpoint(r.RemoteAddr, videoURL, conn, s, s.videoServer)

	err = endpoint.Run()
	if err != nil {
		log.Printf("endpoint.Run failed, err:%s", err.Error())
		return
	}

	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	s.endpoint[endpoint.RemoteAddr()] = endpoint
}
