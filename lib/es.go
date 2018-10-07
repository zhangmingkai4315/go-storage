package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
}

type hit struct {
	Source Metadata `json:"_source"`
}

type searchResult struct {
	Hits struct {
		Total int
		Hits  []hit
	}
}

func getMetadata(name string, versionID int) (meta Metadata, err error) {
	url := fmt.Sprintf(
		"http://%s/metadata/objects/%s_%d/_source",
		os.Getenv("STORAGE_ES_SERVER"),
		name,
		versionID)
	r, err := http.Get(url)
	r.Header.Add("Content-Type", "application/json;charset=utf-8")
	if err != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("fail to get %s_%d:%d", name, versionID, r.StatusCode)
		return
	}
	resutl, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(resutl, &meta)
	return
}

func SearchLatestVersion(name string) (meta Metadata, err error) {
	url := fmt.Sprintf(
		"http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
		os.Getenv("STORAGE_ES_SERVER"),
		url.PathEscape(name),
	)
	r, err := http.Get(url)
	r.Header.Add("Content-Type", "application/json;charset=utf-8")
	if err != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("fail to get %s:%d", name, r.StatusCode)
		return
	}
	httpResutl, _ := ioutil.ReadAll(r.Body)

	var result searchResult
	json.Unmarshal(httpResutl, &result)
	if len(result.Hits.Hits) != 0 {
		meta = result.Hits.Hits[0].Source
	}
	return
}

func GetMetaData(name string, version int) (Metadata, error) {
	if version == 0 {
		return SearchLatestVersion(name)
	}
	return getMetadata(name, version)
}

func PutMetaData(name string, version int, size int64, hash string) error {
	doc := fmt.Sprintf(
		`{"name":"%s","version":"%d","size":"%d","hash":"%s"}`,
		name,
		version,
		size,
		hash,
	)
	client := http.Client{}
	url := fmt.Sprintf(
		"http://%s/metadata/objects/%s_%d?op_type=create",
		os.Getenv("STORAGE_ES_SERVER"),
		name,
		version)
	request, _ := http.NewRequest("PUT", url, strings.NewReader(doc))
	request.Header.Add("Content-Type", "application/json;charset=utf-8")
	r, err := client.Do(request)
	if err != nil {
		return err
	}
	if r.StatusCode == http.StatusConflict {
		return PutMetaData(name, version+1, size, hash)
	}
	if r.StatusCode != http.StatusCreated {
		result, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf("fail to put metadata:%d %s", r.StatusCode, result)
	}
	return nil
}

func AddVersion(name, hash string, size int64) error {
	version, err := SearchLatestVersion(name)
	if err != nil {
		return err
	}
	return PutMetaData(name, version.Version+1, size, hash)
}

func SearchAllVersions(name string, from, size int) ([]Metadata, error) {
	url := fmt.Sprintf(
		"http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
		os.Getenv("STORAGE_ES_SERVER"),
		from,
		size,
	)
	if name != "" {
		url += "&q=name:" + name
	}
	r, err := http.Get(url)
	log.Printf("url=%s", url)
	r.Header.Add("Content-Type", "application/json;charset=utf-8")
	if err != nil {
		return nil, err
	}
	metas := make([]Metadata, 0)
	httpResult, _ := ioutil.ReadAll(r.Body)
	var result searchResult
	json.Unmarshal(httpResult, &result)
	for i := range result.Hits.Hits {
		metas = append(metas, result.Hits.Hits[i].Source)
	}
	return metas, nil
}
