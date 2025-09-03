package parser

import (
	"bufio"
	"fmt"
)

type Writer struct {
	bw *bufio.Writer
}

func NewWriter(bw *bufio.Writer) *Writer {
	return &Writer{
		bw: bw,
	}
}

func (w *Writer) Flush() error { return w.bw.Flush() }

func (w *Writer) SimpleString(s string) error {
	resp := fmt.Sprintf("+%s\r\n", s)
	_, err := w.bw.Write([]byte(resp))
	return err
}

func (w *Writer) Integer(n int) error {
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := w.bw.Write([]byte(resp))
	return err
}

func (w *Writer) Error(error error) error {
	_, err := fmt.Fprintf(w.bw, "-ERROR %s\r\n", error.Error())
	return err
}

// BulkString string is divided into three parts:
// Length ie $d\r\n{length}
// The byte response
// CRF
func (w *Writer) BulkString(b string) error {
	// write length
	_, err := fmt.Fprintf(w.bw, "$%d\r\n", len(b))
	if err != nil {
		return err
	}
	// write response
	if _, err = w.bw.Write([]byte(b)); err != nil {
		return err
	}
	// add CRF
	_, err = w.bw.WriteString("\r\n")
	return err
}

func (w *Writer) Array(arr []string) error {
	// write length
	_, err := fmt.Fprintf(w.bw, "*%d\r\n", len(arr))
	if err != nil {
		return err
	}

	for _, a := range arr {

		_, err = fmt.Fprintf(w.bw, "$%d\r\n", len(a))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w.bw, "%s\r\n", a)
	}
	return err
}

func (w *Writer) NullBulk() error {
	_, err := w.bw.WriteString("$-1\r\n")
	return err
}

func (w *Writer) EmptyString() error {
	_, err := w.bw.WriteString("*0\r\n")
	return err
}
