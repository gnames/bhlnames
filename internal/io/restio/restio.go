package restio

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/gnames/bhlnames/docs"
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
	echoSwagger "github.com/swaggo/echo-swagger"
)

var apiPath = "/api/v1"

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

// @title BHLnames API
// @version 1.0
// @description This API serves the BHLnames app. It locates relevant sections in the Biodiversity Heritage Library that correspond provided names, references or pages. \n\nCode repository: https://github.com/gnames/bhlnames. \n\nAccess the API on the production server: https://bhlnames.globalnames.org/api/v1.

// @contact.name Dmitry Mozzherin
// @contact.url https://github.com/dimus
// @contact.email dmozzherin@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// Server Definitions
// @Server http://localhost:8888 Description for local server
// @Server https://bhlquest.globalnames.org Description for production server

// @host bhlnames.globalnames.org
// @host localhost:8888
// @BasePath /api/v1
func (r restio) Run() {
	log.Info().Msgf("Starting the HTTP API server on port %d.", r.Config().PortREST)

	loggerNSQ := r.setLogger()
	if loggerNSQ != nil {
		defer loggerNSQ.Stop()
	}
	r.GET("/", info)
	r.GET("/apidoc/*", echoSwagger.WrapHandler)
	r.GET("/api", info)
	r.GET(apiPath, info)
	r.GET(apiPath+"/ping", ping)
	r.GET(apiPath+"/version", ver(r.BHLnames))
	r.POST(apiPath+"/name_refs", nameRefs(r.BHLnames, false))
	r.POST(apiPath+"/taxon_refs", nameRefs(r.BHLnames, true))
	r.POST(apiPath+"/nomen_refs", nomenRefs(r.BHLnames))
	r.GET(apiPath+"/references/:page-id", refs(r.BHLnames))

	addr := fmt.Sprintf(":%d", r.Config().PortREST)
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	r.Logger.Fatal(r.StartServer(s))
}

// info gives information where to find docs.
// @Summary Information about the API documentation
// @Description Gives information where to find docs.
// @ID get-info
// @Produce plain
// @Success 200 {string} string "API documentation URL"
// @Router / [get]
func info(c echo.Context) error {
	return c.String(http.StatusOK,
		`The REST API is described at
https://apidoc.globalnames.org/bhlames`)
}

// ping checks if the API is online
// @Summary Check API status
// @Description Checks if the API is online and returns a simple response if it is.
// @ID get-ping
// @Produce plain
// @Success 200 {string} string "API status response"
// @Router /ping [get]
func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

// ver returns back the version of BHLnames
// @Summary Get BHLnames version
// @Description Retrieves the current version of the BHLnames application.
// @ID get-version
// @Produce json
// @Success 200 {object} gnvers.Version "Successful response with version information"
// @Router /version [get]
func ver(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		result := bn.GetVersion()
		return c.JSON(http.StatusOK, result)
	}
}

// refs takes pageID and returns corresponding BHL reference metadata.
// @Summary Get BHL reference metadata by pageID
// @Description Retrieves the BHL reference metadata by pageID.
// @ID get-refs
// @Produce json
// @Param page-id path integer true "Page ID of a reference."
// @Success 200 {object} refbhl.ReferenceNameBHL "Successful response with data about the reference"
// @Router /references/{page-id} [get]
func refs(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		pageIDStr := c.Param("page-id")
		log.Print(pageIDStr)
		pageID, err := strconv.Atoi(pageIDStr)
		if err != nil {
			return err
		}
		res, err := bn.RefByPageID(pageID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
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
