package builder_pg

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	doiKindF = 0
	doiIDF   = 1
	doiF     = 2
)

func (b BuilderPG) prepareDOI() (map[int]string, map[int]string, error) {
	titleMap := make(map[int]string)
	partMap := make(map[int]string)
	log.Println("Processing doi.txt")
	path := filepath.Join(b.Config.DownloadDir, "doi.txt")
	f, err := os.Open(path)
	if err != nil {
		return titleMap, partMap, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	header := true
	for scanner.Scan() {
		if header {
			header = false
			continue
		}
		l := scanner.Text()
		fields := strings.Split(l, "\t")
		id, err := strconv.Atoi(fields[doiIDF])
		if err != nil {
			return titleMap, partMap, err
		}
		switch fields[doiKindF] {
		case "Part":
			partMap[id] = fields[doiF]
		case "Title":
			titleMap[id] = fields[doiF]
		}
	}
	return titleMap, partMap, nil
}
