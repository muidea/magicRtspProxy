package videostream

import (
	"io"
	"log"
	"sync"
)

// NewServer create videoServer
func NewServer() *VideoServer {
	return &VideoServer{videoStream: make(map[string]*VideoStream)}
}

// VideoServer video server
type VideoServer struct {
	videoStream map[string]*VideoStream
	routesLock  sync.RWMutex
}

// Subscribe subsribe video stream
func (s *VideoServer) Subscribe(streamURL string, writer io.Writer) error {
	log.Printf("subscribe stream, url:%s", streamURL)
	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	stream, ok := s.videoStream[streamURL]
	if !ok {
		stream = &VideoStream{}

		err := stream.Start(streamURL)
		if err != nil {
			log.Printf("stream.Start, err:%s", err.Error())
			return err
		}

		s.videoStream[streamURL] = stream
	}

	stream.BindObserver(writer)

	return nil
}

// Unsubscribe subscribe video stream
func (s *VideoServer) Unsubscribe(streamURL string, writer io.Writer) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	stream, ok := s.videoStream[streamURL]
	if ok {
		stream.UnBindObserver(writer)
	}

	if stream.IsIdel() {
		stream.Stop()

		delete(s.videoStream, streamURL)
	}
}
