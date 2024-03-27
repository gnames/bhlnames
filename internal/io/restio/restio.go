package restio

import (
	"fmt"
	"log/slog"
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
	"github.com/sfgrp/lognsq/ent/nsq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var apiPath = "/api/v1"

type restio struct {
	bhlnames.BHLnames
	*echo.Echo
}

// New creates a new REST API server.
func New(bn bhlnames.BHLnames) rest.REST {
	res := restio{BHLnames: bn}
	res.Echo = echo.New()
	res.Use(middleware.Gzip())
	res.Use(middleware.CORS())
	return res
}

// @title BHLnames API
// @version 1.0
// @description This API serves the BHLnames app. It locates relevant sections in the Biodiversity Heritage Library that correspond provided names, references or pages.
// @description
// @description Code repository: https://github.com/gnames/bhlnames.
// @description
// @description Access the API on the production server: https://bhlnames.globalnames.org/api/v1.

// @contact.name Dmitry Mozzherin
// @contact.url https://github.com/dimus
// @contact.email dmozzherin@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// Server Definitions
// @Server https://bhlquest.globalnames.org Description for production server
// @Server http://localhost:8888 Description for local server

// @host bhlnames.globalnames.org
// @host localhost:8888
// @BasePath /api/v1

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func (r restio) Run() {
	str := fmt.Sprintf("Starting the HTTP API server on port %d.", r.Config().PortREST)
	slog.Info(str)

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
	r.GET(apiPath+"/name_refs/:name", nameRefsGet(r.BHLnames))
	r.POST(apiPath+"/name_refs", nameRefsPost(r.BHLnames))
	r.GET(apiPath+"/taxon_refs", taxonRefsGet(r.BHLnames))
	r.POST(apiPath+"/taxon_refs/:name", taxonRefsPost(r.BHLnames))
	r.GET(apiPath+"/references/:page_id", refs(r.BHLnames))
	r.GET(apiPath+"/external_ids/:data_source", externalIDGet(r.BHLnames))
	r.GET(apiPath+"/item_stats/:item_id", itemStatsGet(r.BHLnames))
	r.GET(apiPath+"/items_by_taxon/:taxon_name", itemsByTaxonGet(r.BHLnames))

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
https://bhlnames.globalnames.org/apidoc/index.html`)
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
// @Accept plain
// @Produce json
// @Param page_id path string true "Page ID of a reference." example(6589171)
// @Success 200 {object} refbhl.ReferenceNameBHL "Successful response with data about the reference"
// @Router /references/{page_id} [get]
func refs(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		pageIDStr := c.Param("page_id")
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

// nameRefsGet takes a name, optionally reference and returns
// best matched references to provided data. It can also try to return
// a reference for the nomenclatural event for the name.
// @Summary Finds BHL references for a name
// @Description Finds BHL references for a name, does not include references of synonyms. There is an option to find references for the nomenclatural event of a name.
// @ID get-name-refs
// @Param name path string true "Name to find references for." example("Pardosa moesta")
// @Param reference query string false "Reference data used to filter results." example("Docums Mycol. 34(nos 135-136):50. (2008)")
// @Param nomen_event query boolean false "If true, tries to find nomenclatural event reference." example(true)
// @Accept plain
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Matched references for the provided name"
// @Router /name_refs/{name} [get]
func nameRefsGet(bn bhlnames.BHLnames) func(echo.Context) error {
	bn = bn.ChangeConfig(config.OptWithSynonyms(false))
	return func(c echo.Context) error {
		var inp input.Input
		var res *namerefs.NameRefs

		name := c.Param("name")
		ref := c.QueryParam("reference")
		nomenEvent := c.QueryParam("nomen_event")

		inp.NameString = name
		inp.RefString = ref
		inp.NomenEvent = nomenEvent == "true"

		res, err := bn.NameRefs(inp)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	}
}

// nameRefsPost takes an input.Input with a name, optionally reference and returns
// best matched reference to provided data.
// @Summary Finds BHL references for a name
// @Description Finds BHL references for a name, does not include references of synonyms.
// @ID post-name-refs
// @Param input body input.Input true "Input data"
// @Accept json
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Matched references for the provided name"
// @Router /name_refs [post]
func nameRefsPost(bn bhlnames.BHLnames) func(echo.Context) error {
	enc := gnfmt.GNjson{}
	var err error
	bn = bn.ChangeConfig(config.OptWithSynonyms(false))
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
			slog.Error("nameRefs", "error", err)
			return err
		}
		return err
	}
}

// taxonRefsGet takes a name (could be a synonym), also optionally a reference
// and returns references for all names known for the taxon, including synonyms.
// It can also try to return a referencs for the nomenclatural events
// known for all the names.
// @Summary Finds BHL references for a taxon, including synonyms.
// @Description Finds BHL references for a taxon, including references of synonyms. There is an option to find references for the nomenclatural events for all names of the taxon.
// @ID get-taxon-refs
// @Param name path string true "Name to find references for." example("Pardosa moesta")
// @Param reference query string false "Reference data used to filter results." example("Docums Mycol. 34(nos 135-136):50. (2008)")
// @Param nomen_event query boolean false "If true, tries to find nomenclatural event reference." example(true)
// @Accept plain
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Matched references for the provided name"
// @Router /taxon_refs [get]
func taxonRefsGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return nil
}

// taxonRefsPost takes an input.Input with a name, optionally reference and returns
// best matched reference to provided data.
// @Summary Finds BHL references for a taxon (includes references of synonyms)
// @Description Finds BHL references for a taxon, does include references of synonyms.
// @ID post-taxon-refs
// @Param input body input.Input true "Input data"
// @Accept json
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Matched references for the provided name"
// @Router /taxon_refs [post]
func taxonRefsPost(bn bhlnames.BHLnames) func(echo.Context) error {
	return refsCommon(bn, true)
}

// externalIDGet provides nomenclatural event data for a given external ID.
// @Summary Get nomenclatural event data by external ID from a data source.
// @ID get-external-id
// @Param data_source path string true "Data source name" example("col")
// @Param external_id query string true "External ID" example("BKDDK")
// @Accept plain
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Matched references for the provided external ID"
// @Router /external_ids/{data_source} [get]
func externalIDGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		dataSource := c.Param("data_source")
		if dataSource == "" {
			dataSource = "col"
		}

		externalID := c.QueryParam("external_id")
		if externalID == "" {
			return fmt.Errorf("external_id is required")
		}
		allRefs := c.QueryParam("all_refs") == "true"

		res, err := bn.RefsByExternalID(dataSource, externalID, allRefs)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}

// itemStatsGet provides statistics for a given item.
// @Summary Get taxonomic statistics for a given item.
// @Description Get taxonomic statistics for a given item. Provides most prevalent kingdoms, most prevalent taxa etc.
// @Param item_id path string true "Item ID" example("123456")
// @Accept plain
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Taxonomic statistics for the provided item"
// @Router /item_stats/{item_id} [get]
func itemStatsGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return nil
}

// itemsByTaxonGet provides items for a given taxon.
// @Summary Get items where a given taxon is the most prevalent.
// @Description Get items where a given higher taxon is the most prevalent. For example, if the taxon is 'Aves' it provides items where birds are the most prevalent taxon.
// @Param taxon path string true "Taxon name" example("Aves")
// @Accept plain
// @Produce json
// @Success 200 {object} namerefs.NameRefs  "Items where the taxon is the most prevalent"
// @Router /items_by_taxon [get]
func itemsByTaxonGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return nil
}

func refsCommon(bn bhlnames.BHLnames, withSynonyms bool) func(echo.Context) error {
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
			slog.Error("nameRefs", "error", err)
			return err
		}
		return err
	}
}

func (r restio) setLogger() nsq.NSQ {
	return nil
}
