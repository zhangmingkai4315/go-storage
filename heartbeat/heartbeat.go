package heartbeat

import (
	"os"
	"time"

	"github.com/zhangmingkai4315/go-storage/lib"
)

// StartHeartBeat send heartbeat to rabbit every 5 seconds
func StartHeartBeat() {
	q := lib.NewRabbitMQ(os.Getenv("STORAGE_MQ_SERVER"))
	defer q.Close()
	q.DeclareExchange("apiServer", "fanout")
	q.DeclareExchange("dataServer", "fanout")

	for {
		q.Publish("apiServer", os.Getenv("STORAGE_PORT"))
		time.Sleep(5 * time.Second)
	}
}
