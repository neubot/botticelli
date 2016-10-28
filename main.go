package main

import (
	"github.com/neubot/bernini"
	"github.com/neubot/botticelli/common"
	"github.com/neubot/botticelli/common/negotiate"
	//"github.com/neubot/botticelli/nettests/bittorrent"
	"github.com/neubot/botticelli/nettests/dash"
	"github.com/neubot/botticelli/nettests/ndt"
	//"github.com/neubot/botticelli/nettests/raw"
	"github.com/neubot/botticelli/nettests/speedtest"
	"log"
	"net/http"
)

func main() {
	bernini.InitLogger()
	bernini.InitRng()

	err := bernini.UseSyslog("botticelli")
	if err != nil {
		log.Fatal("cannot initialize syslog")
	}

	log.Printf("botticelli server %s starting up", common.Version)

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
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
