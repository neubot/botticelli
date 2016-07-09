// vim: ts=4:sw=4

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

const kv_test_mid int = 1
const kv_test_c2s int = 2
const kv_test_s2c int = 4
const kv_test_sfw int = 8
const kv_test_status int = 16
const kv_test_meta int = 32

const kv_implemented_tests int = kv_test_s2c | kv_test_meta

/*
 __  __
|  \/  | ___  ___ ___  __ _  __ _  ___  ___
| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \/ __|
| |  | |  __/\__ \__ \ (_| | (_| |  __/\__ \
|_|  |_|\___||___/___/\__,_|\__, |\___||___/
                            |___/

	Message serialization and deserialization.
*/

func read_message(reader io.Reader) (byte, []byte, error) {

	// 1. read type

	type_buff := make([]byte, 1)
	_, err := io.ReadFull(reader, type_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_type := type_buff[0]
	log.Printf("ndt: message type: %d", msg_type)

	// 2. read length

	len_buff := make([]byte, 2)
	_, err = io.ReadFull(reader, len_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_length := binary.BigEndian.Uint16(len_buff)
	log.Printf("ndt: message length: %d", msg_length)

	// 3. read body

	msg_body := make([]byte, msg_length)
	_, err = io.ReadFull(reader, msg_body)
	if err != nil {
		return 0, nil, err
	}
	log.Printf("ndt: message body: '%s'\n", msg_body)

	return msg_type, msg_body, nil
}

type standard_message_t struct {
	Msg string `json:"msg"`
}

func read_standard_message(reader io.Reader) (byte, string, error) {
	msg_type, msg_buff, err := read_message(reader)
	if err != nil {
		return 0, "", err
	}
	s_msg := &standard_message_t{}
	err = json.Unmarshal(msg_buff, &s_msg)
	if err != nil {
		return 0, "", err
	}
	return msg_type, s_msg.Msg, nil
}

func write_message_internal(writer *bufio.Writer, message_type byte,
	encoded_body []byte) error {

	log.Printf("ndt: write any message: type=%d\n", message_type)
	log.Printf("ndt: write any message: length=%d\n", len(encoded_body))
	log.Printf("ndt: write any message: body='%s'\n", string(encoded_body))

	// 1. write type

	err := writer.WriteByte(message_type)
	if err != nil {
		return err
	}

	// 2. write length
	// TODO: make sure endianness conversion is performed correctly

	if len(encoded_body) > 65535 {
		return errors.New("ndt: encoded_body is too long")
	}
	encoded_len := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded_len, uint16(len(encoded_body)))
	_, err = writer.Write(encoded_len)
	if err != nil {
		return err
	}

	// 3. write message body

	_, err = writer.Write(encoded_body)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func write_standard_message(writer *bufio.Writer, message_type byte,
	message_body string) error {
	s_msg := &standard_message_t{
		Msg: message_body,
	}
	log.Printf("ndt: sending standard message: type=%d", message_type)
	log.Printf("ndt: sending standard message: body='%s'", message_body)
	data, err := json.Marshal(s_msg)
	if err != nil {
		return err
	}
	return write_message_internal(writer, message_type, data)
}

type extended_login_message_t struct {
	Msg      string `json:"msg"`
	TestsStr string `json:"tests"`
	Tests    int
}

func read_extended_login(reader io.Reader) (*extended_login_message_t, error) {

	// Read ordinary message

	msg_type, msg_buff, err := read_message(reader)
	if err != nil {
		return nil, err
	}
	if msg_type != 11 {
		return nil, errors.New("ndt: received invalid message")
	}

	// Process input as JSON message and valida its fields

	el_msg := &extended_login_message_t{}
	err = json.Unmarshal(msg_buff, &el_msg)
	if err != nil {
		return nil, err
	}
	log.Printf("ndt: client version: %s", el_msg.Msg)
	log.Printf("ndt: test suite: %s", el_msg.TestsStr)
	el_msg.Tests, err = strconv.Atoi(el_msg.TestsStr)
	if err != nil {
		return nil, err
	}
	log.Printf("ndt: test suite as int: %d", el_msg.Tests)
	if (el_msg.Tests & kv_test_status) == 0 {
		return nil, errors.New("ndt: client does not support TEST_STATUS")
	}

	return el_msg, nil
}

func write_raw_string(writer *bufio.Writer, str string) error {
	log.Printf("ndt: write raw string: '%s'", str)
	_, err := writer.WriteString(str)
	if err != nil {
		return err
	}
	return writer.Flush()
}

/*
 ____ ____   ____
/ ___|___ \ / ___|
\___ \ __) | |
 ___) / __/| |___
|____/_____|\____|

*/

type s2c_message_t struct {
	ThroughputValue  string
	UnsentDataAmount string
	TotalSentByte    string
}

func run_s2c_test(reader *bufio.Reader, writer *bufio.Writer) error {

	// Bind port and tell the port number to the server
	// TODO: choose a random port instead than an hardcoded port

	listener, err := net.Listen("tcp", ":3010")
	if err != nil {
		return err
	}
	err = write_standard_message(writer, 3, "3010")
	if err != nil {
		return err
	}
	defer listener.Close()

	// Wait for client to connect and setup all variables

	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	conn_writer := bufio.NewWriter(conn)
	defer conn.Close()
	output_buff := make([]byte, 8192)
	for i := 0; i < len(output_buff); i += 1 {
		// XXX seed the rng
		// XXX fill the buffer
		output_buff[i] = 'A'
	}

	// Send empty TEST_START message to tell the client to start

	err = write_standard_message(writer, 4, "")
	if err != nil {
		return err
	}

	// Send the buffer to the client for about ten seconds
	// TODO: here we should take `web100` snapshots
	// TODO: this could be refactored as a goroutine

	start := time.Now()
	bytes_sent := int64(0)
	var elapsed time.Duration
	for {
		_, err = conn_writer.Write(output_buff)
		if err != nil {
			log.Println("ndt: failed to write to client")
			break
		}
		err = conn_writer.Flush()
		if err != nil {
			log.Println("ndt: cannot flush connection with client")
			break
		}
		bytes_sent += int64(len(output_buff))
		elapsed = time.Since(start)
		if elapsed.Seconds() > 10.0 {
			log.Println("ndt: enough time elapsed")
			break
		}
	}
	conn.Close() // Explicit to notify the client we're done

	// Send message containing what we measured

	speed_kbits := (8.0 * float64(bytes_sent)) / 1000.0 / elapsed.Seconds()
	message := &s2c_message_t{
		ThroughputValue:  strconv.FormatFloat(speed_kbits, 'f', -1, 64),
		UnsentDataAmount: "0", // XXX
		TotalSentByte:    strconv.FormatInt(bytes_sent, 10),
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = write_message_internal(writer, 5, data)
	if err != nil {
		return err
	}

	// Receive message from client containing its measured speed

	msg_type, msg_body, err := read_standard_message(reader)
	if err != nil {
		return err
	}
	if msg_type != 5 {
		return errors.New("ndt: received unexpected message from client")
	}
	log.Printf("ndt: client measured speed: %s", msg_body)

	// FIXME: here we should send the web100 variables

	// Send the TEST_FINALIZE message that concludes the test

	return write_standard_message(writer, 6, "")
}

// XXX: what about timeouts?

func run_meta_test(reader *bufio.Reader, writer *bufio.Writer) error {

	// Send empty TEST_PREPARE and TEST_START messages to the client

	err := write_standard_message(writer, 3, "")
	if err != nil {
		return err
	}
	err = write_standard_message(writer, 4, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client

	for {
		msg_type, msg_body, err := read_standard_message(reader)
		if err != nil {
			return err
		}
		if msg_type != 5 {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msg_body == "" {
			break
		}
		log.Println("ndt: metadata from client: %s", msg_body)
	}

	// Send empty TEST_FINALIZE to client

	return write_standard_message(writer, 6, "")
}

func handle_connection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read extended loging message

	login_msg, err := read_extended_login(reader)
	if err != nil {
		log.Println("ndt: cannot read extended login")
		return
	}

	// Write kickoff message

	err = write_raw_string(writer, "123456 654321")
	if err != nil {
		log.Println("ndt: cannot write kickoff message")
		return
	}

	// Write queue empty message
	// TODO: here we should implement queue management

	err = write_standard_message(writer, 1, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client

	err = write_standard_message(writer, 2, "v3.7.0 (botticelli/0.0.1)")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs

	status := login_msg.Tests
	status &= ^kv_test_status
	status &= kv_implemented_tests
	tests_message := ""
	if (status & kv_test_s2c) != 0 {
		tests_message += strconv.Itoa(kv_test_s2c)
		tests_message += " "
	}
	if (status & kv_test_meta) != 0 {
		tests_message += strconv.Itoa(kv_test_meta)
	}
	err = write_standard_message(writer, 2, tests_message)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	// Run tests

	if (status & kv_test_s2c) != 0 {
		err = run_s2c_test(reader, writer)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & kv_test_meta) != 0 {
		err = run_meta_test(reader, writer)
		if err != nil {
			log.Println("ndt: failure running meta test")
			return
		}
	}

	// FIXME: send MSG_RESULTS to client

	// Send empty MSG_LOGOUT to client

	err = write_standard_message(writer, 9, "")
	if err != nil {
		return
	}
}

func StartNdtServer(endpoint string) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("ndt: accept() failed")
			continue
		}
		defer conn.Close()
		go handle_connection(conn)
	}
}
