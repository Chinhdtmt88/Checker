package lib

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

//CreatePacket ...
func CreatePacket(commandID, sessionID, replyNumber int, data []byte) ([]byte, error) {
	var payload []byte
	buf := make([]byte, 12)
	//fixed tag
	start, err := hex.DecodeString(StartPacket)
	if err != nil {
		return nil, err
	}
	payload = append(payload, start...)
	//payload size
	binary.LittleEndian.PutUint32(buf[0:4], uint32(0))
	// cmd code
	binary.LittleEndian.PutUint16(buf[4:6], uint16(commandID))
	//checksum
	binary.LittleEndian.PutUint16(buf[6:8], uint16(0))
	// session id
	binary.LittleEndian.PutUint16(buf[8:10], uint16(sessionID))
	//reply number
	binary.LittleEndian.PutUint16(buf[10:12], uint16(replyNumber))
	//append data
	payload = append(payload, buf...)
	if len(data) > 0 {
		payload = append(payload, data...)
	}
	// write size
	binary.LittleEndian.PutUint32(payload[4:8], uint32(len(buf)+len(data)-4))
	//write checksum
	binary.LittleEndian.PutUint16(payload[10:12], uint16(Checksum(payload[8:])))
	return payload, nil
}

// Reply ...
type Reply struct {
	SnCode string
	ReplyCode   int
	SessionCode int
	ReplyNumber int
	PayloadData []byte
}

// ParseReply ...
func ParseReply(packet []byte) (*Reply, error) {
	if fmt.Sprintf("%x", packet[0:4]) != StartPacket {
		return &Reply{}, errors.New("Bad Start Tag")
	}
	if !IsValidPayload(packet[8:]) {
		return &Reply{}, errors.New("Bad Cheksum")
	}
	// payloadSize := binary.LittleEndian.Uint32(packet[4:8])
	replyCode := binary.LittleEndian.Uint16(packet[8:10])
	sessionCode := binary.LittleEndian.Uint16(packet[12:14])
	replyNumber := binary.LittleEndian.Uint16(packet[14:16])
	payloadData := packet[16:]
	return &Reply{
		ReplyCode:   int(replyCode),
		SessionCode: int(sessionCode),
		PayloadData: payloadData,
		ReplyNumber: int(replyNumber),
	}, nil
}
