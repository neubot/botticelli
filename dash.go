package main

import (
    "net/http"
)

func DashDownload(w http.ResponseWriter, r *http.Request) {
}

func DashNegotiate(w http.ResponseWriter, r *http.Request) {
    NegotiateDefaultNegotiate(w, r);
}
