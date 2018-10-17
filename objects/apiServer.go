package objects

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/zhangmingkai4315/go-storage/rs"

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
	size := lib.GetSizeFromHeader(r.Header)
	c, err := storeObject(r.Body, hash, size)
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

	err = lib.AddVersion(name, hash, size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func storeObject(r io.Reader, hash string, size int64) (int, error) {

	if locate.Exist(url.PathEscape(hash)) {
		return http.StatusOK, nil
	}

	stream, err := putStream(url.PathEscape(hash), size)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}

	reader := io.TeeReader(r, stream)
	d := lib.CalculateHash(reader)
	if d != hash {
		stream.Commit(false)
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, calculated=%s, requested=%s", d, hash)
	}
	stream.Commit(true)
	return http.StatusOK, nil
}

func putStream(hash string, size int64) (*TempPutStream, error) {
	servers := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(servers) != rs.ALL_SHARDS {
		return nil, fmt.Errorf("not enough data servers avaliable")
	}
	return rs.NewRSPutStream(servers, hash, size)
}

func apiGet(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	versionID := r.URL.Query()["version"]
	version := 0
	var err error
	if len(versionID) != 0 {
		version, err = strconv.Atoi(versionID[0])
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
	hash := url.PathEscape(meta.Hash)
	stream, err := getStream(hash, meta.Size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.Copy(w, stream)
}

func getStream(hash string, size int64) (io.Reader, error) {
	locateInfo := locate.APIForLocate(hash)
	if len(locateInfo) < rs.DATA_SHARDS {
		return nil, fmt.Errorf("object %s locate fail ,result %v", hash, locateInfo)
	}
	dataServers := make([]string, 0)
	if len(locateInfo) != rs.ALL_SHARDS {
		dataServers = heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS-len(locateInfo), locateInfo)
	}
	return rs.NewGetStream(locateInfo, dataServers, hash, size)
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
