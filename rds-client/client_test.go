package rds_client_test

import (
	"bufio"
	"bytes"
	rds_client "changeme/rds-client"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

type command struct {
	input  []interface{}
	output interface{}
}

func TestConnect(t *testing.T) {
	tt := []struct {
		name     string
		commands []command
	}{
		{
			name: "execute a set and a get",
			commands: []command{
				//{
				//	input: []interface{}{
				//		"SET",
				//		"some-key",
				//		"Maurício",
				//	},
				//	output: "OK",
				//},
				{
					input: []interface{}{
						"GET",
						"test",
					},
					output: "Maurício",
				},
			},
		},
		//{
		//	name: "execute a get",
		//	commands: []command{
		//		{
		//			input: []interface{}{
		//				"GET",
		//				"some-other-key",
		//			},
		//			output: nil,
		//		},
		//	},
		//},
		//{
		//	name: "execute a set and get with UTF characters",
		//	commands: []command{
		//		{
		//			input: []interface{}{
		//				"SET",
		//				"対馬",
		//				"Tsushima",
		//			},
		//			output: "OK",
		//		},
		//		{
		//			input: []interface{}{
		//				"GET",
		//				"対馬",
		//			},
		//			output: "Tsushima",
		//		},
		//	},
		//},
	}

	for _, ts := range tt {
		t.Run(ts.name, func(t *testing.T) {
			client, err := rds_client.Connect(context.Background(), "localhost:6379", "1111")
			require.NoError(t, err)

			defer client.Close()

			for _, c := range ts.commands {
				result, err := client.Send(c.input)
				require.NoError(t, err)
				fmt.Println(result.Content())

				assert.Equal(t, c.output, result.Content())
			}

		})
	}
}
func TestConnectClient(t *testing.T) {
	//rds_client.Connection(context.Background(), "127.0.0.1:6379", "1111")
	rContext, err := rds_client.Connection("localhost:6379", "")
	if err != nil {
		return
	}
	defer rds_client.Close(rContext)
	input := []interface{}{
		"GET",
		"test",
	}
	result, _ := rds_client.SendCommand(rContext, input)
	fmt.Println(result.Content())

}

func TestConnectClient2(t *testing.T) {
	//rds_client.Connection(context.Background(), "127.0.0.1:6379", "1111")
	rContext, err := rds_client.Connection("localhost:6379", "")
	if err != nil {
		return
	}
	nameList := rContext.GetNodeNameList()

	fmt.Println(nameList)

}

func TestRedisKeyScan(t *testing.T) {
	//rds_client.Connection(context.Background(), "127.0.0.1:6379", "1111")
	rds_client.Connection("127.0.0.1:6379", "")

}

//AUTH
func TestConnect3(t *testing.T) {
	dialer := net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 10,
	}

	conn, err := dialer.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("conn err : ", err)
		return
	}

	_, err = conn.Write([]byte("*2\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}
	_, err = conn.Write([]byte("$4\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}
	_, err = conn.Write([]byte("AUTH\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}

	_, err = conn.Write([]byte("$4\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}

	_, err = conn.Write([]byte("1111\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}

	receive := make([]byte, 4096)
	n, err := conn.Read(receive)
	if err != nil {
		fmt.Println("read err : ", err)
		return
	}

	data := receive[:n]
	fmt.Println("read data : ", string(data))

	_, err = conn.Write([]byte("*1\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}
	rCommand := "command"
	lengths := int64(len(rCommand))
	_, err = conn.Write([]byte("$" + strconv.FormatInt(lengths, 10) + "\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}
	_, err = conn.Write([]byte(rCommand + "\r\n"))
	if err != nil {
		fmt.Println("write err : ", err)
		return
	}

	for {
		receive1 := make([]byte, 4096)
		n1, err2 := conn.Read(receive1)
		if err2 != nil {
			if err2 != io.EOF {
				fmt.Println("read Error err : ", err)
			}
			break
		}
		data1 := receive1[:n1]
		fmt.Println(string(data1))

		if n1 < len(receive1) {
			fmt.Println("마지막")
			break
		}
	}

}

func TestStringsda(t *testing.T) {
	input := "foo end bar"
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(split)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Printf("Error: %s\n", scanner.Err())
	}
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	if err == nil && token != nil && bytes.Equal(token, []byte{'e', 'n', 'd'}) {
		return 0, []byte{'E', 'N', 'D'}, bufio.ErrFinalToken
	}
	return
}
