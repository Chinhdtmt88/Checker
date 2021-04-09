package lib

/* func TestCreatePacket(t *testing.T) {
	t.Run("test create packet", func(t *testing.T) {
		packet, err := CreatePacket(2, 3, 4, nil)
		if err != nil {
			t.Error(err)
		}
		checksum := Checksum(packet[8:])
		if checksum != 0 {
			t.Error("bad checksum")
		}
		reply, err := ParseReply(packet)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(Struct2String(*reply))
	})
} */
