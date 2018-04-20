package stoppablenetlistener

// Stoppable listener based on hydrogen18/stoppableListener
import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type StoppableNetListener struct {
	*net.TCPListener
	Timeout    time.Duration
	wg         sync.WaitGroup
	AcceptChan chan net.Conn
	stopped    bool
}

func New(port int) (snl *StoppableNetListener, err error) {
	var tcpL net.Listener

	if port < 1 {
		err = errors.New(fmt.Sprintf("Cannot use port: %v", port))
		return
	}

	if tcpL, err = net.Listen("tcp", fmt.Sprintf(":%v", port)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	snl = &StoppableNetListener{}
	snl.TCPListener = tcpL.(*net.TCPListener)
	snl.Timeout = 1 * time.Second
	snl.AcceptChan = make(chan net.Conn)
	go snl.listen()
	return
}

func (snl *StoppableNetListener) Accept() (net.Conn, error) {
	conn := <-snl.AcceptChan
	return conn, nil
}

func (snl *StoppableNetListener) AcceptOneConnection() (net.Conn, error) {
	var conn net.Conn
	var err error
	for !snl.stopped {
		//Wait up to one second for a new connection
		snl.TCPListener.SetDeadline(time.Now().Add(snl.Timeout))

		conn, err = snl.TCPListener.Accept()

		fmt.Printf("stopped=%v\n", snl.stopped)
		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for
			//new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
				fmt.Printf("timeout: %v\n", err)
				continue
			} else {
				break
			}
		}
		break
	}
	return conn, err
}

func (snl *StoppableNetListener) listen() {
	snl.wg.Add(1)
	defer snl.wg.Done()
	defer snl.TCPListener.Close()

	for !snl.stopped {
		if conn, err := snl.AcceptOneConnection(); err != nil {
			// We should probably handle this
		} else {
			snl.AcceptChan <- conn
		}
	}
}

func (snl *StoppableNetListener) Stop() {
	snl.stopped = true
	// Wait for Accept loop to terminate
	snl.wg.Wait()
}
