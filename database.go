package ziphttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
)

const (
	bucketRoot    = "Root"
	defaultFilter = "default"
)

var defaultDb *bolt.DB
var errNotFound = errors.New("NotFound")
var errExisted = errors.New("AlreadyExisted")

type CodeInfo struct {
	Code   string
	MacId  string
	Filter string
}

type UserInfo struct {
	User    string
	Pwd     string
	Codes   []CodeInfo
	Filters map[string][]string
}

func OpenDb() (err error) {
	if defaultDb != nil {
		return errors.New("AlreadyOpened")
	}

	defaultDb, err = bolt.Open("authentication.db", 0600, nil)
	if err != nil {
		return err
	}

	err = defaultDb.Update(func(tx *bolt.Tx) error {

		_, e := tx.CreateBucketIfNotExists([]byte(bucketRoot))
		return e
	})

	return err
}

func CloseDb() error {
	return defaultDb.Close()
}

func CreateDbUser(usrname, pwd string) error {
	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		ui := UserInfo{User: usrname, Pwd: pwd}
		ui.Codes = []CodeInfo{}
		ui.Filters = make(map[string][]string, 1)
		ui.Filters[defaultFilter] = []string{""}
		js, _ := json.Marshal(&ui)
		return broot.Put([]byte(usrname), js)
	})
}

func UpdateDbPassword(usrname, pwd string) error {
	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		js := broot.Get([]byte(usrname))
		if js == nil {
			return errNotFound
		}
		ui := UserInfo{}
		json.Unmarshal(js, &ui)
		ui.Pwd = pwd
		js, _ = json.Marshal(&ui)
		return broot.Put([]byte(usrname), js)
	})
}

func CreateDbCode(usrname string, code string) error {
	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		js := broot.Get([]byte(usrname))
		if js == nil {
			return errNotFound
		}
		ui := UserInfo{}
		json.Unmarshal(js, &ui)
		ui.Codes = append(ui.Codes, CodeInfo{code, "", defaultFilter})
		js, _ = json.Marshal(&ui)
		return broot.Put([]byte(usrname), js)
	})
}

func UpdateDbCode(code, macId, filter string) error {

	usrInfo, err := GetUserByCode(code)
	if err != nil {
		return err
	}
	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		js := broot.Get([]byte(usrInfo.User))
		ui := UserInfo{}
		json.Unmarshal(js, &ui)
		for i, _ := range ui.Codes {
			if ui.Codes[i].Code == code {
				ui.Codes[i].MacId = macId
				ui.Codes[i].Filter = filter
				break
			}
		}
		js, _ = json.Marshal(&ui)
		return broot.Put([]byte(usrInfo.User), js)
	})
}

//Create a new filter if not existed
func UpdateDbFilter(usrname string, filter string, files ...string) error {

	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		js := broot.Get([]byte(usrname))
		if js == nil {
			return errNotFound
		}
		ui := UserInfo{}
		json.Unmarshal(js, &ui)
		ui.Filters[filter] = files
		js, _ = json.Marshal(&ui)
		return broot.Put([]byte(usrname), js)
	})
}

func ChangeDbFilter(code string, filter string) error {

	usrInfo, err := GetUserByCode(code)
	if err != nil {
		return err
	}
	if usrInfo.Filters[filter] == nil {
		return errNotFound
	}
	return defaultDb.Update(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		js := broot.Get([]byte(usrInfo.User))
		ui := UserInfo{}
		json.Unmarshal(js, &ui)
		for i, _ := range ui.Codes {
			if ui.Codes[i].Code == code {
				ui.Codes[i].Filter = filter
				break
			}
		}
		js, _ = json.Marshal(&ui)
		return broot.Put([]byte(usrInfo.User), js)
	})
}

func WalkDb(f func(*UserInfo) int) {

	defaultDb.View(func(tx *bolt.Tx) error {
		broot := tx.Bucket([]byte(bucketRoot))
		c := broot.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			ui := &UserInfo{}
			json.Unmarshal(v, ui)
			if f(ui) == 0 { //0 : break, !0 : continue
				break
			}
		}
		return nil
	})
}

func ViewDbUsers() {

	WalkDb(func(ui *UserInfo) int {
		js, _ := json.Marshal(ui)
		fmt.Println(string(js))
		return 1
	})
}

func GetUserByCode(code string) (usrInfo *UserInfo, err error) {

	err = errNotFound
	usrInfo = nil
	WalkDb(func(ui *UserInfo) int {
		for _, codeInfo := range ui.Codes {
			if codeInfo.Code == code {
				usrInfo = ui
				err = nil
				return 0
			}
		}
		return 1
	})

	return usrInfo, err
}

func GetUserByMacId(macId string) (usrInfo *UserInfo, err error) {

	err = errNotFound
	usrInfo = nil
	WalkDb(func(ui *UserInfo) int {
		for _, codeInfo := range ui.Codes {
			if codeInfo.MacId == macId {
				usrInfo = ui
				err = nil
				return 0
			}
		}
		return 1
	})

	return usrInfo, err
}

func GetCodeByMacId(macId string) (codeInfo *CodeInfo, err error) {

	err = errNotFound
	codeInfo = nil
	WalkDb(func(ui *UserInfo) int {
		for _, v := range ui.Codes {
			if v.MacId == macId {
				codeInfo = &CodeInfo{v.Code, v.MacId, v.Filter}
				err = nil
				return 0
			}
		}
		return 1
	})

	return codeInfo, err
}

func GetCodeByCode(code string) (codeInfo *CodeInfo, err error) {

	err = errNotFound
	codeInfo = nil
	WalkDb(func(ui *UserInfo) int {
		for _, v := range ui.Codes {
			if v.Code == code {
				codeInfo = &CodeInfo{v.Code, v.MacId, v.Filter}
				err = nil
				return 0
			}
		}
		return 1
	})

	return codeInfo, err
}

func FindCode(code string) (ret bool) {

	ret = false
	WalkDb(func(ui *UserInfo) int {
		for _, codeInfo := range ui.Codes {
			if codeInfo.Code == code {
				ret = true
				return 0
			}
		}
		return 1
	})

	return ret
}
