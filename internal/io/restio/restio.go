package restio

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/rest"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var apiPath = "/api/v1"

type restio struct {
	bn  bhlnames.BHLnames
	cfg config.Config
	*echo.Echo
}

// New creates a new REST API server.
func New(bn bhlnames.BHLnames) rest.REST {
	res := restio{
		bn:  bn,
		cfg: bn.Config(),
	}
	res.Echo = echo.New()
	res.Use(middleware.Gzip())
	res.Use(middleware.CORS())
	return &res
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
func (r *restio) Run() {
	slog.Info("Starting the HTTP API server.", "port", r.cfg.PortREST)

	r.GET("/", info)
	r.GET("/apidoc/*", echoSwagger.WrapHandler)
	r.GET("/api", info)
	r.GET(apiPath, info)
	r.GET(apiPath+"/ping", ping)
	r.GET(apiPath+"/version", ver())
	r.GET(apiPath+"/references/:page_id", refs(r.bn))
	r.GET(apiPath+"/items/:item_id", itemStatsGet(r.bn))
	r.GET(apiPath+"/namerefs/:name", nameRefsGet(r.bn))
	r.POST(apiPath+"/namerefs", nameRefsPost(r.bn))
	r.GET(apiPath+"/cached_refs/:external_id", externalIDGet(r.bn))
	// r.GET(apiPath+"/taxon_items/:taxon_name", itemsByTaxonGet(r.BHLnames))

	addr := fmt.Sprintf(":%d", r.cfg.PortREST)
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
func ver() func(echo.Context) error {
	return func(c echo.Context) error {
		result := bhlnames.GetVersion()
		return c.JSON(http.StatusOK, result)
	}
}

// refs takes pageID and returns corresponding BHL reference metadata.
// @Summary Get BHL reference metadata by pageID
// @Description Retrieves the BHL reference metadata by pageID.
// @ID get-refs
// @Accept plain
// @Produce json
// @Param page_id path integer true "Page ID of a reference." example(6589171)
// @Success 200 {object} bhl.Reference "Successful response with data about the reference"
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
// @Success 200 {object} bhl.RefsByName "Matched references for the provided name"
// @Router /namerefs/{name} [get]
func nameRefsGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		var res *bhl.RefsByName

		name := c.Param("name")
		ref := c.QueryParam("reference")
		nomenEvent := c.QueryParam("nomen_event")
		taxon := c.QueryParam("taxon")

		opts := []input.Option{
			input.OptNameString(name),

			input.OptWithNomenEvent(nomenEvent == "true"),
			input.OptWithTaxon(taxon == "true"),
		}
		if ref != "" {
			opts = append(opts, input.OptRefString(ref))
		}
		inp := input.New(bn.ParserPool(), opts...)

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
// @Success 200 {object} bhl.RefsByName  "Matched references for the provided name"
// @Router /namerefs [post]
func nameRefsPost(bn bhlnames.BHLnames) func(echo.Context) error {
	var err error
	var res *bhl.RefsByName
	return func(c echo.Context) error {
		var inp input.Input
		err = c.Bind(&inp)
		if err != nil {
			return err
		}

		res, err = bn.NameRefs(inp)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	}
}

// externalIDGet provides nomenclatural event data for a given external ID.
// @Summary Get nomenclatural event data by external ID from a data source.
// @ID get-cached-refs
// @Param external_id path string true "External ID" example("BKDDK")
// @Param all_refs query string true "All Cached References" example("true")
// @Accept plain
// @Produce json
// @Success 200 {object} bhl.RefsByName  "Matched references for the provided external ID"
// @Router /cached_refs/{external_id} [get]
func externalIDGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		externalID := c.Param("external_id")
		allRefs := c.QueryParam("all_refs") == "true"

		res, err := bn.RefsByExtID(externalID, 1, allRefs)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}

// itemStatsGet provides metadata and stats of a BHL item.
// @Summary Get metadata and taxonomic statistics of a BHL item.
// @ID get-item
// @Param item_id path integer true "Item ID" example(73397)
// @Accept plain
// @Produce json
// @Success 200 {object} bhl.Item  "BHL item metadata and statistics"
// @Router /items/{item_id} [get]
func itemStatsGet(bn bhlnames.BHLnames) func(echo.Context) error {
	return func(c echo.Context) error {
		itemIDStr := c.Param("item_id")
		itemID, err := strconv.Atoi(itemIDStr)
		if err != nil {
			return err
		}
		res, err := bn.ItemStats(itemID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}
