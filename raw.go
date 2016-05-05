package main

import (
    "net/http"
)

func RawNegotiate(w http.ResponseWriter, r *http.Request) {
    NegotiateDefaultNegotiate(w, r);
}
