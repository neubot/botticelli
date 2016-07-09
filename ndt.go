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

const kvTestMid int = 1
const kvTestC2s int = 2
const kvTestS2c int = 4
const kvTestSfw int = 8
const kvTestStatus int = 16
const kvTestMeta int = 32

const kvImplementedTests int = kvTestS2c | kvTestMeta;

func readMessage(reader io.Reader) (byte, []byte, error) {

	type_buff := make([]byte, 1)
	_, err := io.ReadFull(reader, type_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_type := type_buff[0]
	log.Printf("ndt: message type: %d", msg_type)

	len_buff := make([]byte, 2)
	_, err = io.ReadFull(reader, len_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_length := binary.BigEndian.Uint16(len_buff)
	log.Printf("ndt: message length: %d", msg_length)

	msg_body := make([]byte, msg_length)
	_, err = io.ReadFull(reader, msg_body)
	if err != nil {
		return 0, nil, err
	}
	log.Printf("ndt: message body: '%s'\n", msg_body)

	return msg_type, msg_body, nil
}

type standardMessage struct {
	Msg string `json:"msg"`
}

func readStandardMessage(reader io.Reader) (byte, string, error) {
	msg_type, msg_buff, err := readMessage(reader)
	if err != nil {
		return 0, "", err
	}
	s_msg := &standardMessage{}
	err = json.Unmarshal(msg_buff, &s_msg)
	if err != nil {
		return 0, "", err
	}
	return msg_type, s_msg.Msg, nil
}

func writeAnyMessage(writer *bufio.Writer, message_type byte,
		encoded_body []byte) (error) {
	log.Printf("ndt: write any message: type=%d\n", message_type)
	log.Printf("ndt: write any message: length=%d\n", len(encoded_body))
	log.Printf("ndt: write any message: body='%s'\n", string(encoded_body))
	if len(encoded_body) > 65535 {
		return errors.New("ndt: encoded_body is too long")
	}
	err := writer.WriteByte(message_type)
	if err != nil {
		return err
	}
	encoded_len := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded_len, uint16(len(encoded_body)))
	_, err = writer.Write(encoded_len)
	if err != nil {
		return err
	}
	_, err = writer.Write(encoded_body)
	if err != nil {
		return err
	}
	return writer.Flush();
}

func writeStandardMessage(writer *bufio.Writer, message_type byte,
		message_body string) (error) {
	s_msg := &standardMessage{
		Msg: message_body,
	}
	log.Printf("ndt: sending standard message: type=%d", message_type)
	log.Printf("ndt: sending standard message: body='%s'", message_body)
	data, err := json.Marshal(s_msg)
	if err != nil {
		return err
	}
	return writeAnyMessage(writer, message_type, data)
}

type extendedLoginMessage struct {
	Msg      string `json:"msg"`
	TestsStr string `json:"tests"`
	Tests    int
}

func readExtendedLogin(reader io.Reader) (*extendedLoginMessage, error) {
	msg_type, msg_buff, err := readMessage(reader)
	if err != nil {
		return nil, err
	}
	if msg_type != 11 {
		return nil, errors.New("ndt: received invalid message")
	}
	el_msg := &extendedLoginMessage{}
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
	if (el_msg.Tests & kvTestStatus) == 0 {
		return nil, errors.New("ndt: client does not support TEST_STATUS")
	}
	return el_msg, nil
}

func writeString(writer *bufio.Writer, str string) (error) {
	log.Printf("ndt: sending: '%s'", str)
	_, err := writer.WriteString(str)
	if err != nil {
		return err
	}
	return writer.Flush()
}

type s2c_message struct {
	ThroughputValue string
	UnsentDataAmount string
	TotalSentByte string
}

func run_s2c_test(reader *bufio.Reader, writer *bufio.Writer) (error) {

	// Bind port and tell the port number to the server
	// TODO: choose a random port instead than an hardcoded port
	listener, err := net.Listen("tcp", ":3010")
	if err != nil {
		return err
	}
	err = writeStandardMessage(writer, 3, "3010")
	if err != nil {
		return err
	}
	defer listener.Close()

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
	err = writeStandardMessage(writer, 4, "")
	if err != nil {
		return err
	}

	// Send the buffer to the client for about ten seconds
	// TODO: here we should take `web100` snapshots
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
	message := &s2c_message{
		ThroughputValue: strconv.FormatFloat(speed_kbits, 'f', -1, 64),
		UnsentDataAmount: "0", // XXX
		TotalSentByte: strconv.FormatInt(bytes_sent, 10),
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = writeAnyMessage(writer, 5, data)
	if err != nil {
		return err
	}

	// Receive message from client containing its measured speed
	msg_type, msg_body, err := readStandardMessage(reader)
	if err != nil {
		return err
	}
	if msg_type != 5 {
		return errors.New("ndt: received unexpected message from client")
	}
	log.Printf("ndt: client measured speed: %s", msg_body)

	// FIXME: here we should send the web100 variables

	// Send the TEST_FINALIZE message that concludes the test
	return writeStandardMessage(writer, 6, "")
}

// XXX: what about timeouts?

func run_meta_test(reader *bufio.Reader, writer *bufio.Writer) (error) {

	// Send empty TEST_PREPARE and TEST_START messages to the client
	err := writeStandardMessage(writer, 3, "")
	if err != nil {
		return err
	}
	err = writeStandardMessage(writer, 4, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client
	for {
		msg_type, msg_body, err := readStandardMessage(reader)
		if err != nil {
			return err
		}
		if msg_type != 5 {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msg_body == "" {
			break
		}
	}

	// Send empty TEST_FINALIZE to client
	return writeStandardMessage(writer, 6, "")
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read extended loging message
	login_msg, err := readExtendedLogin(reader)
	if err != nil {
		log.Println("ndt: cannot read extended login")
		return
	}

	// Write kickoff message
	err = writeString(writer, "123456 654321")
	if err != nil {
		log.Println("ndt: cannot write kickoff message")
		return
	}

	// Write queue empty message
	// TODO: here we should implement queue management
	err = writeStandardMessage(writer, 1, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client
	err = writeStandardMessage(writer, 2, "v3.7.0 (botticelli/0.0.1)")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs
	status := login_msg.Tests
	status &= ^kvTestStatus;
	status &= kvImplementedTests;
	tests_message := ""
	if (status & kvTestS2c) != 0 {
		tests_message += strconv.Itoa(kvTestS2c)
		tests_message += " "
	}
	if (status & kvTestMeta) != 0 {
		tests_message += strconv.Itoa(kvTestMeta)
	}
	err = writeStandardMessage(writer, 2, tests_message)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	if (status & kvTestS2c) != 0 {
		err = run_s2c_test(reader, writer)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & kvTestMeta) != 0 {
		err = run_meta_test(reader, writer)
		if err != nil {
			log.Println("ndt: failure running meta test")
			return
		}
	}

	// FIXME: send MSG_RESULTS to client

	// Send empty MSG_LOGOUT to client
	err = writeStandardMessage(writer, 9, "")
	if err != nil {
		return
	}

	conn.Close()
}

func StartNdtServer(endpoint string) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("ndt: accept() failed")
			continue
		}
		defer conn.Close()
		go handleConnection(conn)
	}
}
