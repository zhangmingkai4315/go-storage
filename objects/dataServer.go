package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func dataPut(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Split(r.URL.EscapedPath(), "/")[2]
	f, e := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + fileName)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, r.Body)
}

func dataGet(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Split(r.URL.EscapedPath(), "/")[2]
	f, e := os.Open(os.Getenv("STORAGE_ROOT") + "/objects/" + fileName)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}
