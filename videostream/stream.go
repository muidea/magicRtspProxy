package videostream

import (
	"io"
	"log"
	"os/exec"
)

const (
	bindObserver = iota
	unBindObserver
	dispatchData
	terminateStream
	checkStatus
)

type action struct {
	command int
	data    interface{}
	reply   interface{}
}

type actionChannel chan *action

// VideoStream videoStream
type VideoStream struct {
	ffmpegCmd    *exec.Cmd
	observerList []io.Writer

	runningFlag   bool
	actionChannel actionChannel
}

func (s *VideoStream) runInternel() {
	s.runningFlag = true

	for s.runningFlag {
		action, ok := <-s.actionChannel
		if ok {
			switch action.command {
			case bindObserver:
				s.bindObserverInternal(action.data.(io.Writer))
			case unBindObserver:
				s.unBindObserverInternal(action.data.(io.Writer))
				if action.reply != nil {
					action.reply.(chan bool) <- true
				}
			case dispatchData:
				s.writerInternal(action.data.([]byte))
			case checkStatus:
				if action.reply != nil {
					action.reply.(chan bool) <- (len(s.observerList) == 0)
				}
			case terminateStream:
				s.runningFlag = false
				if action.reply != nil {
					action.reply.(chan bool) <- true
				}
				break
			}
		}
	}
}

// Start start video stream
// "rtsp://admin:chenjiping1314@192.168.10.254:554/Streaming/Channels/101?transportmode=unicast"
func (s *VideoStream) Start(streamURL string) error {
	log.Printf("start capture video stream, url:%s", streamURL)

	s.actionChannel = make(actionChannel)

	argList := []string{
		"-i",
		streamURL,
		"-f",
		"mpegts",
		"-codec:v",
		"mpeg1video",
		"-an",
		"-",
	}
	s.ffmpegCmd = exec.Command("ffmpeg", argList...)
	s.ffmpegCmd.Stdout = s

	err := s.ffmpegCmd.Start()
	if err != nil {
		log.Printf("ffmpegCmd.Start failed, err:%s", err.Error())
		return err
	}

	go s.runInternel()

	return err
}

// Stop stop video stream
func (s *VideoStream) Stop() {
	log.Print("stop capture video stream")

	reply := make(chan bool)
	action := &action{command: terminateStream, data: nil, reply: reply}
	s.actionChannel <- action
	<-reply

	err := s.ffmpegCmd.Process.Kill()
	if err != nil {
		log.Printf("ffmpegCmd.Process.Kill failed, err:%s", err.Error())
	}

	s.ffmpegCmd.Wait()
}

// IsDone videostream is finish
func (s *VideoStream) IsDone() bool {
	return s.ffmpegCmd.ProcessState.Exited()
}

// IsIdel observerList is empty
func (s *VideoStream) IsIdel() bool {
	reply := make(chan bool)
	action := &action{command: checkStatus, data: nil, reply: reply}
	s.actionChannel <- action
	val := <-reply

	return val
}

// Write io.Writer
func (s *VideoStream) Write(data []byte) (int, error) {
	action := &action{command: dispatchData, data: data, reply: nil}
	s.actionChannel <- action

	return len(data), nil
}

func (s *VideoStream) writerInternal(data []byte) {
	for _, val := range s.observerList {
		val.Write(data)
	}
}

// BindObserver bind observer
func (s *VideoStream) BindObserver(writer io.Writer) {
	action := &action{command: bindObserver, data: writer, reply: nil}
	s.actionChannel <- action
}

func (s *VideoStream) bindObserverInternal(writer io.Writer) {
	log.Printf("bindObserverInternal....")
	s.observerList = append(s.observerList, writer)
}

// UnBindObserver unbind observer
func (s *VideoStream) UnBindObserver(writer io.Writer) {
	reply := make(chan bool)
	action := &action{command: unBindObserver, data: writer, reply: reply}
	s.actionChannel <- action

	<-reply
}

func (s *VideoStream) unBindObserverInternal(writer io.Writer) {
	log.Printf("unBindObserverInternal....")
	newList := []io.Writer{}
	for _, val := range s.observerList {
		if val != writer {
			newList = append(newList, val)
		}
	}

	s.observerList = newList
}
