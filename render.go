package ziphttp

import (
	"net/http"
	"os"
	"log"
	"strings"
	"io/ioutil"
	"html/template"
)

const rootPath = "template/html/private/"

type renderInfo struct {
	File     string
	Title    string
	Abstract string
}
func Render(w http.ResponseWriter, usr string, filter []string) {

	path := rootPath + usr

	var fileInfo []os.FileInfo
	//Default filter
	if filter == nil {
		dir, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return
		}

		//List all the files in user's directory
		fileInfo, err = dir.Readdir(0)
		if err != nil {
			log.Println(err)
			return
		}

		filter = []string{}
		for _, v := range fileInfo {
			if strings.Contains(strings.ToLower(v.Name()), ".pdf") {
				filter = append(filter, v.Name())
			}
		}
	}

	//Collect information of files
	rndInfo := make([]renderInfo, len(filter))
	for i, file := range filter {
		rndInfo[i].File = file
		rndInfo[i].Title = strings.Split(file, ".pdf")[0]
		rndInfo[i].Abstract = "Abstract ..."
		f, err := os.Open(path + "/" + rndInfo[i].Title + ".txt")
		if err == nil {
			content, _ := ioutil.ReadAll(f)
			rndInfo[i].Abstract = string(content)
		}
	}

	//Rending the webpage
	t, _ := template.ParseFiles(rootPath + "index.html")
	t.Execute(w, rndInfo)
}