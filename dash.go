package main

import (
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func DashDownload(w http.ResponseWriter, r *http.Request) {

	//DASH_DEFAULT_BODY_SIZE := 1000
	DASH_MAXIMUM_BODY_SIZE := 104857600
	DASH_MAXIMUM_REPETITIONS := 60

	if !strings.HasPrefix(r.URL.Path, "/dash/download/") {
		log.Println("dash: unexpected URI")
		http.NotFound(w, r)
	}

	count := 0
	count++

	if count > DASH_MAXIMUM_REPETITIONS {
		log.Print("dash: too many repetitions")
		return
		//TODO close connection
	}

	//body_size := DASH_MAXIMUM_BODY_SIZE
	pattern := regexp.MustCompile(`/dash/download/([0-9+]*)$`)
	resource_size := pattern.FindStringSubmatch(r.URL.Path)
	if len(resource_size) < 2 {
		log.Println("Body size not valid")
		return
	}

	body_size, err := strconv.Atoi(resource_size[1])
	if err != nil {
		log.Println("dash: error body_syze type cast")
		return
	}

	if body_size < 0 {
		log.Println("dash: negative body size")
		return
	}
	if body_size > DASH_MAXIMUM_BODY_SIZE {
		body_size = DASH_MAXIMUM_BODY_SIZE
	}

	w.Header().Set("mimetype", "video/mp4")
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
