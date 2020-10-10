package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/gnames/lib/encode"
	"github.com/gorilla/mux"
)

func Run(api APIProvider) {
	log.Printf("Starting the HTTP API server on port %d.", api.Port())
	r := mux.NewRouter()

	r.HandleFunc("/name_refs", nameRefsHTTP(api)).Methods("POST")
	r.HandleFunc("/taxon_refs", taxonRefsHTTP(api)).Methods("POST")
	r.HandleFunc("/nomen_refs", nomenRefsHTTP(api)).Methods("POST")

	addr := fmt.Sprintf(":%d", api.Port())

	server := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func nameRefsHTTP(api APIProvider) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		enc := encode.GNjson{}
		var err error
		var body []byte
		var nameStrings []string
		if body, err = ioutil.ReadAll(req.Body); err != nil {
			log.Printf("nameRefsHTTP: cannot read message from request : %s.", err)
			return
		}

		err = enc.Decode(body, &nameStrings)
		if err != nil {
			log.Printf("nameRefsHTTP: cannot decode request's body: %s", err)
			return
		}

		res := api.NameRefs(nameStrings)
		resJSON, err := enc.Encode(res)
		if err != nil {
			log.Printf("nameRefsHTTP: cannot encode NameRefs: %s", err)
			return
		}
		w.Write(resJSON)
	}
}

func taxonRefsHTTP(api APIProvider) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		enc := encode.GNjson{}
		var err error
		var body []byte
		var nameStrings []string
		if body, err = ioutil.ReadAll(req.Body); err != nil {
			log.Printf("taxonRefsHTTP: cannot read message from request : %s.", err)
			return
		}

		err = enc.Decode(body, &nameStrings)
		if err != nil {
			log.Printf("taxonRefsHTTP: cannot decode request's body: %s", err)
			return
		}

		res := api.TaxonRefs(nameStrings)
		resJSON, err := enc.Encode(res)
		if err != nil {
			log.Printf("taxonRefsHTTP: cannot encode NameRefs: %s", err)
			return
		}
		w.Write(resJSON)
	}
}

func nomenRefsHTTP(api APIProvider) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		enc := encode.GNjson{}
		var err error
		var body []byte
		var inputs []linkent.Input
		if body, err = ioutil.ReadAll(req.Body); err != nil {
			log.Printf("nomenRefsHTTP: cannot read message from request : %s.", err)
			return
		}

		err = enc.Decode(body, &inputs)
		if err != nil {
			log.Printf("nomenRefsHTTP: cannot decode request's body: %s", err)
			return
		}

		res := api.NomenRefs(inputs)
		resJSON, err := enc.Encode(res)
		if err != nil {
			log.Printf("nomenRefsHTTP: cannot encode bhlinker.Output: %s", err)
			return
		}
		w.Write(resJSON)
	}
}
