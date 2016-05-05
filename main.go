package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Antani");
}

func main() {
    http.HandleFunc("/dash/download", DashDownload);

    http.HandleFunc("/collect/*", NegotiateCollect);
    http.HandleFunc("/negotiate/*", NegotiateNegotiate);

    http.HandleFunc("/speedtest/collect", SpeedtestCollect);
    http.HandleFunc("/speedtest/latency", SpeedtestLatency);
    http.HandleFunc("/speedtest/negotiate", SpeedtestNegotiate);
    http.HandleFunc("/speedtest/download", SpeedtestDownload);
    http.HandleFunc("/speedtest/upload", SpeedtestUpload);

    err := http.ListenAndServe(":8080", nil);
    if err != nil {
        log.Fatal(err);
    }
}
