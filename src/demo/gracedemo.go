package main

import (
	"fmt"
	"gracetcp"
	"net"
)

func main() {
	fmt.Println("test")
	server_conn, err := gracetcp.ListenTCP("127.0.0.1", 22222)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(server_conn.GetFd())

	for {
		client_conn, err := server_conn.Accept()
		fmt.Println("client_conn:", client_conn)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				fmt.Println("timeout break")
				break
			}
		}
		go HandleConnection(client_conn)
	}

	fmt.Println("main:end")

	// 退出for循环，等待处理都结束
	(*server_conn).Wait()
}

func HandleConnection(conn *net.Conn) {
	defer func() {
		fmt.Println("1")
		(*conn).Close()
	}()
	defer func() {
		fmt.Println("2")
		if r := recover(); r != nil {
			fmt.Println(r.(string))
		}
	}()
	buf := make([]byte, 1024)
	for {
		n, err := (*conn).Read(buf)
		if err != nil {
			break
		}
		if n == 0 {
			fmt.Println(err)
			break
		}
		fmt.Println(string(buf[:n]))
		(*conn).Write(buf[:n])
	}
}
