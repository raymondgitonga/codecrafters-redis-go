package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/parser"
	"strings"
)

type Store interface {
	Get(key string) (string, error)
	Set(key string, data []string) error
	LRange(key string, args []string) ([]string, error)
	RPush(key string, data []string) (*int, error)
	LPush(key string, data []string) (*int, error)
	LLen(key string) (int, error)
	LPop(key string) (string, error)
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
	case "GET":
		value, err := h.Store.Get(args[1])
		if err != nil {
			return h.Writer.NullBulk()
		}
		return h.Writer.BulkString(value)
	case "SET":
		if len(args) < 3 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		value := args[2:]
		err := h.Store.Set(key, value)
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

		size, err := h.Store.RPush(key, values)
		if err != nil {
			return err
		}

		if size == nil {
			return h.Writer.Error(fmt.Errorf("-Error: invalid list size returned"))
		}

		return h.Writer.Integer(*size)
	case "LPUSH":
		if len(args) < 3 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		values := args[2:]

		size, err := h.Store.LPush(key, values)
		if err != nil {
			return err
		}

		if size == nil {
			return h.Writer.Error(fmt.Errorf("-Error: invalid list size returned"))
		}

		return h.Writer.Integer(*size)
	case "LRANGE":
		if len(args) < 4 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		resp, err := h.Store.LRange(key, args[2:])
		if err != nil {
			return err
		}

		if len(resp) < 1 {
			return h.Writer.EmptyString()
		}

		return h.Writer.Array(resp)
	case "LLEN":
		if len(args) < 2 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		resp, err := h.Store.LLen(key)
		if err != nil {
			return err
		}

		return h.Writer.Integer(resp)
	case "LPOP":
		if len(args) < 2 {
			return fmt.Errorf("not enough arguments")
		}

		key := args[1]
		resp, err := h.Store.LPop(key)
		if err != nil {
			return err
		}

		return h.Writer.SimpleString(resp)
	default:
		return h.Writer.Error(fmt.Errorf("-Error: Unknown command"))
	}
}
