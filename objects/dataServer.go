package objects

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhangmingkai4315/go-storage/locate"
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
	file := getFile(fileName)
	if file == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sendFile(w, file)
}

func getFile(name string) string {
	// file := os.Getenv("STORAGE_ROOT" + "/objects/" + hash)
	// f, _ := os.Open(file)

	// d := url.PathEscape(lib.CalculateHash(f))
	// f.Close()

	// if d != hash {
	// 	log.Println("hash mismatch remove it")
	// 	locate.Del(hash)
	// 	os.Remove(file)
	// 	return ""
	// }
	// return file
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + name + ".*")
	if len(files) != 1 {
		return ""
	}
	file := files[0]
	h := sha256.New()
	sendFile(h, file)
	hash := strings.Split(file, ".")[2]
	d := url.PathEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	if d != hash {
		log.Println("object has mismatch remove", file)
		locate.Del(hash)
		os.Remove(file)
		return ""
	}
	return file

}

func sendFile(w io.Writer, file string) {
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(w, f)
}
