// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Dmitry Mozzherin",
            "url": "https://github.com/dimus",
            "email": "dmozzherin@gmail.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "Gives information where to find docs.",
                "produces": [
                    "text/plain"
                ],
                "summary": "Information about the API documentation",
                "operationId": "get-info",
                "responses": {
                    "200": {
                        "description": "API documentation URL",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "Checks if the API is online and returns a simple response if it is.",
                "produces": [
                    "text/plain"
                ],
                "summary": "Check API status",
                "operationId": "get-ping",
                "responses": {
                    "200": {
                        "description": "API status response",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/references/{page-id}": {
            "get": {
                "description": "Retrieves the BHL reference metadata by pageID.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get BHL reference metadata by pageID",
                "operationId": "get-refs",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page ID of a reference.",
                        "name": "page-id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with data about the reference",
                        "schema": {
                            "$ref": "#/definitions/refbhl.ReferenceNameBHL"
                        }
                    }
                }
            }
        },
        "/version": {
            "get": {
                "description": "Retrieves the current version of the BHLnames application.",
                "produces": [
                    "application/json"
                ],
                "summary": "Get BHLnames version",
                "operationId": "get-version",
                "responses": {
                    "200": {
                        "description": "Successful response with version information",
                        "schema": {
                            "$ref": "#/definitions/gnvers.Version"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "gnvers.Version": {
            "type": "object",
            "properties": {
                "build": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "output.OddsDetails": {
            "type": "object",
            "additionalProperties": {
                "type": "number"
            }
        },
        "refbhl.ItemStats": {
            "description": "ItemStats provides insights about a Reference's Item. This data can be used to infer the prevalent taxonomic groups within the Item.",
            "type": "object",
            "properties": {
                "itemKingdom": {
                    "description": "ItemKingdom a consensus kingdom for names from the Item (journal volume).",
                    "type": "string",
                    "example": "Animalia"
                },
                "itemKingdomPercent": {
                    "description": "ItemKingdomPercent indicates the percentage of names that belong\nto the consensus Kingdom.",
                    "type": "integer",
                    "example": 81
                },
                "itemMainTaxon": {
                    "description": "ItemMainTaxon provides a clade that contains a majority of scientific names\nmentioned in the Item.",
                    "type": "string",
                    "example": "Arthropoda"
                },
                "statNamesNum": {
                    "description": "UniqNamesNum is the number of unique names in the Item.",
                    "type": "integer",
                    "example": 1234
                }
            }
        },
        "refbhl.NameData": {
            "description": "NameData contains details about a scientific name provided in the search.",
            "type": "object",
            "properties": {
                "annotNomen": {
                    "description": "AnnotNomen is a nomenclatural annotation located near the matchted name.",
                    "type": "string",
                    "example": "sp. nov."
                },
                "editDistance": {
                    "description": "EditDistance is the number of differences (edit events)\nbetween Name and MatchName according to Levenshtein algorithm.",
                    "type": "integer",
                    "example": 0
                },
                "matchName": {
                    "description": "MatchedName is a scientific name match from the reference's text.",
                    "type": "string",
                    "example": "Pardosa moesta Banks, 1892"
                },
                "name": {
                    "description": "Name is a scientific name from the query.",
                    "type": "string",
                    "example": "Pardosa moesta"
                }
            }
        },
        "refbhl.Part": {
            "description": "Part represents a distinct entity, usually a scientific paper,",
            "type": "object",
            "properties": {
                "doiPart": {
                    "description": "DOI provides DOI for a part (usually a paper/publication).",
                    "type": "string",
                    "example": "10.1234/5678"
                },
                "partId": {
                    "description": "ID is the BHL database ID for the Part (usually a scientific paper).",
                    "type": "integer",
                    "example": 12345
                },
                "partName": {
                    "description": "Name is the publication title.",
                    "type": "string",
                    "example": "The choanal papillae of the Cheloniidae"
                },
                "partPages": {
                    "description": "Pages are the start and end pages of a publication.",
                    "type": "string",
                    "example": "123-145"
                }
            }
        },
        "refbhl.Reference": {
            "description": "Reference represents a BHL reference that matched the query. This could be a book, a journal, or a scientific paper.",
            "type": "object",
            "properties": {
                "doiTitle": {
                    "description": "TitleDOI provides DOI for a book or journal",
                    "type": "string",
                    "example": "10.1234/5678"
                },
                "itemId": {
                    "description": "ItemID is the BHL database ID for Item (usually a volume).",
                    "type": "integer",
                    "example": 12345
                },
                "itemStats": {
                    "description": "ItemStats provides insights about the Reference Item.\nFrom this data it is possible to infer what kind of\ntaxonomic groups are prevalent in the text.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/refbhl.ItemStats"
                        }
                    ]
                },
                "itemYearEnd": {
                    "description": "ItemYearEnd is the year when an Item ceased publication.",
                    "type": "integer",
                    "example": 1893
                },
                "itemYearStart": {
                    "description": "ItemYearStart is the year when an Item began publication (most\nitems will have only ItemYearStart).",
                    "type": "integer",
                    "example": 1892
                },
                "pageId": {
                    "description": "PageID is the BHL database ID for the page where the name was found.\nIt is provided by BHL.",
                    "type": "integer",
                    "example": 12345
                },
                "pageNum": {
                    "description": "PageNum is the page number provided by the hard copy of the publication.",
                    "type": "integer",
                    "example": 123
                },
                "part": {
                    "description": "Part corresponds to a scientific paper, or other\ndistinct entity in an Item.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/refbhl.Part"
                        }
                    ]
                },
                "titleId": {
                    "description": "TitleID is the BHL database ID for the Title (book or journal).",
                    "type": "integer",
                    "example": 12345
                },
                "titleName": {
                    "description": "TitleName is the name of a title (a book or a journal).",
                    "type": "string",
                    "example": "Bulletin of the American Museum of Natural History"
                },
                "titleYearEnd": {
                    "description": "TitleYearEnd is the year when the journal ceased publication.",
                    "type": "integer",
                    "example": 1922
                },
                "titleYearStart": {
                    "description": "TitleYearStart is the year the when book is published,\nor when the journal was first published.",
                    "type": "integer",
                    "example": 1890
                },
                "url": {
                    "description": "URL is the URL of the reference in BHL.",
                    "type": "string",
                    "example": "https://www.biodiversitylibrary.org/page/12345"
                },
                "volume": {
                    "description": "Volume is the information about a volume in a journal.",
                    "type": "string",
                    "example": "vol. 12"
                },
                "yearAggr": {
                    "description": "YearAggr is the most precise year information available for the\nreference. This could be from the reference year (part),\nthe year of a Volume (item), or from the title (usually a book\nor journal).",
                    "type": "integer",
                    "example": 1892
                },
                "yearType": {
                    "description": "YearType indicates the source of the YearAggr value.",
                    "type": "string",
                    "example": "part"
                }
            }
        },
        "refbhl.ReferenceNameBHL": {
            "description": "ReferenceNameBHL represents a BHL entity that includes a matched scientific name and the reference where this name was discovered.",
            "type": "object",
            "properties": {
                "isNomenRef": {
                    "description": "IsNomenRef states is the reference likely contains\na nomenclatural event for the name.",
                    "type": "boolean",
                    "example": true
                },
                "name": {
                    "description": "NameData contains detailed information about the scientific name.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/refbhl.NameData"
                        }
                    ]
                },
                "reference": {
                    "description": "Reference is the BHL reference where the name was detected.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/refbhl.Reference"
                        }
                    ]
                },
                "score": {
                    "description": "Score is the overall score of the match between the reference and\na name-string or a reference-string.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/refbhl.Score"
                        }
                    ]
                }
            }
        },
        "refbhl.Score": {
            "description": "Score provides a qualitative estimation of a match quality to a name-string, a nomen, or a reference-string.",
            "type": "object",
            "properties": {
                "annot": {
                    "description": "Annot is a score important for nomenclatural events and provides match\nfor nomenclatural annotations.",
                    "type": "integer",
                    "example": 3
                },
                "labels": {
                    "description": "Labels provide types for each match",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "odds": {
                    "description": "Odds is total Naive Bayes odds for the score.",
                    "type": "number",
                    "example": 0.1234
                },
                "oddsDetail": {
                    "description": "OddsDetail provides details of the odds calculation.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/output.OddsDetails"
                        }
                    ]
                },
                "pages": {
                    "description": "RefPages is a score derived from matching pages in a reference\nand a page in BHL.",
                    "type": "integer",
                    "example": 3
                },
                "title": {
                    "description": "RefTitle is the score of matching reference's titleName.",
                    "type": "integer"
                },
                "total": {
                    "description": "Total is a simple sum of all available individual scores.",
                    "type": "integer",
                    "example": 15
                },
                "volume": {
                    "description": "RefVolume is a score derived from matching volume from\nreference and BHL Volume.",
                    "type": "integer",
                    "example": 3
                },
                "year": {
                    "description": "Year is a score representing the quality of a year match\nin a reference-string or the name-string.",
                    "type": "integer",
                    "example": 3
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "bhlnames.globalnames.org",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "BHLnames API",
	Description:      "This API serves the BHLnames app. It locates relevant sections in the Biodiversity Heritage Library that correspond provided names, references or pages. \\n\\nCode repository: https://github.com/gnames/bhlnames. \\n\\nAccess the API on the production server: https://bhlnames.globalnames.org/api/v1.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
