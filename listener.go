package stoppablenetlistener

// Stoppable listener based on hydrogen18/stoppableListener
import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hydrogen18/stoppableListener"
)

type StoppableNetListener struct {
	*stoppableListener.StoppableListener
	stop         chan struct{}
	finishedStop chan struct{}
	Timeout      time.Duration
}

func New(port int) (snl *StoppableNetListener, err error) {
	var tcpL net.Listener
	var sl *stoppableListener.StoppableListener

	if port < 1 {
		err = errors.New(fmt.Sprintf("Cannot use port: %v", port))
		return
	}

	if tcpL, err = net.Listen("tcp", fmt.Sprintf(":%v", port)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// This statement *cannot* fail
	sl, _ = stoppableListener.New(tcpL)

	snl = &StoppableNetListener{}
	snl.stop = make(chan struct{})
	snl.finishedStop = make(chan struct{})
	snl.StoppableListener = sl
	snl.Timeout = 1 * time.Second
	return
}

func (snl *StoppableNetListener) Accept() (net.Conn, error) {
	var stop bool = false

	for {
		//Wait up to one second for a new connection
		snl.StoppableListener.SetDeadline(time.Now().Add(snl.Timeout))

		newConn, err := snl.TCPListener.Accept()

		//Check for the channel being closed
		select {
		case <-snl.stop:
			stop = true
		default:
			//If the channel is still open, continue as normal
		}

		if stop {
			break
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}
		return newConn, err
	}
	close(snl.finishedStop)
	return nil, stoppableListener.StoppedError
}

func (snl *StoppableNetListener) Stop() {
	close(snl.stop)
	// Wait for Accept loop to terminate
	_, _ = <-snl.finishedStop
}
