package builderio

import (
	"bufio"
	"log/slog"
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

// prepareDOI reads doi.txt file and prepares two maps: one for titles and one
// for parts. Each map has doi id as key and title or part as value.
func (b builderio) prepareDOI() (map[int]string, map[int]string, error) {
	titleMap := make(map[int]string)
	partMap := make(map[int]string)
	slog.Info("Processing doi.txt.")

	path := filepath.Join(b.cfg.ExtractDir, "doi.txt")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Cannot open doi.txt.", "path", path, "error", err)
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
			slog.Error("Cannot convert doi id to int.", "id", fields[doiIDF])
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
