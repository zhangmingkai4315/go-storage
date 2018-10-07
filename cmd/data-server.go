package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/go-storage/heartbeat"
	"github.com/zhangmingkai4315/go-storage/locate"
	"github.com/zhangmingkai4315/go-storage/objects"
)

// RunDataServer start a new data server
func RunDataServer() {
	hostAndPort := os.Getenv("DATA_SERVER_PORT")
	go heartbeat.StartHeartBeat(hostAndPort)
	go locate.StartLocate()
	http.HandleFunc("/objects/", objects.DataHandler)
	log.Println("storage data server listen at port " + hostAndPort)
	log.Fatal(http.ListenAndServe(hostAndPort, nil))
}
