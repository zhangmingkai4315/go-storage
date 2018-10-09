package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
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
	log.Printf("content-length is %d", size)
	return size
}

func CalculateHash(r io.Reader) string {
	h := sha256.New()
	io.Copy(h, r)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
