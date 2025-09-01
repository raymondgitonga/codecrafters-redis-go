package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/parser"
	"strings"
)

type Store interface {
	Set(key string, data []byte, ex string) error
	Get(key string) ([]byte, error)
	Del(key string)
}

type Handler struct {
	Store  Store
	Reader *parser.Reader
	Writer *parser.Writer
}

func NewCommandHandler(store Store, r *parser.Reader, w *parser.Writer) *Handler {
	return &Handler{
		Store:  store,
		Reader: r,
		Writer: w,
	}
}
func (h *Handler) Handle(args []string) error {
	cmd := strings.ToUpper(args[0])
	switch cmd {
	case "PING":
		return h.Writer.SimpleString("PONG")
	case "ECHO":
		s := strings.Join(args[1:], " ")
		return h.Writer.SimpleString(s)
	case "SET":
		var expiry string
		if len(args) < 3 {
			return fmt.Errorf("not enough arguments")
		}

		if len(args) >= 4 {
			expiry = args[4]
		}
		err := h.Store.Set(args[1], []byte(args[2]), expiry)
		if err != nil {
			return err
		}
		return h.Writer.SimpleString("OK")
	case "GET":
		value, err := h.Store.Get(args[1])
		if err != nil {
			return h.Writer.NullBulk()
		}
		return h.Writer.Bulk(value)
	default:
		return h.Writer.Error(fmt.Errorf("-Error: Unknown command"))
	}
}
