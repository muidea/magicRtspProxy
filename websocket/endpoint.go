package websocket

import (
	"log"
	"net"
	"supconcloud/rtspproxy/videostream"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Closer endpoint closer
type Closer interface {
	Close(endpoint *Endpoint)
}

// Endpoint WebSocket endpoint
type Endpoint struct {
	remoteAddr  string
	videoURL    string
	conn        net.Conn
	closer      Closer
	videoServer *videostream.VideoServer
}

// NewEndpoint create Endpoint
func NewEndpoint(remoteAddr, videoURL string, conn net.Conn, closer Closer, videoServer *videostream.VideoServer) *Endpoint {
	return &Endpoint{remoteAddr: remoteAddr, videoURL: videoURL, conn: conn, closer: closer, videoServer: videoServer}
}

// RemoteAddr return remoteAddr
func (s *Endpoint) RemoteAddr() string {
	return s.remoteAddr
}

// Release release endpoint
func (s *Endpoint) Release() {
	s.close()
}

// Run run endpoint
func (s *Endpoint) Run() error {
	err := s.videoServer.Subscribe(s.videoURL, s)
	if err != nil {
		log.Printf("videoServer.Subscribe failed, err:%s", err.Error())
		s.close()
		return err
	}

	return nil
}

func (s *Endpoint) close() {
	s.videoServer.Unsubscribe(s.videoURL, s)

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	if s.closer != nil {
		s.closer.Close(s)
		s.closer = nil
	}
}

func (s *Endpoint) Read() ([]byte, error) {
	msg, _, err := wsutil.ReadClientData(s.conn)
	if err != nil {
		log.Printf("ReadClientData failed, err:%s", err.Error())
		go s.close()
	}

	return msg, err
}

func (s *Endpoint) Write(data []byte) (int, error) {
	err := wsutil.WriteServerMessage(s.conn, ws.OpBinary, data)
	if err != nil {
		log.Printf("WriteServerMessage failed, err:%s", err.Error())

		go s.close()
	}

	return len(data), nil
}
