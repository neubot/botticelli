package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

const default_body_size = 1000
const maximum_body_size = 104857600
const maximum_repetitions = 60

func DashDownload(w http.ResponseWriter, r *http.Request) {
	var err error

	// For robustness
	if !strings.HasPrefix(r.URL.Path, "/dash/download") {
		log.Println("dash: unexpected URI")
		http.NotFound(w, r)
		return
	}

	// TODO: limit maximum number of requests from a client

	body_size := default_body_size
	resource_size := strings.Replace(r.URL.Path, "/dash/download", "", -1)
	if strings.HasPrefix(resource_size, "/") {
		resource_size = resource_size[1:]
	}
	if resource_size != "" {
		body_size, err = strconv.Atoi(resource_size)
		if err != nil {
			log.Println("dash: error body_syze type cast")
			http.NotFound(w, r)
			return
		}
	}

	if body_size < 0 {
		log.Println("dash: negative body size")
		http.NotFound(w, r)
		return
	}
	if body_size > maximum_body_size {
		body_size = maximum_body_size
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", resource_size)
	w.Write(RandByte(body_size))
}

// XXX taken from https://stackoverflow.com/questions/22892120\
// /how-to-generate-a-random-string-of-a-fixed-length-in-golang
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandByte(n int) []byte {
	b := make([]byte, n)
	src := rand.NewSource(123455)

	// A src.Int63() generates 63 random bits, enough for
	// letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

func DashNegotiate(w http.ResponseWriter, r *http.Request) {
	NegotiateDefaultNegotiate(w, r)
}
