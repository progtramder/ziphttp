package ziphttp

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const host = "192.168.1.3:6397"
const hostAuth = "192.168.1.3:6399"
const magicNum = 0xef

var mux http.ServeMux

type ZipConn struct {
	net.Conn
}

func (c *ZipConn) magic(b []byte) {
	for i, _ := range b {
		b[i] ^= magicNum
	}
}

func (c *ZipConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	c.magic(b[0:n])
	return n, err
}

func (c *ZipConn) Write(b []byte) (n int, err error) {

	c.magic(b)
	return c.Conn.Write(b)
}

type listener struct {
	net.Listener
}

func (l *listener) Accept() (net.Conn, error) {
	tl := l.Listener.(*net.TCPListener)
	c, err := tl.AcceptTCP()
	if err != nil {
		return nil, err
	}

	c.SetKeepAlive(true)
	c.SetKeepAlivePeriod(3 * time.Minute)

	return &ZipConn{c}, nil
}

func ZipListener(inner net.Listener) net.Listener {
	l := &listener{Listener: inner}
	return l
}

type fileHandler struct {
	root string
}

func FileServer(dir string) http.Handler {
	return &fileHandler{dir}
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	macid := r.Header["User-Agent"][0]
	usrInfo, _ := GetUserByMacId(macid)
	root := f.root
	path := r.URL.Path
	if path == "/" {
		var filterName string
		for _, v := range usrInfo.Codes {
			if macid == v.MacId {
				filterName = v.Filter
				break
			}
		}

		var filter []string = nil
		if filterName != defaultFilter {
			filter = usrInfo.Filters[filterName]
		}
		Render(w, usrInfo.User, filter)
		return
	}

	if !strings.Contains(path, startPage) {
		root += "/" + usrInfo.User
	}

	http.FileServer(http.Dir(root)).ServeHTTP(w, r)
}

func ListenAndServe(addr string, handler http.Handler) error {
	if handler == nil {
		handler = &mux
	}
	server := &http.Server{Addr: addr, Handler: handler}

	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return err
	}

	return server.Serve(ZipListener(ln))
}

func Handle(pattern string, handler http.Handler) {
	mux.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.HandleFunc(pattern, handler)
}

func StartProxy(addr, macid string) error {

	return StartTcpServer(addr, func(c net.Conn) {

		config := &tls.Config{}
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0] = rootCert
		//Set the connection as TLS connection
		tlsConn := tls.Server(c, config)
		defer tlsConn.Close()

		//Read request from client
		request, err := ReadRequest(tlsConn)
		if err != nil {
			return
		}

		agent := request.Header["User-Agent"]
		if agent == nil || agent[0] != GUID {
			return
		}

		if strings.Contains(request.URL.Path, startPage) {
			errRecover := RecoverPacketJson()
			errRemove := RemoveMainJs()
			if errRecover != nil || errRemove != nil {
				return
			}
		}

		request.Header["User-Agent"] = []string{macid}

		//Connect to remote host
		con, err := net.DialTimeout("tcp", host, time.Second*10)
		if err != nil {
			log.Println("Fail to connect to remote server.")
			return
		}

		zc := &ZipConn{con}
		defer zc.Close()

		//Send request to remote host
		response, err := DoRequest(request, zc)
		if err != nil {
			return
		}

		//Respond to client
		response.Write(tlsConn)
	})
}

func ReadRequest(c net.Conn) (*http.Request, error) {

	r := bufio.NewReader(c)
	c.SetReadDeadline(time.Now().Add(time.Second * 10))
	return http.ReadRequest(r)
}

//Send request to server and get a response
func DoRequest(r *http.Request, c net.Conn) (*http.Response, error) {

	//Transfer request to remote server
	err := r.Write(c)

	if err != nil {
		return nil, err
	}

	//Request has been sent to server, get the response
	reader := bufio.NewReader(c)
	return http.ReadResponse(reader, r)
}
