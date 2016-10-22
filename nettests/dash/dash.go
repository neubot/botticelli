package dash

import (
	"github.com/bassosimone/botticelli/common"
	"log"
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
	w.Write(common.RandByteMaskingImproved(body_size))
}
