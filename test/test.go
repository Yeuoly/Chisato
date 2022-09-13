package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/Yeuoly/Chisato/chisato"
	"github.com/aceld/zinx/znet"
)

func main() {
	request := chisato.ChisatoRequestTesting{
		Testcase: []chisato.ChisatoTestcase{
			{
				Stdin:  "2,4",
				Stdout: "6",
			},
			{
				Stdin:  "11,11",
				Stdout: "22",
			},
			{
				Stdin:  "1111,22222",
				Stdout: "23333",
			},
		},
		Code: `
#get two numbers from stdin
tuple = input()
(a, b) = tuple
print(int(a) + int(b))
		`,
		Language: "python2",
	}

	conn, err := net.Dial("tcp", "localhost:7171")
	if err != nil {
		panic(err)
	}

	go recv(conn)

	text, _ := json.Marshal(request)
	dp := znet.NewDataPack()
	msg, _ := dp.Pack(znet.NewMsgPackage(chisato.MESSAGEID_TESTING, text))
	_, err = conn.Write(msg)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("sent")
	}

	select {}
}

func recv(conn net.Conn) {
	dp := znet.NewDataPack()

	for {
		fmt.Println("trying fetch message header")
		head_data := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(conn, head_data)
		if err != nil {
			panic("failed to read message header")
		}

		msg_head, err := dp.Unpack(head_data)
		if err != nil {
			panic("failed to unpack message header")
		}

		fmt.Printf("get message with id %d\n", msg_head.GetMsgID())

		if msg_head.GetDataLen() > 0 {
			msg := msg_head.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				panic("failed to read message data")
			}

			fmt.Println(string(msg.Data))
		}
		return
	}
}
