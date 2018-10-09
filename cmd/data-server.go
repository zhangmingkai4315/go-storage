package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/go-storage/heartbeat"
	"github.com/zhangmingkai4315/go-storage/locate"
	"github.com/zhangmingkai4315/go-storage/objects"
	"github.com/zhangmingkai4315/go-storage/temp"
)

// RunDataServer start a new data server
func RunDataServer() {
	locate.CollectObjects()
	hostAndPort := os.Getenv("DATA_SERVER_PORT")
	go heartbeat.StartHeartBeat(hostAndPort)
	go locate.StartLocate()
	http.HandleFunc("/objects/", objects.DataHandler)
	http.HandleFunc("/temp/", temp.Handler)
	log.Println("storage data server listen at port " + hostAndPort)
	log.Fatal(http.ListenAndServe(hostAndPort, nil))
}
