package lib

import (
	"net/http"
	"strconv"
)

func GetHashFromHeader(header http.Header) string {
	digest := header.Get("digest")
	if len(digest) < 9 {
		return ""
	}
	if digest[:8] != "SHA-256=" {
		return ""
	}
	return digest[8:]
}

func GetSizeFromHeader(header http.Header) int64 {
	size, _ := strconv.ParseInt(header.Get("content-length"), 0, 64)
	return size
}
