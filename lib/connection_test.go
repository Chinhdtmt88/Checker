package lib

import (
	"fmt"
	"testing"
)

func TestGetTime(t *testing.T) {
	t.Run("test make comkey", func(t *testing.T) {
		zk, err := MustConnect("172.16.70.177", 4370)
		defer func() {
			if v := recover(); v != nil {
				fmt.Println(v)
			}
			zk.Disconnect()
		}()
		if err != nil {
			panic(err)
		}
		time, err := zk.GetTime()
		fmt.Println(time)
		reply, err := zk.SetTime()
		fmt.Println(Struct2String(reply))
		time, err = zk.GetTime()
		fmt.Println(time)
	})
}
