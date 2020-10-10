package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/gnames/gnames/lib/encode"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const url = "http://:8888/"

var _ = Describe("Rest", func() {
	Describe("NameRefs", func() {
		It("finds references for name-strings", func() {
			var response []*entity.NameRefs
			enc := encode.GNjson{}
			request := []string{
				"Not name", "Bubo bubo", "Pomatomus",
				"Pardosa moesta", "Plantago major var major",
				"Cytospora ribis mitovirus 2",
				"A-shaped rods", "Alb. alba",
				"Diapria conica",
				"Monohamus galloprovincialis Rüschkamp, 1928",
			}
			req, err := enc.Encode(request)
			Expect(err).To(BeNil())
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"name_refs", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())
			enc.Decode(respBytes, &response)

			Expect(len(response)).To(Equal(10))

			bad := response[0]
			Expect(bad.NameString).To(Equal("Not name"))
			Expect(len(bad.References)).To(Equal(0))

			moesta := response[3]
			Expect(moesta.NameString).To(Equal("Pardosa moesta"))
			Expect(len(moesta.References)).To(BeNumerically(">", 10))
			Expect(moesta.CurrentCanonical).To(Equal("Pardosa moesta"))

			gall := response[9]
			Expect(gall.NameString).To(Equal("Monohamus galloprovincialis Rüschkamp, 1928"))
			Expect(gall.Canonical).To(Equal("Monohamus galloprovincialis"))
			Expect(gall.CurrentCanonical).To(Equal("Monochamus galloprovincialis"))
			Expect(len(gall.References)).To(BeNumerically(">", 2))
			Expect(len(gall.References)).To(BeNumerically("<", 10))
		})
	})

	Describe("TaxonRefs", func() {
		It("finds references for name-strings", func() {
			var response []*entity.NameRefs
			enc := encode.GNjson{}
			request := []string{
				"Not name", "Bubo bubo", "Pomatomus",
				"Pardosa moesta", "Plantago major var major",
				"Cytospora ribis mitovirus 2",
				"A-shaped rods", "Alb. alba",
				"Diapria conica (Fabricius, 1775)",
				"Monohamus galloprovincialis Rüschkamp, 1928",
			}
			req, err := enc.Encode(request)
			Expect(err).To(BeNil())
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"taxon_refs", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())
			enc.Decode(respBytes, &response)

			Expect(len(response)).To(Equal(10))

			bad := response[0]
			Expect(bad.NameString).To(Equal("Not name"))
			Expect(bad.Canonical).To(Equal(""))
			Expect(len(bad.References)).To(Equal(0))

			conica := response[8]
			Expect(conica.NameString).To(Equal("Diapria conica (Fabricius, 1775)"))
			Expect(conica.Canonical).To(Equal("Diapria conica"))
			Expect(len(conica.References)).To(Equal(0))

			gall := response[9]
			Expect(gall.NameString).To(Equal("Monohamus galloprovincialis Rüschkamp, 1928"))
			Expect(gall.Canonical).To(Equal("Monohamus galloprovincialis"))
			Expect(gall.CurrentCanonical).To(Equal("Monochamus galloprovincialis"))
			Expect(len(gall.References)).To(BeNumerically(">", 100))
			Expect(len(gall.References)).To(BeNumerically("<", 200))
		})
	})

	Describe("NomenRefs", func() {
		It("finds references for name-strings", func() {
			var response []linkent.Output
			enc := encode.GNjson{}
			request := []linkent.Input{
				{
					ID:        "",
					Name:      linkent.Name{Canonical: "Sagenia longicruris"},
					Reference: linkent.Reference{Year: "1906"},
				},
				{
					ID:        "1",
					Name:      linkent.Name{NameString: "Pseudotrochalus niger"},
					Reference: linkent.Reference{Year: "1903"},
				},
				{
					ID:        "myid",
					Name:      linkent.Name{NameString: "Diapria conica"},
					Reference: linkent.Reference{Year: "1775"},
				},
			}
			req, err := enc.Encode(request)
			Expect(err).To(BeNil())
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"nomen_refs", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())
			enc.Decode(respBytes, &response)

			Expect(len(response)).To(Equal(3))

			match := response[0]
			Expect(len(match.InputID)).To(BeNumerically(">", 10))
			Expect(match.BHLref.AnnotNomen).To(Equal("SP_NOV"))
			Expect(match.Score.Overall).To(BeNumerically(">", 0))

			nomatch := response[2]
			Expect(nomatch.InputID).To(Equal("myid"))
			Expect(nomatch.BHLref).To(BeNil())
			Expect(nomatch.Score.Overall).To(Equal(float32(0.0)))
		})
	})
})
