package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/parser"
	"strings"
)

type Store interface {
	SetString(key string, data []string) error
	SetList(key string, data []string) (*int, error)
	Get(key string) (string, error)
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
		if len(args) < 3 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		value := args[2:]
		err := h.Store.SetString(key, value)
		if err != nil {
			return err
		}

		return h.Writer.SimpleString("OK")
	case "RPUSH":
		if len(args) < 3 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		values := args[2:]
		size, err := h.Store.SetList(key, values)
		if err != nil {
			return err
		}

		if size == nil {
			return h.Writer.Error(fmt.Errorf("-Error: invalid list size returned"))
		}

		return h.Writer.Integer(*size)
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
