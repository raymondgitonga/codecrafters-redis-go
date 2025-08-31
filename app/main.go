package main

import (
	"bufio"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/cmdHandler"
	"github.com/codecrafters-io/redis-starter-go/parser"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	br := bufio.NewReader(conn)
	for {
		prefix, err := br.ReadByte()
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
		}

		switch prefix {
		case '*':
			cmd, err := parser.ArrayString(br)
			if err != nil {
				_, err := conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
				if err != nil {
					conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
				}
			}

			resp, err := cmdHandler.Handle(cmd)
			if err != nil {
				_, err := conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
				if err != nil {
					conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
				}
			}

			conn.Write([]byte(resp))
		default:
			_, err = conn.Write([]byte("-Error: unknown command"))
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("-Error: %s", err)))
			}
			return
		}
	}
}
