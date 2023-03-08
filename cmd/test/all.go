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
	type testcase struct {
		Code     string
		Language string
		Cases    []chisato.ChisatoTestcase
	}

	testcases := []testcase{
		{
			Code: `
#include <stdio.h>
int main() {
	int a, b;
	scanf("%d,%d", &a, &b);
	printf("%d", a + b);
	return 0;
}
			`,
			Language: "c",
			Cases:    nil,
		},
		//testcase for cpp
		{
			Code: `
#include <iostream>
using namespace std;
int main() {
	int a, b;
	cin >> a;
	getchar();
	cin >> b;
	cout << a + b;
	return 0;
}			`,
			Language: "cpp",
			Cases:    nil,
		},
		//testcase for python
		{
			Code: `
a, b = map(int, input().split(','))
print(a + b)
			`,
			Language: "python3",
			Cases:    nil,
		},
		{
			Code: `
a, b = map(int, raw_input().split(','))
print(a + b)
			`,
			Language: "python2",
			Cases:    nil,
		},
		//testcase for java
		{
			Code: `
package cn.srmxy.chisato.main;
public class Main {
	public static void main(String[] args) {
		java.util.Scanner scanner = new java.util.Scanner(System.in);
		String line = scanner.nextLine();
		String[] arr = line.split(",");
		int a = Integer.parseInt(arr[0]);
		int b = Integer.parseInt(arr[1]);
		System.out.println(a + b);
	}
}
					`,
			Language: "java",
			Cases:    nil,
		},
		//testcase for go
		{
			Code: `
package main
import "fmt"
func main() {
	var a, b int
	fmt.Scanf("%d,%d", &a, &b)
	fmt.Println(a + b)
}
			`,
			Language: "go",
			Cases:    nil,
		},
	}

	for i, v := range testcases {
		v := v
		i := i
		go func() {
			for {
				test(i, v.Code, v.Language)
			}
		}()
	}

	select {}
}

func test(id int, code string, language string) {
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
		Code:     code,
		Language: language,
	}

	conn, err := net.Dial("tcp", "localhost:7172")
	if err != nil {
		panic(err)
	}

	close_chan := make(chan int, 1)

	go recv(conn, close_chan)

	text, _ := json.Marshal(request)
	dp := znet.NewDataPack()
	msg, _ := dp.Pack(znet.NewMsgPackage(chisato.MESSAGEID_TESTING, text))
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
