package common

// Make sure I/O operations have a deadline
// See <https://groups.google.com/forum/#!topic/golang-nuts/afgEYsoV8j0>

import (
	"bufio"
	"net"
	"io"
	"time"
)

const kv_deadline = 10.0 * time.Second

func IoReadFull(conn net.Conn, reader io.Reader, data []byte) (int, error) {
	count := 0
	err := conn.SetReadDeadline(time.Now().Add(kv_deadline))
	if err != nil {
		return count, err
	}
	count, err = io.ReadFull(reader, data)
	if err != nil {
		return count, err
	}
	err = conn.SetReadDeadline(time.Time{})
	return count, err
}

func IoWriteByte(conn net.Conn, writer *bufio.Writer, data byte) error {
	err := conn.SetWriteDeadline(time.Now().Add(kv_deadline))
	if err != nil {
		return err
	}
	err = writer.WriteByte(data)
	if err != nil {
		return err
	}
	return conn.SetWriteDeadline(time.Time{})
}

func IoWrite(conn net.Conn, writer io.Writer, data []byte) (int, error) {
	count := 0
	err := conn.SetWriteDeadline(time.Now().Add(kv_deadline))
	if err != nil {
		return count, err
	}
	count, err = writer.Write(data)
	if err != nil {
		return count, err
	}
	err = conn.SetWriteDeadline(time.Time{})
	return count, err
}

func IoFlush(conn net.Conn, writer *bufio.Writer) error {
	err := conn.SetWriteDeadline(time.Now().Add(kv_deadline))
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return conn.SetWriteDeadline(time.Time{})
}

func IoWriteString(conn net.Conn, writer *bufio.Writer, data string) (
                   int, error) {
	count := 0
	err := conn.SetWriteDeadline(time.Now().Add(kv_deadline))
	if err != nil {
		return count, err
	}
	count, err = writer.WriteString(data)
	if err != nil {
		return count, err
	}
	err = conn.SetWriteDeadline(time.Time{})
	return count, err
}
