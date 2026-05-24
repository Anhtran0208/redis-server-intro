package core

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const CRLF string = "\r\n"

var RespNil = []byte("$-1\r\n")

// encode strng
func encodeString(s string) []byte {
	return []byte(fmt.Sprintf("$%d%s%s%s", len(s), CRLF, s, CRLF))
}

// encode array of string
func encodeStringArray(arr []string) []byte {
	var b []byte
	buff := bytes.NewBuffer(b)
	for _, element := range arr {
		buff.Write(encodeString(element))
	}
	return []byte(fmt.Sprintf("*%d%s%s", len(arr), CRLF, buff.Bytes()))
}

// raw data => RESP format data
func Encode(value interface{}, isSimpleString bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimpleString {
			return []byte(fmt.Sprintf("+%s%s", v, CRLF))
		}
		return encodeString(v)

	case int64, int32, int16, int8, int:
		return []byte(fmt.Sprintf(":%d\r\n", v))

	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v))

	case []string:
		return encodeStringArray(v)

	case [][]string:
		var b []byte
		buff := bytes.NewBuffer(b)
		for _, arr := range value.([][]string) {
			buff.Write(encodeStringArray(arr))
		}
		return []byte(fmt.Sprintf("*%d%s%s", len(value.([][]string)), CRLF, buff.Bytes()))

	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, x := range value.([]interface{}) {
			buf.Write(Encode(x, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(value.([]interface{})), buf.Bytes()))
	default:
		return RespNil
	}
}

// from RESP -> raw data, len(string), error
func readSimpleString(data []byte) (string, int, error) {
	curr_idx := 1
	for data[curr_idx] != '\r' {
		curr_idx += 1
	}
	return string(data[1:curr_idx]), curr_idx + 2, nil
}

func readInt64(data []byte) (int64, int, error) {
	curr_idx := 1
	var curr_val int64 = 0
	var sign int64 = 1
	if data[curr_idx] == '-' {
		sign = -1
		curr_idx++
	}
	if data[curr_idx] == '+' {
		curr_idx++
	}
	for data[curr_idx] != '\r' {
		curr_val = curr_val*10 + int64(data[curr_idx]-'0')
		curr_idx++
	}
	return sign * curr_val, curr_idx + 2, nil
}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// $5\r\nhello\r\n => 5, 4
func readLen(data []byte) (int, int) {
	res, pos, _ := readInt64(data)
	return int(res), pos
}

// $5\r\nhello\r\n => "hello"
func readBulkString(data []byte) (string, int, error) {
	length, pos := readLen(data)
	return string(data[pos:(pos + length)]), pos + length + 2, nil
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n => {"hello", "world"}
func readArray(data []byte) (interface{}, int, error) {
	length, pos := readLen(data)
	var res []interface{} = make([]interface{}, length)

	for i := range res {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		res[i] = elem
		pos += delta
	}
	return res, pos, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case ':':
		return readInt64(data)
	case '-':
		return readError(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, nil
}

// RESP format data => raw data
func Decode(data []byte) (interface{}, error) {
	res, _, err := DecodeOne(data)
	return res, err
}

func ParseCmd(data []byte) (*Command, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	array := value.([]interface{})
	tokens := make([]string, len(array))
	for i := range tokens {
		tokens[i] = array[i].(string)
	}
	res := &Command{Cmd: strings.ToUpper(tokens[0]), Args: tokens[1:]}
	return res, nil
}
