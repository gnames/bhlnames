package restio

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/rest"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/sfgrp/lognsq/ent/nsq"
)

var apiPath = "/api/v0"

type restio struct {
	bhlnames.BHLnames
	*echo.Echo
}

func New(bn bhlnames.BHLnames) rest.REST {
	res := restio{BHLnames: bn}
	res.Echo = echo.New()
	res.Use(middleware.Gzip())
	res.Use(middleware.CORS())
	return res
}

func (r restio) Run() {
	log.Info().Msgf("Starting the HTTP API server on port %d.", r.Config().PortREST)

	loggerNSQ := r.setLogger()
	if loggerNSQ != nil {
		defer loggerNSQ.Stop()
	}
	r.GET("/", info)
	r.GET("/api", info)
	r.GET(apiPath, info)
	r.GET(apiPath+"/ping", ping)
	r.GET(apiPath+"/version", ver(r.BHLnames))
	r.POST(apiPath+"/name_refs", nameRefs(r.BHLnames, false))
	r.POST(apiPath+"/taxon_refs", nameRefs(r.BHLnames, true))
	r.POST(apiPath+"/nomen_refs", nomenRefs(r.BHLnames))

	addr := fmt.Sprintf(":%d", r.Config().PortREST)
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	r.Logger.Fatal(r.StartServer(s))
}

func info(c echo.Context) error {
	return c.String(http.StatusOK,
		`The REST API is described at
https://apidoc.globalnames.org/bhlames`)
}

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func ver(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		result := bn.GetVersion()
		return c.JSON(http.StatusOK, result)
	}
}

func nameRefs(bn bhlnames.BHLnames, withSynonyms bool) func(echo.Context) error {
	enc := gnfmt.GNjson{}
	var err error
	bn = bn.ChangeConfig(config.OptWithSynonyms(withSynonyms))
	var res *namerefs.NameRefs
	return func(c echo.Context) error {
		var inp input.Input
		err = c.Bind(&inp)

		if err == nil {
			res, err = bn.NameRefs(inp)
		}

		if err == nil {
			o := enc.Output(res, gnfmt.CompactJSON)
			err = c.String(http.StatusOK, o)
		}

		if err != nil {
			log.Warn().Err(err).Msg("nameRefs")
		}
		return err
	}
}

func nomenRefs(bn bhlnames.BHLnames) func(echo.Context) error {
	enc := gnfmt.GNjson{}
	var err error
	var res *namerefs.NameRefs
	return func(c echo.Context) error {
		var inp input.Input
		err = c.Bind(&inp)

		if err == nil {
			res, err = bn.NomenRefs(inp)
		}

		if err == nil {
			o := enc.Output(res, gnfmt.CompactJSON)
			err = c.String(http.StatusOK, o)
		}

		if err != nil {
			log.Warn().Err(err).Msg("nomenRefs")
		}
		return err
	}
}

func (r restio) setLogger() nsq.NSQ {
	return nil
}
