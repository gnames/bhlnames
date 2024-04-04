package input

import (
	"regexp"
	"strconv"
)

var pagePatterns = []*regexp.Regexp{
	regexp.MustCompile(
		`[\d]+[\s]*\:[\s]*([\d]+)[\s]*[\-]{0,2}[\s]*([\d]*)`,
	), // matches 12: 188-189
	regexp.MustCompile(
		`[\d]+[\s]*\(.+\)[\s]*\:[\s]*([\d]+)[\-]{0,2}[\s]*([\d]*)`,
	), //matches 12(issue): 188-189
	regexp.MustCompile(
		`([\d]+)[\s]*[\-]{0,2}[\s]*([\d]*)[pp]{1,2}`,
	), // matches 12pp, 12p
	regexp.MustCompile(
		`[Ppg][.]*[\s]*([\d]+)`,
	), // matches Pg. 20, P. 20, P.20, P20
}

var volumePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b[Vv]ol[\.]*[\s]*([\d]+)`),
	regexp.MustCompile(`([\d]+)[\s]*\:[\s]*[\d]+`),
	regexp.MustCompile(`([\d]+)[\s]*\(.+\)[\s]*\:[\s]*[\d]+`),
	regexp.MustCompile(`[Vv][.]*[\s]*([\d]+)`),
	regexp.MustCompile(`([\d]+):\s*No`),
}

var yearPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\((17|18|19|20)(\d\d)[\s]*-?[\s]*((17|18|19|20)(\d\d)|(\d\d)){0,1}\)`),
	regexp.MustCompile(`(17|18|19|20)(\d\d)[\s]*-?[\s]*((17|18|19|20)(\d\d)|(\d\d)){0,1}`),
}

// [2008-2009) 20 08 2009 20]
func parseYears(ref string) []int {
	var m []string
	for _, re := range yearPatterns {
		m = re.FindStringSubmatch(ref)
		if m != nil {
			yearStart, _ := strconv.Atoi(m[1] + m[2])
			if len(m) < 3 {
				return []int{yearStart, 0}
			}
			yearEnd := 0
			if m[5] != "" {
				yearEnd, _ = strconv.Atoi(m[1] + m[5])
			} else if m[6] != "" {
				yearEnd, _ = strconv.Atoi(m[1] + m[6])
			}
			return []int{yearStart, yearEnd}
		}
	}
	return []int{0, 0}
}

func parseVolume(ref string) int {
	var m []string
	for _, p := range volumePatterns {
		m = p.FindStringSubmatch(ref)
		if m != nil {
			volume, _ := strconv.Atoi(m[1])
			return volume
		}
	}
	return 0
}

func parsePages(ref string) []int {
	var m []string
	for _, p := range pagePatterns {
		m = p.FindStringSubmatch(ref)
		if m != nil {
			pageStart, _ := strconv.Atoi(m[1])
			if len(m) < 3 {
				return []int{pageStart, 0}
			}
			pageEnd, _ := strconv.Atoi(m[2])
			return []int{pageStart, pageEnd}
		}
	}
	return []int{0, 0}
}

func parseRefString(inp *Input) {
	if inp.PageStart == 0 {
		pages := parsePages(inp.RefString)
		inp.PageStart = pages[0]
		inp.PageEnd = pages[1]
	}

	if inp.RefYearStart == 0 {
		years := parseYears(inp.RefString)
		inp.RefYearStart = years[0]
		inp.RefYearEnd = years[1]
	}

	if inp.Volume == 0 {
		inp.Volume = parseVolume(inp.RefString)
	}
}
