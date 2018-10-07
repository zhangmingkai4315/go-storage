package objects

import (
	"net/http"
)

func APIHandler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		apiPut(w, r)
		return
	}
	if m == http.MethodGet {
		apiGet(w, r)
		return
	}
	if m == http.MethodDelete {
		apiDelete(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		dataPut(w, r)
		return
	}
	if m == http.MethodGet {
		dataGet(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
