package cmdHandler

import (
	"fmt"
	"strings"
)

func Handle(args []string) (string, error) {
	cmd := args[0]
	switch cmd {
	case "PING":
		return "+PONG\r\n", nil
	case "ECHO":
		s := strings.Join(args[1:], " ")
		return fmt.Sprintf("+%s\r\n", s), nil
	default:
		return "", fmt.Errorf("-Error: Unknown command")
	}
}
