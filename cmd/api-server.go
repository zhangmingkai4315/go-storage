package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/go-storage/heartbeat"
	"github.com/zhangmingkai4315/go-storage/objects"
)

func RunAPIServer() {

	go heartbeat.StartHeartBeat()
	// go locate.StartLocate()
	hostAndPort := os.Getenv("STORAGE_PORT")
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate", locate.Handler)
	log.Println("storage api server listen at port " + hostAndPort)
	log.Fatal(http.ListenAndServe(hostAndPort, nil))

}
