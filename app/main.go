package main

import (
	"bufio"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/commands"
	"github.com/codecrafters-io/redis-starter-go/parser"
	"github.com/codecrafters-io/redis-starter-go/store"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	db := store.NewStore()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn, db)
	}
}

func handleConn(conn net.Conn, ds *store.DataStore) {
	defer conn.Close()

	bw := bufio.NewWriter(conn)
	writer := parser.NewWriter(bw)

	br := bufio.NewReader(conn)
	reader := parser.NewReader(br)

	handler := commands.NewCommandHandler(ds, reader, writer)

	for {
		prefix, err := br.ReadByte()
		if err != nil {
			_ = writer.Error(fmt.Errorf("-Error: %s", err.Error()))
			return
		}

		switch prefix {

		case '*':
			cmd, err := reader.ArrayString()
			if err != nil {
				_ = writer.Error(fmt.Errorf("protocol error: %v", err))
				_ = writer.Flush()
				return
			}

			if err = handler.Handle(cmd); err != nil {
				_ = writer.Error(err)
				_ = writer.Flush()
				return
			}

			if err := writer.Flush(); err != nil {
				return
			}

		default:
			_ = writer.Error(fmt.Errorf("-Error: unknown command"))
			_ = writer.Flush()
			return
		}
	}
}
