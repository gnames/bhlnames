package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
)

// curateData is used for creation of a golden standard. The input is set
// through nomen calculation and the result is sent to terminal for a human
// desition, if found references are nomenclatural events or not.
// For example io/bayes/data/gold.json was created from such curation
// events.
func curateData(out <-chan *namerefs.NameRefs, outputPath string) {
	var res []*namerefs.NameRefs
	for r := range out {
		if r.Error != nil {
			log.Warn().Err(r.Error)
		}
		res = append(res, r)
	}
	curate(res, outputPath)
}

func printKeys() {
	c := color.New(color.FgMagenta)
	c.Println("Selections: [z] right & forward; ['] wrong & next; [b] back;")
}

func curate(nrs []*namerefs.NameRefs, outputPath string) {
	cursor := 0
	printKeys()
	var choice string
NEXT_NR:
	for cursor < len(nrs) {
		name := nrs[cursor].Input.NameString
		refName := nrs[cursor].Input.RefString

		fmt.Println()
		c := color.New(color.FgYellow)
		c.Printf("%d: %s\n", cursor, name)
		c = color.New(color.FgCyan)
		c.Println(refName)

		refs := nrs[cursor].References

		for i := range refs {
			ref := refs[i]

			c := color.New(color.FgYellow)
			c.Printf("%d.%d: ", cursor, i)
			c = color.New(color.FgHiBlue)
			c.Printf("Refs: %d ", nrs[cursor].ReferenceNumber)
			c = color.New(color.FgWhite)
			c.Printf("%s ", ref.TitleName)
			c = color.New(color.FgYellow)
			c.Printf("%s ", ref.Volume)
			c = color.New(color.FgGreen)
			c.Printf("p.%d ", ref.PageNum)
			c = color.New(color.FgCyan, color.Bold)
			c.Printf("%d ", ref.YearAggr)
			c = color.New(color.FgMagenta)
			c.Printf("Score: %d ", ref.Score.Total)
			c = color.New(color.FgRed)
			c.Printf("Odds: %0.2f\n", ref.Score.Odds)
			c = color.New(color.FgBlue)
			c.Printf("URL: %s\n", ref.URL)

			fmt.Scanf("%s", &choice)
			switch choice {
			case "b":
				for ii := range refs {
					refs[ii].IsNomenRef = false
				}
				continue NEXT_NR
			case "'":
				cursor++
				continue NEXT_NR
			case "z":
				refs[i].IsNomenRef = true
			default:
				printKeys()
				continue NEXT_NR
			}
		}
		cursor++
	}
	getOutput(nrs, outputPath)
}

func getOutput(nrs []*namerefs.NameRefs, outputPath string) {
	enc := gnfmt.GNjson{Pretty: true}
	var res []*namerefs.NameRefs
	for i := range nrs {
		if len(nrs[i].References) > 0 {
			res = append(res, nrs[i])
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Input.NameString < res[j].Input.NameString
	})

	resJSON, err := enc.Encode(res)
	if err != nil {
		err = fmt.Errorf("cmd getOutput: %#w", err)
		log.Fatal().Err(err).Msg("getOutput")
	}

	err = os.WriteFile(outputPath, resJSON, 0644)
	if err != nil {
		err = fmt.Errorf("cmd getOutput: %#w", err)
		log.Fatal().Err(err).Msg("GetOutput")
	}
}
