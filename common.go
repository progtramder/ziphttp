package ziphttp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
)

func StartTcpServer(addr string, serveHandler func(net.Conn)) error {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return err
	}

	for {
		conn, _ := l.Accept()
		if conn != nil {
			go serveHandler(conn)
		}
	}

	l.Close()
	return nil
}

func ReadInput() string {
	r := bufio.NewReader(os.Stdin)
	input, _, _ := r.ReadLine()
	return string(input)
}

func CmdLineLoop(prompt string, f func(string) int) {
	for {
		fmt.Println(prompt)
		input := ReadInput()
		if f(string(input)) == 0 {
			break
		}
	}
}

func GetMacId() string {

	var mac string
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		mac = inter.HardwareAddr.String()
		if mac != "" && mac != "00:00:00:00:00:00" {
			break
		}
	}

	return mac
}

func interfaceToString(in interface{}) string {
	var input string
	if s, ok := in.(string); ok {
		input = s
	} else if e, ok := in.(error); ok {
		input = e.Error()
	} else {
		panic("Not supported type.")
	}

	return input
}

func ColorRed(in interface{}) string {

	return "\033[31m" + interfaceToString(in) + "\033[0m"
}

func ColorBlue(in interface{}) string {
	return "\033[34m" + interfaceToString(in) + "\033[0m"
}

func ColorGreen(in interface{}) string {
	return "\033[32m" + interfaceToString(in) + "\033[0m"
}

type Channel struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

func NewChannel(c net.Conn) *Channel {
	return &Channel{conn: c, r: bufio.NewReader(c), w: bufio.NewWriter(c)}
}

func (this *Channel) Close() {
	this.conn.Close()
}

func (this *Channel) ReadByte() (byte, error) {
	return this.r.ReadByte()
}

func (this *Channel) UnReadByte() error {
	return this.r.UnreadByte()
}

func (this *Channel) ReadInt() (int32, error) {
	buf := make([]byte, 4)
	_, err := this.r.Read(buf)

	return int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3]), err
}

func (this *Channel) WriteByte(b byte) error {

	return this.w.WriteByte(b)
}

func (this *Channel) WriteInt(i int32) (err error) {

	err = this.w.WriteByte(byte((uint32(i) & 0xFF000000) >> 24))
	err = this.w.WriteByte(byte((i & 0x00FF0000) >> 16))
	err = this.w.WriteByte(byte((i & 0x0000FF00) >> 8))
	err = this.w.WriteByte(byte(i & 0x0000FF))

	return err
}

func (this *Channel) WriteString(s string) (err error) {

	err = this.WriteInt(int32((len(s))))
	_, err = this.w.Write([]byte(s))

	return err
}

func (this *Channel) ReadString() (string, error) {

	bytes, err := this.ReadInt()

	if err != nil {
		return "", err
	}

	buf := make([]byte, bytes)
	_, err = this.r.Read(buf)
	return string(buf), err
}

func (this *Channel) Write(s []byte) (int, error) {
	return this.w.Write(s)
}

func (this *Channel) Read(s []byte) (int, error) {

	return this.r.Read(s)
}

func (this *Channel) Flush() error {
	return this.w.Flush()
}

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{new(bytes.Buffer)}
}

func (this *Buffer) WriteInt(i int32) (err error) {

	err = this.WriteByte(byte((uint32(i) & 0xFF000000) >> 24))
	err = this.WriteByte(byte((i & 0x00FF0000) >> 16))
	err = this.WriteByte(byte((i & 0x0000FF00) >> 8))
	err = this.WriteByte(byte(i & 0x0000FF))
	return err
}
