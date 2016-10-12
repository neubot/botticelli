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

const kv_comm_failure byte = 0
const kv_srv_queue byte = 1
const kv_msg_login byte = 2
const kv_test_prepare byte = 3
const kv_test_start byte = 4
const kv_test_msg byte = 5
const kv_test_finalize byte = 6
const kv_msg_error byte = 7
const kv_msg_results byte = 8
const kv_msg_logout byte = 9
const kv_msg_waiting byte = 10
const kv_msg_extended_login byte = 11

const kv_test_mid int = 1
const kv_test_c2s int = 2
const kv_test_s2c int = 4
const kv_test_sfw int = 8
const kv_test_status int = 16
const kv_test_meta int = 32
const kv_test_c2s_ext int = 64
const kv_test_s2c_ext int = 128

const kv_srv_queue_heartbeat string = "9990"
const kv_srv_queue_server_fault string = "9977"
const kv_srv_queue_server_busy string = "9987"
const kv_srv_queue_server_busy_60s string = "9999"

const kv_parallel_streams int = 4

/*
 __  __
|  \/  | ___  ___ ___  __ _  __ _  ___  ___
| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \/ __|
| |  | |  __/\__ \__ \ (_| | (_| |  __/\__ \
|_|  |_|\___||___/___/\__,_|\__, |\___||___/
                            |___/

	Message serialization and deserialization.
*/

func read_message_internal(cc net.Conn, reader io.Reader) (
                           byte, []byte, error) {

	// 1. read type

	type_buff := make([]byte, 1)
	_, err := DoReadFull(cc, reader, type_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_type := type_buff[0]
	log.Printf("ndt: message type: %d", msg_type)

	// 2. read length

	len_buff := make([]byte, 2)
	_, err = DoReadFull(cc, reader, len_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_length := binary.BigEndian.Uint16(len_buff)
	log.Printf("ndt: message length: %d", msg_length)

	// 3. read body

	msg_body := make([]byte, msg_length)
	_, err = DoReadFull(cc, reader, msg_body)
	if err != nil {
		return 0, nil, err
	}
	log.Printf("ndt: message body: '%s'\n", msg_body)

	return msg_type, msg_body, nil
}

type standard_message_t struct {
	Msg string `json:"msg"`
}

func read_standard_message(cc net.Conn, reader io.Reader) (
                           byte, string, error) {
	msg_type, msg_buff, err := read_message_internal(cc, reader)
	if err != nil {
		return 0, "", err
	}
	s_msg := &standard_message_t{}
	err = json.Unmarshal(msg_buff, &s_msg)
	if err != nil {
		return 0, "", err
	}
	if s_msg == nil {
		return 0, "", errors.New("ndt: received literal 'null'")
	}
	return msg_type, s_msg.Msg, nil
}

func write_message_internal(cc net.Conn, writer *bufio.Writer,
                            message_type byte, encoded_body []byte) error {

	log.Printf("ndt: write any message: type=%d\n", message_type)
	log.Printf("ndt: write any message: length=%d\n", len(encoded_body))
	log.Printf("ndt: write any message: body='%s'\n", string(encoded_body))

	// 1. write type

	err := DoWriteByte(cc, writer, message_type)
	if err != nil {
		return err
	}

	// 2. write length

	if len(encoded_body) > 65535 {
		return errors.New("ndt: encoded_body is too long")
	}
	encoded_len := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded_len, uint16(len(encoded_body)))
	_, err = DoWrite(cc, writer, encoded_len)
	if err != nil {
		return err
	}

	// 3. write message body

	_, err = DoWrite(cc, writer, encoded_body)
	if err != nil {
		return err
	}
	return DoFlush(cc, writer)
}

func write_standard_message(cc net.Conn, writer *bufio.Writer,
                            message_type byte, message_body string) error {

	s_msg := &standard_message_t{
		Msg: message_body,
	}
	log.Printf("ndt: sending standard message: type=%d", message_type)
	log.Printf("ndt: sending standard message: body='%s'", message_body)
	data, err := json.Marshal(s_msg)
	if err != nil {
		return err
	}
	return write_message_internal(cc, writer, message_type, data)
}

type extended_login_message_t struct {
	Msg      string `json:"msg"`
	TestsStr string `json:"tests"`
	Tests    int
}

func read_extended_login(cc net.Conn, reader io.Reader) (
                         *extended_login_message_t, error) {

	// Read ordinary message

	msg_type, msg_buff, err := read_message_internal(cc, reader)
	if err != nil {
		return nil, err
	}
	if msg_type != kv_msg_extended_login {
		return nil, errors.New("ndt: received invalid message")
	}

	// Process input as JSON message and validate its fields

	el_msg := &extended_login_message_t{}
	err = json.Unmarshal(msg_buff, &el_msg)
	if err != nil {
		return nil, err
	}
	if el_msg == nil {
		return nil, errors.New("ndt: received literal 'null'")
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

func write_raw_string(cc net.Conn, writer *bufio.Writer, str string) error {
	log.Printf("ndt: write raw string: '%s'", str)
	_, err := DoWriteString(cc, writer, str)
	if err != nil {
		return err
	}
	return DoFlush(cc, writer)
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

func run_s2c_test(cc net.Conn, reader *bufio.Reader, writer *bufio.Writer,
                  is_extended bool) error {

	// Bind port and tell the port number to the server
	// TODO: choose a random port instead than an hardcoded port

	listener, err := net.Listen("tcp", ":3010")
	if err != nil {
		return err
	}
	prepare_message := "3010"
	if is_extended {
		prepare_message += " 10000.0 1 500.0 0.0 "
		prepare_message += strconv.Itoa(kv_parallel_streams)
	}
	err = write_standard_message(cc, writer, kv_test_prepare, prepare_message)
	if err != nil {
		return err
	}
	defer listener.Close()

	// Wait for client(s) to connect

	nstreams := 1
	if is_extended {
		nstreams = kv_parallel_streams
	}

	conns := make([]net.Conn, nstreams)
	for idx := 0; idx < len(conns); idx += 1 {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		conns[idx] = conn
	}

	// Send empty TEST_START message to tell the client to start

	err = write_standard_message(cc, writer, kv_test_start, "")
	if err != nil {
		return err
	}

	// Run the N streams in parallel

	channel := make(chan int64)

	output_buff := RandAsciiRemainder(8192)
	start := time.Now()

	for idx := 0; idx < len(conns); idx += 1 {
		log.Printf("ndt: start stream with id %d\n", idx)
		go func(conn net.Conn) {
			// Send the buffer to the client for about ten seconds
			// TODO: here we should take `web100` snapshots

			conn_writer := bufio.NewWriter(conn)
			defer conn.Close()

			for {
				_, err = DoWrite(conn, conn_writer, output_buff)
				if err != nil {
					log.Println("ndt: failed to write to client")
					break
				}
				err = DoFlush(conn, conn_writer)
				if err != nil {
					log.Println("ndt: cannot flush connection with client")
					break
				}
				channel <- int64(len(output_buff))
				if time.Since(start).Seconds() > 10.0 {
					log.Println("ndt: enough time elapsed")
					break
				}
			}

			conn.Close()   // Explicit to notify the client we're done
			channel <- -1  // Tell the controller we're done
		}(conns[idx])
	}

	bytes_sent := int64(0)
	for num_complete := 0; num_complete < len(conns); {
		count := <-channel
		if count < 0 {
			log.Printf("ndt: a stream just terminated...");
			num_complete += 1
			continue
		}
		bytes_sent += count
	}
	elapsed := time.Since(start)

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
	err = write_message_internal(cc, writer, kv_test_msg, data)
	if err != nil {
		return err
	}

	// Receive message from client containing its measured speed

	msg_type, msg_body, err := read_standard_message(cc, reader)
	if err != nil {
		return err
	}
	if msg_type != kv_test_msg {
		return errors.New("ndt: received unexpected message from client")
	}
	log.Printf("ndt: client measured speed: %s", msg_body)

	// FIXME: here we should send the web100 variables

	// Send the TEST_FINALIZE message that concludes the test

	return write_standard_message(cc, writer, kv_test_finalize, "")
}

/*
 __  __ _____ _____  _
|  \/  | ____|_   _|/ \
| |\/| |  _|   | | / _ \
| |  | | |___  | |/ ___ \
|_|  |_|_____| |_/_/   \_\

*/

func run_meta_test(cc net.Conn, reader *bufio.Reader,
                   writer *bufio.Writer) error {

	// Send empty TEST_PREPARE and TEST_START messages to the client

	err := write_standard_message(cc, writer, kv_test_prepare, "")
	if err != nil {
		return err
	}
	err = write_standard_message(cc, writer, kv_test_start, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client

	for {
		msg_type, msg_body, err := read_standard_message(cc, reader)
		if err != nil {
			return err
		}
		if msg_type != kv_test_msg {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msg_body == "" {
			break
		}
		log.Printf("ndt: metadata from client: %s", msg_body)
	}

	// Send empty TEST_FINALIZE to client

	return write_standard_message(cc, writer, kv_test_finalize, "")
}

/*
 ____            _                  _
|  _ \ _ __ ___ | |_ ___   ___ ___ | |
| |_) | '__/ _ \| __/ _ \ / __/ _ \| |
|  __/| | | (_) | || (_) | (_| (_) | |
|_|   |_|  \___/ \__\___/ \___\___/|_|

*/

func update_queue_pos(cc net.Conn, reader *bufio.Reader, writer *bufio.Writer,
                      position int) error {
	err := write_standard_message(cc, writer, kv_srv_queue,
			strconv.Itoa(position))
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE message")
	}
	err = write_standard_message(cc, writer, kv_srv_queue,
			kv_srv_queue_heartbeat)
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE heartbeat message")
	}
	msg_type, _, err := read_standard_message(cc, reader)
	if err != nil {
		return errors.New("ndt: cannot read MSG_WAITING message")
	}
	if msg_type != kv_msg_waiting {
		return errors.New("ndt: received unexpected message from client")
	}
	return nil
}

var kv_test_pending bool = false

func handle_connection(cc net.Conn) {
	defer cc.Close()

	reader := bufio.NewReader(cc)
	writer := bufio.NewWriter(cc)

	// Read extended login message

	login_msg, err := read_extended_login(cc, reader)
	if err != nil {
		log.Println("ndt: cannot read extended login")
		return
	}

	// Write kickoff message

	err = write_raw_string(cc, writer, "123456 654321")
	if err != nil {
		log.Println("ndt: cannot write kickoff message")
		return
	}

	// Queue management
	// XXX The current implementation of queue management is minimal, and
	// possibly also very ugly and stupid. Must be improved.

	for kv_test_pending {
		err = update_queue_pos(cc, reader, writer, 1)
		if err != nil {
			log.Println("ndt: failed to update client of its queue position")
			return
		}
		time.Sleep(3.0 * time.Second)
	}
	log.Println("ndt: this test is now running")
	kv_test_pending = true
	defer func() {
		log.Println("ndt: test complete; allowing another test to run")
		kv_test_pending = false
	}()

	// Write queue empty message

	err = write_standard_message(cc, writer, kv_srv_queue, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client

	err = write_standard_message(cc, writer, kv_msg_login,
			"v3.7.0 (" + Product + ")")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs

	status := login_msg.Tests
	tests_message := ""
	if (status & kv_test_s2c_ext) != 0 {
		tests_message += strconv.Itoa(kv_test_s2c_ext)
		tests_message += " "
	}
	if (status & kv_test_s2c) != 0 {
		tests_message += strconv.Itoa(kv_test_s2c)
		tests_message += " "
	}
	if (status & kv_test_meta) != 0 {
		tests_message += strconv.Itoa(kv_test_meta)
	}
	err = write_standard_message(cc, writer, kv_msg_login, tests_message)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	// Run tests

	if (status & kv_test_s2c_ext) != 0 {
		err = run_s2c_test(cc, reader, writer, true)
		if err != nil {
			log.Println("ndt: failure to run s2c_ext test")
			return
		}
	}
	if (status & kv_test_s2c) != 0 {
		err = run_s2c_test(cc, reader, writer, false)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & kv_test_meta) != 0 {
		err = run_meta_test(cc, reader, writer)
		if err != nil {
			log.Println("ndt: failure running meta test")
			return
		}
	}

	// FIXME: send MSG_RESULTS to client

	// Send empty MSG_LOGOUT to client

	err = write_standard_message(cc, writer, kv_msg_logout, "")
	if err != nil {
		return
	}
}

/*
 ____        _     _ _           _    ____ ___
|  _ \ _   _| |__ | (_) ___     / \  |  _ \_ _|
| |_) | | | | '_ \| | |/ __|   / _ \ | |_) | |
|  __/| |_| | |_) | | | (__   / ___ \|  __/| |
|_|    \__,_|_.__/|_|_|\___| /_/   \_\_|  |___|

*/

func StartNdtServer(endpoint string) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatal(err)
	}
	for {
		cc, err := listener.Accept()
		if err != nil {
			log.Println("ndt: accept() failed")
			continue
		}
		go handle_connection(cc)
	}
}
