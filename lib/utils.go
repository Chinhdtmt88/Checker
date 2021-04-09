package lib

import (
	"bytes"
	"fmt"
	"reflect"
)

//Checksum calculate checksum for given data bytes
func Checksum(data []byte) int {
	chk32B := 0
	j := 1
	if len(data)%2 == 1 {
		data = append(data, 0x00)
	}
	for j < len(data) {
		num16B := int(data[j-1]) + (int(data[j]) << 8)
		chk32B = chk32B + num16B
		j += 2
	}

	chk32B = (chk32B & 0xffff) + ((chk32B & 0xffff0000) >> 16)
	return chk32B ^ 0xffff
}

func IsValidPayload(payload []byte) bool {
	if Checksum(payload) != 0 {
		return false
	}
	return true
}
func Struct2String(theStruct interface{}) string {
	reflectV := reflect.ValueOf(theStruct)
	structType := reflectV.Type()
	b := &bytes.Buffer{}
	b.WriteString("{")
	for i := 0; i < reflectV.NumField(); i++ {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(structType.Field(i).Name)
		b.WriteString(": ")
		fieldValue := reflectV.Field(i)
		fieldType := reflectV.Field(i).Type()
		fieldKind := reflectV.Field(i).Kind()
		if (fieldKind == reflect.Slice || fieldKind == reflect.Array) && fieldType.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(b, "%s", fieldValue)
		} else {
			fmt.Fprint(b, fieldValue)
		}
	}
	b.WriteString("}")
	return b.String()
}
//func Struct2json