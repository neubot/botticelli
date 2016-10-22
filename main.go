package main

import (
	"github.com/bassosimone/botticelli/common"
	//"github.com/bassosimone/botticelli/nettests/bittorrent"
	"github.com/bassosimone/botticelli/nettests/dash"
	"github.com/bassosimone/botticelli/nettests/ndt"
	//"github.com/bassosimone/botticelli/nettests/raw"
	"github.com/bassosimone/botticelli/nettests/speedtest"
	"log"
	"net/http"
	"math/rand"
	"time"
)

func main() {
	// Make sure we seed the random number generator properly
	//   see <http://stackoverflow.com/a/12321192>
	rand.Seed(time.Now().UTC().UnixNano())

	ndt.StartNdtServer(":3001")

	http.HandleFunc("/dash/download", dash.DashDownload)
	http.HandleFunc("/dash/download/", dash.DashDownload)

	http.HandleFunc("/collect/", common.NegotiateCollect)
	http.HandleFunc("/negotiate/", common.NegotiateNegotiate)

	http.HandleFunc("/speedtest/collect", speedtest.SpeedtestCollect)
	http.HandleFunc("/speedtest/latency", speedtest.SpeedtestLatency)
	http.HandleFunc("/speedtest/negotiate", speedtest.SpeedtestNegotiate)
	http.HandleFunc("/speedtest/download", speedtest.SpeedtestDownload)
	http.HandleFunc("/speedtest/upload", speedtest.SpeedtestUpload)

	http.HandleFunc("/", http.NotFound)

	server := &http.Server{Addr: ":8080", Handler: nil}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
