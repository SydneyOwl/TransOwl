package netutil

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/gookit/slog"
	"github.com/sydneyowl/TransOwl/internal/cfg"
	"io"
	"net"
	"time"
)

type TCPModule struct {
	conn     *net.Conn
	listener *net.TCPListener
}

func NewSenderTCPModule(ip net.IP, interfaced NetInterface) (*TCPModule, error) {
	myAddr := &net.TCPAddr{
		IP:   interfaced.CurrentIP,
		Port: cfg.TCP_PORT,
	}
	recAddr := &net.TCPAddr{
		IP:   ip,
		Port: cfg.TCP_PORT,
	}
	dial := net.Dialer{
		Timeout:   cfg.MAX_TCP_DIAL_TIMEOUT * time.Second,
		LocalAddr: myAddr,
	}
	conn, err := dial.Dial("tcp", recAddr.String())
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(cfg.MAX_TCP_READ_TIMEOUT * time.Second))
	//conn.SetWriteDeadline(time.Now().Add(cfg.MAX_TCP_WRITE_TIMEOUT * time.Second))
	return &TCPModule{conn: &conn}, nil
}

func NewReceiverTCPModule(interfaced NetInterface) (*TCPModule, error) {
	myAddr := &net.TCPAddr{
		IP:   interfaced.CurrentIP,
		Port: cfg.TCP_PORT,
	}
	listen, err := net.ListenTCP("tcp", myAddr)
	if err != nil {
		return nil, err
	}
	return &TCPModule{listener: listen, conn: nil}, nil
}
func (sm *TCPModule) BlockTillSenderRecv() ([]byte, error) {
	buf := make([]byte, cfg.FILE_SLICE_SIZE)
	(*(sm.conn)).SetDeadline(time.Now().Add(cfg.MAX_TCP_READ_TIMEOUT * time.Second))
	n, err := (*(sm.conn)).Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
func (sm *TCPModule) BlockTillAcceptWithTimeout(timeout time.Duration) (*net.Conn, error) {
	err := sm.listener.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}
	conn, err := sm.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &conn, nil
}
func (sm *TCPModule) ShutConn() {
	_ = (*(sm.conn)).Close()
}
func (sm *TCPModule) ShutListener() {
	_ = sm.listener.Close()
}
func (sm *TCPModule) SendData(file []byte) error {
	customBinary, err := GenerateCustomBinary(file)
	if err != nil {
		return err
	}
	_, err = (*(sm.conn)).Write(customBinary)
	return err
}

// This function aims to solve tcp sticky packet!
// We add length in front of each packet
func GenerateCustomBinary(data []byte) ([]byte, error) {
	headerBytes := make([]byte, cfg.HEAD_SIZE)
	length := uint16(len(data))
	pkg := new(bytes.Buffer)
	binary.BigEndian.PutUint16(headerBytes, length)
	err := binary.Write(pkg, binary.BigEndian, headerBytes)
	if err != nil {
		return nil, err
	}
	err = binary.Write(pkg, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

func DecodeCustomBinary(reader *bufio.Reader) ([]byte, error) {
	count := 0
	for {
		if count >= 100 {
			//20 KB
			// Force quit
			return nil, ERR_BUFFER_SIZE_TOO_SMALL
		}
		lengthByte, err := reader.Peek(cfg.HEAD_SIZE)
		if err != nil {
			if errors.Is(io.EOF, err) {
				continue
			}
			return nil, err
		}
		lengthBuff := bytes.NewReader(lengthByte)
		var length uint16
		err = binary.Read(lengthBuff, binary.BigEndian, &length)
		if err != nil {
			return nil, err
		}
		if length > 65535 || length < 0 {
			slog.Errorf("Invalid length!")
			return nil, ERR_INVALID_LENGTH
		}
		if int32(reader.Buffered()) < int32(length+cfg.HEAD_SIZE) {
			count += 1
			continue
		}
		pack := make([]byte, cfg.HEAD_SIZE+length)
		_, err = reader.Read(pack)
		if err != nil {
			return nil, err
		}
		return pack[cfg.HEAD_SIZE:], nil
	}
}
func DecodeCustomBinaryO(data []byte) ([]byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	headerLength, _ := reader.Peek(cfg.HEAD_SIZE)
	lengthBuff := bytes.NewBuffer(headerLength)
	var length uint16
	err := binary.Read(lengthBuff, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	if int32(reader.Buffered()) < int32(length+cfg.HEAD_SIZE) {
		return nil, ERR_INVALID_LENGTH
	}
	pack := make([]byte, int(cfg.HEAD_SIZE+length))
	_, err = reader.Read(pack)
	if err != nil {
		return nil, err
	}
	return pack[cfg.HEAD_SIZE:], nil
}
