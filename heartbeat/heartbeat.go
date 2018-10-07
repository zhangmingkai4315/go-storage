package heartbeat

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/zhangmingkai4315/go-storage/lib"
)

// StartHeartBeat send heartbeat to rabbit every 5 seconds
func StartHeartBeat(server string) {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	defer q.Close()
	q.DeclareExchange("apiServer", "fanout")
	q.DeclareExchange("dataServer", "fanout")

	for {
		q.Publish("apiServer", server)
		time.Sleep(5 * time.Second)
	}
}

var dataServers = make(map[string]time.Time)
var mutex sync.Mutex

func ListenHeartBeat() {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	defer q.Close()
	q.Bind("apiServer")
	c := q.Consume()
	go removeExpirtDataServer()
	for msg := range c {
		dataServer, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			panic(e)
		}
		mutex.Lock()
		dataServers[dataServer] = time.Now()
		mutex.Unlock()
	}
}

func removeExpirtDataServer() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for s, t := range dataServers {
			if t.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		mutex.Unlock()
	}
}

// GetDataServers return list of dataservers
// type []string
func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	ds := make([]string, 0)
	for s := range dataServers {
		ds = append(ds, s)
	}
	return ds
}

// ChooseRandomDataServer return a random
// choosed data server
func ChooseRandomDataServer() string {
	ds := GetDataServers()
	n := len(ds)
	if n == 0 {
		return ""
	}
	return ds[rand.Intn(n)]
}
