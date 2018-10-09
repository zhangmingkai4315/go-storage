package objects

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type PutStream struct {
	write *io.PipeWriter
	c     chan error
}

type TempPutStream struct {
	Server string
	UUID   string
}

func NewTempPutStream(server, hash string, size int64) (*TempPutStream, error) {
	request, err := http.NewRequest("POST", "http://"+server+"/temp/"+hash, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("size", fmt.Sprintf("%d", size))
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	uuid, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return &TempPutStream{server, string(uuid)}, nil
}

func (w *TempPutStream) Write(p []byte) (n int, err error) {
	request, err := http.NewRequest("PATCH", "http://"+w.Server+"/temp/"+w.UUID, strings.NewReader(string(p)))
	if err != nil {
		return 0, err
	}
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dataserver return http code %d", response.StatusCode)
	}

	return len(p), nil
}

func (w *TempPutStream) Commit(good bool) {
	method := "DELETE"
	if good {
		method = "PUT"
	}
	requst, _ := http.NewRequest(method, "http://"+w.Server+"/temp/"+w.UUID, nil)
	client := http.Client{}
	client.Do(requst)
}

func NewPutStream(server, object string) *PutStream {
	reader, writer := io.Pipe()
	c := make(chan error)

	go func() {
		request, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object, reader)
		client := http.Client{}
		r, err := client.Do(request)
		if err == nil && r.StatusCode != http.StatusOK {
			err = fmt.Errorf("dataserver return http code %d", r.StatusCode)
		}
		c <- err
	}()
	return &PutStream{writer, c}
}

func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.write.Write(p)
}

func (w *PutStream) Close() error {
	w.write.Close()
	return <-w.c
}

type GetStream struct {
	reader io.Reader
}

func createGetStream(url string) (*GetStream, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data server return http code %d", r.StatusCode)
	}

	return &GetStream{r.Body}, nil
}

func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server or object name")
	}
	return createGetStream("http://" + server + "/objects/" + object)
}

func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
