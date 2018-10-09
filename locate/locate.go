package locate

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zhangmingkai4315/go-storage/lib"
)

var objects = make(map[string]int)

var mutex sync.Mutex

// Locate will return file exist or not
func Locate(hash string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	_, ok := objects[hash]
	return ok
}

func Add(hash string) {
	mutex.Lock()
	defer mutex.Unlock()
	objects[hash] = 1
}

func Del(hash string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(objects, hash)
}

// StartLocate will receive new query filename, and
// return current server port if exist or not
func StartLocate() {
	log.Printf("start locate service for query files")
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	defer q.Close()

	q.Bind("dataServer")
	c := q.Consume()

	for msg := range c {
		hash, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			panic(err)
		}
		if Locate(hash) {
			q.Send(msg.ReplyTo, os.Getenv("DATA_SERVER_PORT"))
		}
	}
}

func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT" + "/objects/*"))
	for i := range files {
		hash := filepath.Base(files[i])
		objects[hash] = 1
	}
}

// Handler for locate url process
func APIHandler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	info := APIForLocate(strings.Split(r.URL.EscapedPath(), "/")[2])
	if len(info) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, _ := json.Marshal(info)
	w.Write(b)
}

// APIForLocate get server ip hold the file
func APIForLocate(name string) string {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	q.Publish("dataServer", name)
	c := q.Consume()

	go func() {
		time.Sleep(time.Second * 2)
		q.Close()
	}()
	msg := <-c
	s, _ := strconv.Unquote(string(msg.Body))
	return s
}

func Exist(name string) bool {
	return APIForLocate(name) != ""
}
