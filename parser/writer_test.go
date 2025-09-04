package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriter(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		has := "hello"
		want := "+hello\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.SimpleString(has)
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)

	})

	t.Run("simple integer", func(t *testing.T) {
		has := 4
		want := ":4\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.Integer(has)
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)

	})

	t.Run("error string", func(t *testing.T) {
		has := fmt.Errorf("bad error")
		want := "-ERROR bad error\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.Error(has)
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)

	})

	t.Run("bulk string", func(t *testing.T) {
		has := "hello"
		want := "$5\r\nhello\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.BulkString(has)
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)

	})

	t.Run("null bulk string", func(t *testing.T) {
		want := "$-1\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.NullBulk()
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)
	})

	t.Run("array", func(t *testing.T) {
		has := []string{"a", "b", "c"}
		want := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.Array(has)
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)
	})

	t.Run("null array", func(t *testing.T) {
		want := "*-1\r\n"

		var buf bytes.Buffer
		br := bufio.NewWriter(&buf)
		w := NewWriter(br)

		err := w.NullArray()
		require.NoError(t, err)
		require.NoError(t, w.Flush())

		got := buf.String()
		assert.Equal(t, want, got)
	})
}
