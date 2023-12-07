package tcpip

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/misterdelle/invt_logger_reader/ports"
)

const timeout = 20 * time.Second

type tcpIpPort struct {
	name string
	conn net.Conn
}

func New(portName string) ports.CommunicationPort {
	return &tcpIpPort{
		name: portName,
	}
}

func (s *tcpIpPort) Open() error {

	var err error

	d := net.Dialer{Timeout: 3 * time.Second}
	s.conn, err = d.Dial("tcp", s.name)

	if err != nil {
		return err
	}

	return nil
}

func (s *tcpIpPort) Close() error {

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}

func (s *tcpIpPort) Read(buf []byte) (int, error) {
	if s.conn == nil {
		return 0, fmt.Errorf("connection is not open")
	}

	reader := bufio.NewReader(s.conn)

	if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	return reader.Read(buf)
}

func (s *tcpIpPort) Write(payload []byte) (int, error) {
	if s.conn == nil {
		return 0, fmt.Errorf("connection is not open")
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	return s.conn.Write(payload)
}
