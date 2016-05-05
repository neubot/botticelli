package main

import (
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "regexp"
)

type defaultResponse struct {
    Queue_pos int `json:"queue_pos"`
    Real_address string `json:"real_address"`
    Unchoked bool `json:"unchoked"`
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
}

func NegotiateDefaultNegotiate(w http.ResponseWriter, r *http.Request) {
    addr, err := addressWithoutPort(r.RemoteAddr)
    if err != nil {
        w.WriteHeader(500)
        return
    }
    message := &defaultResponse{
        0, addr, true,
    }
    data, err := json.Marshal(message)
    if err != nil {
        w.WriteHeader(500)
        return
    }
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
        BittorrentNegotiate(w, r)
    case match[1] == "dash":
        DashNegotiate(w, r)
    case match[1] == "speedtest":
        SpeedtestNegotiate(w, r)
    case match[1] == "raw":
        RawNegotiate(w, r)
    default:
        log.Println("unknown module")
        http.NotFound(w, r)
    }
}
