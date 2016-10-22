package common

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
)

type defaultResponse struct {
	Authorization string `json:"authorization"`
	QueuePos      int    `json:"queue_pos"`
	RealAddress   string `json:"real_address"`
	Unchoked      bool   `json:"unchoked"`
}

func addressWithoutPort(s string) (string, error) {
	pattern := regexp.MustCompile("^(.*):[0-9_]+$")
	match := pattern.FindStringSubmatch(s)
	if len(match) != 2 {
		return "", errors.New("invalid input address")
	}
	return match[1], nil
}

func NegotiateCollect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func NegotiateDefaultNegotiate(w http.ResponseWriter, r *http.Request) {
	addr, err := addressWithoutPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	message := &defaultResponse{
		Authorization: "deadbeef",
		QueuePos:      0,
		RealAddress:   addr,
		Unchoked:      true,
	}
	data, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func NegotiateNegotiate(w http.ResponseWriter, r *http.Request) {

	// Find name of the module for which we are negotiating

	pattern := regexp.MustCompile("^/negotiate/([A-Za-z0-9_]+)$")
	match := pattern.FindStringSubmatch(r.URL.Path)
	if len(match) != 2 {
		log.Println("invalid module name")
		http.NotFound(w, r)
		return
	}

	// Dispatch control to the specified module

	switch {
	case match[1] == "bittorrent":
	case match[1] == "dash":
	case match[1] == "speedtest":
	case match[1] == "raw":
		NegotiateDefaultNegotiate(w, r)
	default:
		log.Println("unknown module")
		http.NotFound(w, r)
	}
}
