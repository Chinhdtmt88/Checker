package lib

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Attendance ...
type Attendance struct {
	StationID int `json:"stationID"`
	SerialNumber string `json:"serialNumber"`
	MachineIP  string `json:"machineIp"`
	UserID     string `json:"userId"`
	VerifyType int `json:"verifyType"`
	Status     int `json:"status"`
	AttTime    string `json:"atttime"`
}


func ParseAttendance(ip string, data []byte) (*Attendance, error) {

	fmt.Println(len(data))
	fmt.Printf("%x \n", data)
	if len(data) < 12 {
		return nil, errors.New("Not valid attendance packet")
	}
	// userID := string(data[0:9])
	// verifyType := binary.LittleEndian.Uint16(data[24:26])
	// attTime := string(data[26:32])

	if len(data) == 12 {
		fmt.Println("12")
		userID := binary.LittleEndian.Uint32(data[0:4])
		status := data[25]
		verifyType := data[26]
		attTimeBytes := data[26:32]
		attTime := fmt.Sprintf("%d/%d/%d %d:%d:%d",
			attTimeBytes[0],
			attTimeBytes[1],
			attTimeBytes[2],
			attTimeBytes[3],
			attTimeBytes[4],
			attTimeBytes[5],
		)
		return &Attendance{
			MachineIP:  ip,
			UserID:     fmt.Sprintf("%d", userID),
			VerifyType: int(verifyType),
			Status:     int(status),
			AttTime:    attTime,
		}, nil
	}

	if len(data) >= 32 {
		var userID []byte
		for _, b := range data[0:24] {
			if int(b) > 0 {
				userID = append(userID, b)
			}
		}
		status := data[24]
		verifyType := data[25]
		attTimeBytes := data[26:32]
		attTime := fmt.Sprintf("%d/%d/%d %d:%d:%d",
			attTimeBytes[0],
			attTimeBytes[1],
			attTimeBytes[2],
			attTimeBytes[3],
			attTimeBytes[4],
			attTimeBytes[5],
		)
		return &Attendance{
			MachineIP:  ip,
			UserID:     string(userID),
			VerifyType: int(status),
			Status:     int(verifyType),
			AttTime:    attTime,
		}, nil
	}

	return nil, errors.New("bad packet")
}
