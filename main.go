package main

import (
	"github.com/bassosimone/botticelli/common/negotiate"
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

	ndt.Start(":3001")

	http.HandleFunc("/dash/download", dash.Download)
	http.HandleFunc("/dash/download/", dash.Download)

	http.HandleFunc("/collect/", negotiate.Collect)
	http.HandleFunc("/negotiate/", negotiate.Negotiate)

	http.HandleFunc("/speedtest/collect", speedtest.Collect)
	http.HandleFunc("/speedtest/latency", speedtest.Latency)
	http.HandleFunc("/speedtest/negotiate", speedtest.Negotiate)
	http.HandleFunc("/speedtest/download", speedtest.Download)
	http.HandleFunc("/speedtest/upload", speedtest.Upload)

	http.HandleFunc("/", http.NotFound)

	server := &http.Server{Addr: ":8080", Handler: nil}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
