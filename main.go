package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangmingkai4315/go-storage/objects"
)

func main() {
	hostAndPort := os.Getenv("STORAGE_PORT")
	if hostAndPort == "" {
		hostAndPort = "localhost:4000"
	}
	http.HandleFunc("/objects/", objects.Handler)
	log.Println("storage server listen at port " + hostAndPort)
	log.Fatal(http.ListenAndServe(hostAndPort, nil))
}
