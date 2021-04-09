package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"github.com/gorilla/mux"
	"log"
	"net/http"

	//"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"zk_protocol/lib"
)

//StartNewSocket ...
func StartNewSocket(station int,host string, port int, replyChan chan lib.Attendance) {
	zk, err := lib.MustConnect(host, port)
	defer func() {
		if v := recover(); v != nil {
			fmt.Println(v)
		}
		zk.Disconnect()
	}()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connect success", host, port)
	zk.SetTime()
	sn, err := zk.GetSerialNumber()
	if err != nil {
		panic(err)
	}
	zk.SerialNumber = strings.Split(string(sn.PayloadData), "=")[1]//cắt chuỗi
	err = zk.EnableRealtime()
	if err != nil {
		panic(err)
	}
	for {
		event, err := zk.RecieveEvent()
		if err != nil {
			panic(err)
		}
		fmt.Println(lib.Struct2String(event))
		if event.SessionCode == lib.EF_ATTLOG {
			data := event.PayloadData
			att, err := lib.ParseAttendance(host, data)
			att.SerialNumber = zk.SerialNumber
			att.StationID = station
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(lib.Struct2String(*att))
			replyChan <- *att //Gán

			fmt.Println("Running POST Api...")

		}
	}
}


func processEvent(url string,attChan chan lib.Attendance) {
	for {
		select {
		case reply := <-attChan:
			{
				dataform := &Attendance{
					StationID:reply.StationID,
					SerialNumber: reply.SerialNumber[0:len(reply.SerialNumber)-1],
					MachineIP:    reply.MachineIP,
					UserID:       reply.UserID,
					VerifyType:   reply.VerifyType,
					Status:       reply.Status,
					AttTime:      reply.AttTime,
				}
				b, err := json.Marshal(dataform) //Struct to Json: Json -> map[string]{} json.Unmarshal
				if err != nil {
					fmt.Println(">>>>>>>>>>>>>>")
					fmt.Println(err)
					return
				}
				fmt.Println(">>>>>>>>>>>>>>")
				fmt.Println(string(b))
				fmt.Println("________________")
				req, err := http.NewRequest("POST",url, bytes.NewBuffer(b))
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				//fmt.Println("response Status:", resp.Status)
				//fmt.Println("response Headers:", resp.Header)

				if err != nil{
					log.Fatal(err)
				}
			}
		}
	}

}

type Attendance struct {
	StationID    int  `json:"stationID"`
	SerialNumber string `json:"serialNumber"`
	MachineIP    string `json:"machineIp"`
	UserID       string `json:"userId"`
	VerifyType   int    `json:"verifyType"`
	Status       int    `json:"status"`
	AttTime      string `json:"atttime"`

}

func main() {


	replyChan := make(chan lib.Attendance)

	go StartNewSocket(7,"172.16.70.200", 4370, replyChan)
	go StartNewSocket(7,"172.16.70.177", 4370, replyChan)
	go StartNewSocket(5,"172.16.50.175", 4370, replyChan)
	// go StartNewSocket("172.16.80.234", 4370, replyChan)
	// go StartNewSocket("172.16.80.210", 4370, replyChan)
	// go StartNewSocket("172.16.60.110", 4370, replyChan)
	// go StartNewSocket("172.16.50.171", 4370, replyChan)
	// go StartNewSocket("172.16.30.45", 4370, replyChan)
	//go StartNewSocket("172.16.70.251", 14370, replyChan)
	go processEvent("http://restapi3.apiary.io/notes",replyChan)
	//Create server
	//router := mux.NewRouter()
	//
	//router.HandleFunc("/api/v1/keeptime")



	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("exit")

	fmt.Println("connect ok")

}
