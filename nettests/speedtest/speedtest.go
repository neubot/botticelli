package speedtest

import (
	"github.com/neubot/botticelli/common"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const TARGET = 5
const CHUNK = 8192

func Negotiate(w http.ResponseWriter, r *http.Request) {
}

func Collect(w http.ResponseWriter, r *http.Request) {
}

func Download(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("range") != "" {
		w.Header().Set("Content-Type", "application/octet-stream")

		start_time := time.Now()
		for time.Since(start_time).Seconds() >= TARGET {
			w.Write(common.RandByteMaskingImproved(CHUNK))
		}
		return
	}

	re := regexp.MustCompile("[0-9]+")
	ranges := re.FindAllString(r.Header.Get("range"), -1)
	ranges_int := []int{}

	for _, i := range ranges {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		ranges_int = append(ranges_int, j)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(common.RandByteMaskingImproved(ranges_int[1] - ranges_int[0] - 1))

}

func Latency(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))

}

func Upload(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
