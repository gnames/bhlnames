package web

import (
	"log"
	"net/http"

	"github.com/gnames/bhlnames"
)

func Run(bhln bhlnames.BHLnames) {
	log.Printf("Starting the HTTP API server on port %d.", 8080)
	// r := mux.NewRouter()
	//
	// r.HandleFunc("/",
	// 	func(resp http.ResponseWriter, req *http.Request) {
	// 		homeHTTP(resp, req, bhln)
	// 	})
	// addr := fmt.Sprintf("127.0.0.1:%d", 8080)
	// _ = addr
	//
	// server := &http.Server{
	// 	Handler:      r,
	// 	Addr:         "localhost:7777",
	// 	WriteTimeout: 300 * time.Second,
	// 	ReadTimeout:  300 * time.Second,
	// }
	//
	// log.Fatal(server.ListenAndServe())
}

func homeHTTP(resp http.ResponseWriter, req *http.Request, bhln bhlnames.BHLnames) {
	resp.Write([]byte("hello"))
}
