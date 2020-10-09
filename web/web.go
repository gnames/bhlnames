package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gnames/bhlnames"
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
	resp.Write([]byte("hello"))
}
