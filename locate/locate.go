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
	"github.com/zhangmingkai4315/go-storage/rs"
)

type LocateMessage struct {
	ID   int
	Addr string
}

var objects = make(map[string]int)

var mutex sync.Mutex

// Locate will return file exist or not
func Locate(hash string) int {
	mutex.Lock()
	defer mutex.Unlock()
	id, ok := objects[hash]
	if !ok {
		return -1
	}
	return id
}

func Add(hash string, id int) {
	mutex.Lock()
	defer mutex.Unlock()
	objects[hash] = id
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
		id := Locate(hash)
		if id != -1 {
			q.Send(msg.ReplyTo, LocateMessage{Addr: os.Getenv("DATA_SERVER_PORT"), ID: id})
		}
	}
}

func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT" + "/objects/*"))
	for i := range files {
		// hash := filepath.Base(files[i])
		// objects[hash] = 1
		file := strings.Split(filepath.Base(files[i]), ".")
		if len(file) != 3 {
			panic(files[i])
		}
		hash := file[0]
		id, err := strconv.Atoi(file[i])
		if err != nil {
			panic(err)
		}
		objects[hash] = id
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
func APIForLocate(name string) (locateInfo map[int]string) {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	q.Publish("dataServer", name)
	c := q.Consume()

	go func() {
		time.Sleep(time.Second * 2)
		q.Close()
	}()

	locateInfo = make(map[int]string)
	for i := 0; i < rs.ALL_SHARDS; i++ {
		msg := <-c
		if len(msg.Body) == 0 {
			return
		}
		var info LocateMessage
		json.Unmarshal(msg.Body, &info)
		locateInfo[info.ID] = info.Addr
	}
	return
}

func Exist(name string) bool {
	return len(APIForLocate(name)) >= rs.DATA_SHARDS
}
