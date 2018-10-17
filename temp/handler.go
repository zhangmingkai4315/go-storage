package temp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/zhangmingkai4315/go-storage/lib"

	"github.com/zhangmingkai4315/go-storage/locate"

	uuid "github.com/satori/go.uuid"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		put(w, r)
		return
	}
	if m == http.MethodPatch {
		patch(w, r)
		return
	}
	if m == http.MethodPost {
		post(w, r)
		return
	}
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type tempInfo struct {
	UUID string
	Name string
	Size int64
}

func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}

func post(w http.ResponseWriter, r *http.Request) {
	rawUUID, _ := uuid.NewV4()
	uuid := fmt.Sprintf("%s", rawUUID)
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	size, err := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ti := tempInfo{uuid, name, size}
	err = ti.writeToFile()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + ti.UUID + ".dat")
	w.Write([]byte(uuid))
}

func (ti *tempInfo) writeToFile() error {
	f, err := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + ti.UUID)
	if err != nil {
		return err
	}

	defer f.Close()
	b, _ := json.Marshal(ti)
	f.Write(b)
	return nil
}

func patch(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempInfo, err := readFromFile(uuid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	dataFile := infoFile + ".dat"
	f, err := os.OpenFile(dataFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer f.Close()

	_, err = io.Copy(f, r.Body)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	info, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actual := info.Size()
	if actual > tempInfo.Size {
		os.Remove(dataFile)
		os.Remove(infoFile)
		log.Println("actual size exceeds")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func readFromFile(uuid string) (*tempInfo, error) {
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	f, err := os.Open(infoFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	var info tempInfo
	json.Unmarshal(b, &info)
	return &info, nil
}

func put(w http.ResponseWriter, r *http.Request) {

	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempInfo, err := readFromFile(uuid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	dataFile := infoFile + ".dat"
	f, err := os.Open(dataFile)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actual := info.Size()
	os.Remove(infoFile)
	if actual != tempInfo.Size {
		os.Remove(dataFile)
		log.Println("actual size not equal expected")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	commitTempObject(dataFile, tempInfo)
}

func commitTempObject(dataFile string, tempinfo *tempInfo) {
	// err := os.Rename(dataFile, os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// locate.Add(tempinfo.Name)

	f, _ := os.Open(dataFile)
	d := url.PathEscape(lib.CalculateHash(f))
	f.Close()
	os.Rename(dataFile, os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name+"."+d)
	locate.Add(tempinfo.hash(), tempinfo.id())
}

func del(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	os.Remove(infoFile)
	os.Remove(datFile)
}
