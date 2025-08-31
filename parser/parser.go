package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

func ArrayString(br *bufio.Reader) ([]string, error) {
	nStr, err := readLine(br)
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(nStr)
	if err != nil || n < 0 {
		return nil, err
	}

	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		b, err := br.ReadByte()
		if err != nil {
			return nil, err
		}

		if b != '$' {
			return nil, fmt.Errorf("expected $ but got %c", b)
		}

		s, err := parseBulkString(br)
		if err != nil {
			return nil, err
		}
		args = append(args, s)
	}
	return args, nil
}

func parseBulkString(br *bufio.Reader) (string, error) {
	lenStr, err := readLine(br)
	if err != nil {
		return "", err
	}

	n, err := strconv.Atoi(lenStr)
	if err != nil {
		return "", err
	}

	buf := make([]byte, n+2) // include \r\n -> dont leave them in the buffer
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func readLine(s *bufio.Reader) (string, error) {
	b, err := s.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	if len(b) > 0 && b[len(b)-1] == '\r' {
		b = b[:len(b)-1]
	}

	return string(b), nil
}
