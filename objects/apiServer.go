package objects

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/zhangmingkai4315/go-storage/heartbeat"
	"github.com/zhangmingkai4315/go-storage/lib"
	"github.com/zhangmingkai4315/go-storage/locate"
)

func apiPut(w http.ResponseWriter, r *http.Request) {

	hash := lib.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println("missing hash in header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := storeObject(r.Body, url.PathEscape(hash))
	if err != nil {
		log.Println(err)
		w.WriteHeader(c)
		return
	}
	if c != http.StatusOK {
		w.WriteHeader(c)
		return
	}

	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	size := lib.GetSizeFromHeader(r.Header)
	err = lib.AddVersion(name, hash, size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func storeObject(r io.Reader, object string) (int, error) {
	stream, err := putStream(object)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	io.Copy(stream, r)
	err = stream.Close()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func putStream(object string) (*PutStream, error) {
	server := heartbeat.ChooseRandomDataServer()
	if server == "" {
		return nil, fmt.Errorf("no data server avaliable")
	}

	return NewPutStream(server, object), nil

}

func apiGet(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	versionId := r.URL.Query()["version"]
	version := 0
	var err error
	if len(versionId) != 0 {
		version, err = strconv.Atoi(versionId[0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	meta, err := lib.GetMetaData(name, version)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	object := url.PathEscape(meta.Hash)
	stream, err := getStream(object)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.Copy(w, stream)
}

func getStream(object string) (io.Reader, error) {
	server := locate.APIForLocate(object)
	if server == "" {
		return nil, fmt.Errorf("object %s locate fail", object)
	}
	return NewGetStream(server, object)
}

func apiDelete(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	version, err := lib.SearchLatestVersion(name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = lib.PutMetaData(name, version.Version+1, 0, "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
