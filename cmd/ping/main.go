package main

import (
	"fmt"
	"io"
	"net"

	"github.com/Yeuoly/Chisato/chisato"
	"github.com/aceld/zinx/znet"
)

/*
	Author: Yeuoly

	Used to ping the server
*/

func main() {
	conn, err := net.Dial("tcp", "localhost:7172")
	if err != nil {
		panic(err)
	}

	close_chan := make(chan int, 1)

	go recv(conn, close_chan)

	dp := znet.NewDataPack()
	msg, _ := dp.Pack(znet.NewMsgPackage(chisato.MESSAGEID_PING, []byte{}))
	_, err = conn.Write(msg)
	if err != nil {
		panic(err)
	}

	<-close_chan
	conn.Close()
}

func recv(conn net.Conn, close_chan chan int) {
	dp := znet.NewDataPack()

	head_data := make([]byte, dp.GetHeadLen())
	_, err := io.ReadFull(conn, head_data)
	if err != nil {
		return
	}

	msg_head, err := dp.Unpack(head_data)
	if err != nil {
		panic("failed to unpack message header")
	}

	if msg_head.GetDataLen() > 0 {
		msg := msg_head.(*znet.Message)
		msg.Data = make([]byte, msg.GetDataLen())

		_, err := io.ReadFull(conn, msg.Data)
		if err != nil {
			panic("failed to read message data")
		}

		fmt.Println(string(msg.Data))
	}

	close_chan <- 1
}
