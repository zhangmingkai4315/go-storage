package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/go-storage/heartbeat"
	"github.com/zhangmingkai4315/go-storage/locate"
	"github.com/zhangmingkai4315/go-storage/objects"
	"github.com/zhangmingkai4315/go-storage/versions"
)

// RunAPIServer start a new api server
func RunAPIServer() {
	hostAndPort := os.Getenv("API_SERVER_PORT")
	go heartbeat.ListenHeartBeat()
	http.HandleFunc("/versions/", versions.Handler)
	http.HandleFunc("/objects/", objects.APIHandler)
	http.HandleFunc("/locate/", locate.APIHandler)
	log.Println("storage api server listen at port " + hostAndPort)
	log.Fatal(http.ListenAndServe(hostAndPort, nil))
}
