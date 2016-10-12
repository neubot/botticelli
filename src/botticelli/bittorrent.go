package main

import (
	"net/http"
)

func BittorrentNegotiate(w http.ResponseWriter, r *http.Request) {
	NegotiateDefaultNegotiate(w, r)
}
