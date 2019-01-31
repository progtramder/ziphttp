package ziphttp

import (
	"net"
	"encoding/json"
	"errors"
	"log"
	"time"
	"math/rand"
)

const cmdFail    = 0
const cmdSuccess = 1
const cmdUnbind  = 2
const cmdAuth    = 3
const cmdLogin   = 4
const buffsize = 128

var ErrorAuthFail = errors.New("cmdFail")

type Auth struct{
	Cmd   int
	Code  string
	MacId string
}

func Login(code, macid string) error {
	au := &Auth{cmdLogin, code, macid}
	return cmdAuthenticate(au)
}

func Authenticate(code, macid string) error  {
	au := &Auth{cmdAuth, code, macid}
	return cmdAuthenticate(au)
}

func Unbind(code, macid string) error  {
	au := &Auth{cmdUnbind, code, macid}
	return cmdAuthenticate(au)
}

func StartAuthServer() error {

	return StartTcpServer(":6399", func (conn net.Conn) {
		defer conn.Close()
		au := Auth{}
		err := readAuthCmd(conn, &au)
		if err != nil {
			log.Println(err)
			return
		}

		var codeInfo *CodeInfo
		if au.Cmd == cmdAuth {
			codeInfo, err = GetCodeByCode(au.Code)
			if err == nil && codeInfo.MacId == "" {
				codeInfo.MacId = au.MacId
				au.Cmd = cmdSuccess
			}
		} else if au.Cmd == cmdLogin {
			codeInfo, err = GetCodeByMacId(au.MacId)
			if err == nil {
				au.Cmd = cmdSuccess
			}
		} else if au.Cmd == cmdUnbind {
			codeInfo, err = GetCodeByMacId(au.MacId)
			if err == nil {
				codeInfo.MacId = ""
				au.Cmd = cmdSuccess
			}
		}

		if au.Cmd != cmdSuccess {
			au.Cmd = cmdFail
		} else {
			UpdateDbCode(codeInfo.Code, codeInfo.MacId, codeInfo.Filter)
		}

		sendAuthCmd(conn, &au)
	})
}

func sendAuthCmd(c net.Conn, au *Auth) {
	s, _ := json.Marshal(*au)
	c.Write(s)
}

func readAuthCmd(c net.Conn, au *Auth) error {
	buf := make([]byte, buffsize)
	n, err := c.Read(buf)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf[ : n],  au)
}

func cmdAuthenticate(au *Auth) error {

	con, err := net.DialTimeout("tcp", hostAuth, time.Second*10)
	if err != nil {
		return err
	}
	defer con.Close()

	sendAuthCmd(con, au)

	err = readAuthCmd(con, au)
	if err != nil {
		return err
	}

	if au.Cmd == cmdFail {
		err = ErrorAuthFail
	}

	return err
}

func RandomCode() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 6; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GenAuthCode() (code string) {

	for {
		code = RandomCode()
		if !FindCode(code) {
			break
		}
	}

	return code
}
