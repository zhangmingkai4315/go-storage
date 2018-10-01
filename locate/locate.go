package locate

import (
	"os"
	"strconv"

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
			q.Send(msg.ReplyTo, os.Getenv("STORAGE_PORT"))
		}
	}
}
