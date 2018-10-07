package locate

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zhangmingkai4315/go-storage/lib"
)

// Locate will return file exist or not
func Locate(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
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
		object, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			panic(err)
		}
		if Locate(os.Getenv("STORAGE_ROOT") + "/objects/" + object) {
			q.Send(msg.ReplyTo, os.Getenv("DATA_SERVER_PORT"))
		}
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

func APIForLocate(name string) string {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	log.Printf("send query file name [%s] to data servers", name)
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
