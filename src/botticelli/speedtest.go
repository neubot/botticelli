package main

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const TARGET = 5
const CHUNK = 8192

func SpeedtestCollect(w http.ResponseWriter, r *http.Request) {
}

func SpeedtestDownload(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("range") != "" {
		w.Header().Set("Content-Type", "application/octet-stream")

		start_time := time.Now()
		for time.Since(start_time).Seconds() >= TARGET {
			w.Write(RandByteMaskingImproved(CHUNK))
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
	w.Write(RandByteMaskingImproved(ranges_int[1] - ranges_int[0] - 1))

}

func SpeedtestNegotiate(w http.ResponseWriter, r *http.Request) {
	NegotiateDefaultNegotiate(w, r)
}

func SpeedtestLatency(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))

}

func SpeedtestUpload(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
