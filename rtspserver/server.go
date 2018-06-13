package rtspserver

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"supconcloud/rtspproxy/videostream"
	"supconcloud/rtspproxy/websocket"
	"sync"

	"github.com/gobwas/ws"
)

const (
	realTime = 1
	playBack = 2
)

type streamFilter struct {
	streamType int
	serverURL  string
	channelNo  int
	startTime  string
	endTime    string
}

func (s *streamFilter) Parse(request *http.Request) error {
	str := request.URL.Query().Get("streamType")
	streamType, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("Parse failed, illegal streamType,streamType:%s", str)
		return err
	}
	s.streamType = streamType

	s.serverURL = request.URL.Query().Get("serverUrl")
	str = request.URL.Query().Get("channelNo")
	channelNo, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("Parse failed, illegal channelNo,channelNo:%s", str)
		return err
	}
	s.channelNo = channelNo

	s.startTime = request.URL.Query().Get("startTime")
	s.endTime = request.URL.Query().Get("endTime")

	validFlag := false
	switch s.streamType {
	case realTime:
		validFlag = s.verifyRealTime()
	case playBack:
		validFlag = s.verifyPlayBack()
	}

	if !validFlag {
		log.Printf("illegal query url, url:%s", request.URL.RawQuery)
		return errors.New("illegal query url")
	}

	return nil
}

func (s *streamFilter) FormatURL(serverURL string) string {
	if s.streamType == realTime {
		return s.FormatRealTime(serverURL)
	}

	return s.FormatPlayBack(serverURL)
}

func (s *streamFilter) verifyRealTime() bool {
	return len(s.serverURL) > 0 && s.channelNo >= 0
}

func (s *streamFilter) verifyPlayBack() bool {
	return len(s.serverURL) > 0 && s.channelNo >= 0 && len(s.startTime) > 0 && len(s.endTime) > 0
}

func (s *streamFilter) FormatRealTime(serverURL string) string {
	return fmt.Sprintf("rtsp://%s/%s/%d/preview?transportmode=unicast", serverURL, s.serverURL, s.channelNo)
}

func (s *streamFilter) FormatPlayBack(serverURL string) string {
	return fmt.Sprintf("rtsp://%s/%s/%d/playback?starttime=%s&endtime=%s", serverURL, s.serverURL, s.channelNo, s.startTime, s.endTime)
}

// NewRtspServer create new rtspServer
func NewRtspServer(serverURL string) *RtspServer {
	return &RtspServer{rtspServer: serverURL, videoServer: videostream.NewServer(), endpoint: make(map[string]*websocket.Endpoint)}
}

// RtspServer rtsp server
type RtspServer struct {
	rtspServer  string
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

	streamFilter := streamFilter{}
	err = streamFilter.Parse(r)
	if err != nil {
		log.Printf("streamFilter.Parse failed, err:%s", err.Error())
		return
	}

	videoURL := streamFilter.FormatURL(s.rtspServer)
	//videoURL := "rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast"
	//videoURL := "rtsp://192.168.11.124:8554/192.168.10.254/1/preview?transportmode=unicast"

	log.Printf("new client, address:%s, queryUrl:%s", r.RemoteAddr, r.URL.RawQuery)
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
