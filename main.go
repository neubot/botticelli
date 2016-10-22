package main

import (
	"github.com/bassosimone/botticelli/nettests/ndt"
	//"log"
	//"net/http"
	"math/rand"
	"time"
)

func main() {
	// Make sure we seed the random number generator properly
	//   see <http://stackoverflow.com/a/12321192>
	rand.Seed(time.Now().UTC().UnixNano())

	ndt.StartNdtServer(":3001")

	// Disabled: for now we shall coexist with neubot-server
/*
	http.HandleFunc("/dash/download", DashDownload)
	http.HandleFunc("/dash/download/", DashDownload)

	http.HandleFunc("/collect/", NegotiateCollect)
	http.HandleFunc("/negotiate/", NegotiateNegotiate)

	http.HandleFunc("/speedtest/collect", SpeedtestCollect)
	http.HandleFunc("/speedtest/latency", SpeedtestLatency)
	http.HandleFunc("/speedtest/negotiate", SpeedtestNegotiate)
	http.HandleFunc("/speedtest/download", SpeedtestDownload)
	http.HandleFunc("/speedtest/upload", SpeedtestUpload)

	http.HandleFunc("/", http.NotFound)

	server := &http.Server{Addr: ":8080", Handler: nil}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
*/
}
