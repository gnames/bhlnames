package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gnames/bhlnames"
	"github.com/gnames/gnames/lib/encode"
	"github.com/gnames/gnames/lib/format"
	"github.com/gorilla/mux"
)

func Run(bhln bhlnames.BHLnames) {
	log.Printf("Starting the HTTP API server on port %d.", 8080)
	r := mux.NewRouter()

	r.HandleFunc("/",
		func(resp http.ResponseWriter, req *http.Request) {
			homeHTTP(resp, req, bhln)
		})
	addr := fmt.Sprintf(":%d", 8080)

	server := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func homeHTTP(resp http.ResponseWriter, req *http.Request, bhln bhlnames.BHLnames) {
	enc := encode.GNjson{}
	res, _ := bhln.Refs("Pomatomus saltator")
	out := enc.Output(res, format.PrettyJSON)
	resp.Write([]byte(out))
}
