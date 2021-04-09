package lib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// ZKConnection ...
type ZKConnection struct {
	*net.TCPConn
	sessionID     int
	replyNumber   int
	SerialNumber  string
	passcode      int
	commandString string
	responseSize  int
}

// SendCommand ...
func (c *ZKConnection) SendCommand(commandID, sessionID, replyNumber int, data []byte) (Reply, error) {
	replyNumber = replyNumber + 1
	packet, err := CreatePacket(commandID, sessionID, replyNumber, data)
	if err != nil {
		return Reply{}, err
	}
	// fmt.Printf("%x \n", packet)
	//send
	_, err = c.TCPConn.Write(packet)
	if err != nil {
		return Reply{}, err
	}
	buf := make([]byte, 1024)
	//recv reply
	nBytes, err := c.TCPConn.Read(buf)
	if err != nil {
		return Reply{}, err
	}
	reply, err := ParseReply(buf[0:nBytes])
	if err != nil {
		return Reply{}, err
	}
	//c.sessionID = reply.SessionCode
	fmt.Println(Struct2String(*reply))
	return *reply, nil
}

// MustConnect ...
func MustConnect(host string, port int) (*ZKConnection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		return nil, err
	}

	zk := NewZKConnection(conn)
	reply, err := zk.SendCommand(CMD_CONNECT, 0, 0, nil)
	zk.sessionID = reply.SessionCode
	zk.replyNumber = reply.ReplyNumber

	if reply.ReplyCode != CMD_ACK_OK {
		reply, err = zk.SendCommand(CMD_AUTH, zk.sessionID, zk.replyNumber, zk.MakeComkey())
		zk.replyNumber = reply.ReplyNumber
		if reply.ReplyCode != CMD_ACK_OK {
			return zk, errors.New("Cannot authen client")
		}
	}

	if err != nil {
		return nil, err
	}
	return zk, nil
}

// Disconnect ...
func (c *ZKConnection) Disconnect() error {
	// fmt.Println("disconnect")
	_, err := c.SendCommand(CMD_EXIT, c.sessionID, c.replyNumber, nil)
	if err != nil {
		return err
	}
	// fmt.Println(Struct2String(reply))
	c.TCPConn.Close()
	return nil
}

// EnableRealtime ...
func (c *ZKConnection) EnableRealtime() error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(EF_ATTLOG))
	reply, err := c.SendCommand(CMD_REG_EVENT, c.sessionID, c.replyNumber, []byte{0xff, 0xff, 0x00, 0x0})
	if err != nil {
		return err
	}
	if reply.ReplyCode != CMD_ACK_OK {
		return errors.New("Device return not OK")
	}
	return nil
}

// AckOk ...
func (c *ZKConnection) AckOk() error {
	packet, err := CreatePacket(CMD_ACK_OK, c.sessionID, 0, nil)
	if err != nil {
		return err
	}
	// fmt.Printf("%x \n", packet)
	//send
	_, err = c.TCPConn.Write(packet)
	if err != nil {
		return err
	}
	// fmt.Println("send ack ok")
	return nil
}

//RecieveEvent ...
func (c *ZKConnection) RecieveEvent() (Reply, error) {
	var result []byte
	buf := make([]byte, 1024)
	fmt.Println("try receive event")
	numRecv, err := c.TCPConn.Read(buf)
	if err != nil {
		return Reply{}, err
	}

	result = append(result, buf...)
	totalSize := binary.LittleEndian.Uint32(buf[4:8]) + 8
	// loop to read data
	for {
		fmt.Println("in loop")
		if numRecv == int(totalSize) {
			break
		}
		nBytes, err := c.TCPConn.Read(buf)
		if err != nil {
			return Reply{}, err
		}
		result = append(result, buf...)
		numRecv += nBytes
	}
	// fmt.Println("done receive event")

	reply, err := ParseReply(buf[0:totalSize])
	if err != nil {
		return Reply{}, err
	}
	// send ACK
	c.AckOk()

	return *reply, nil
}

// GetTime ...
func (c *ZKConnection) GetTime() (string, error) {
	reply, err := c.SendCommand(CMD_GET_TIME, c.sessionID, c.replyNumber, nil)
	if err != nil {
		return "", err
	}
	if len(reply.PayloadData) != 4 {
		return "", errors.New("Bad time data")
	}

	timeNumber := binary.LittleEndian.Uint32(reply.PayloadData)
	seconds := int(timeNumber % 60)
	minutes := int((timeNumber / 60.) % 60)
	hour := int((timeNumber / (3600.)) % 24)
	day := int((timeNumber/(3600.*24.))%31) + 1
	month := int((timeNumber/(3600.*24.*31.))%12) + 1
	year := int((timeNumber/(3600.*24.))/365) + 2000
	return fmt.Sprintf("%.2d/%.2d/%.2d %.2d:%.2d:%.2d", year, month, day, hour, minutes, seconds), nil
}

// SetTime ...
func (c *ZKConnection) SetTime() (Reply, error) {
	t := time.Now()
	timeNumber := ((t.Year()%100)*12*31+((int(t.Month())-1)*31)+t.Day()-1)*(24*60*60) + (t.Hour()*60+t.Minute())*60 + t.Second()
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(timeNumber))
	reply, err := c.SendCommand(CMD_SET_TIME, c.sessionID, c.replyNumber, buf)
	c.replyNumber = reply.ReplyNumber
	if err != nil {
		return Reply{}, err
	}
	return reply, nil
}

//GetSerialnumber
func (c *ZKConnection) GetSerialNumber() (Reply, error) {
	buf := []byte("~SerialNumber")
	buf = append(buf, 0x00)                                                       //nối thêm mã hex
	reply, err := c.SendCommand(CMD_OPTIONS_RRQ, c.sessionID, c.replyNumber, buf) //(buf:=data)
	c.replyNumber = reply.ReplyNumber

	if err != nil {
		return Reply{}, err
	}
	//fmt.Printf("%x", reply.PayloadData)

	return reply, nil
}

// MakeComkey ...
func (c *ZKConnection) MakeComkey() []byte {
	key := c.passcode
	sessionID := c.sessionID
	k := 0
	for i := 0; i < 32; i++ {
		if (key & (1 << i)) == 1 {
			k = (k<<1 | 1)

		} else {
			k = k << 1
		}
	}
	k += sessionID
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(k)) //lấy mảng 4 Byte
	// fmt.Println("B", B)

	tempBuf := []byte{
		byte(int(buf[0]) ^ 90), //Z
		byte(int(buf[1]) ^ 75), //K
		byte(int(buf[2]) ^ 83), //S
		byte(int(buf[3]) ^ 79), //O
	}

	B := 0xff & 50
	commKey := []byte{
		byte(int(tempBuf[2]) ^ B),
		byte(int(tempBuf[3]) ^ B),
		byte(B),
		byte(int(tempBuf[1]) ^ B),
	}

	return commKey
}

// NewZKConnection ...
func NewZKConnection(tcpConn *net.TCPConn) *ZKConnection {
	return &ZKConnection{
		TCPConn:     tcpConn,
		sessionID:   0,
		replyNumber: 0,
		passcode:    0,
	}
}
