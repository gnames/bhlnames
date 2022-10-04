package namesbhlio

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
)

func (n namesbhlio) saveNomenRefs(chIn <-chan []db.NomenRef) error {
	total := 0
	columns := []string{"record_id", "ref"}

	for refs := range chIn {
		total += len(refs)
		transaction, err := n.db.Begin()
		if err != nil {
			return err
		}

		stmt, err := transaction.Prepare(pq.CopyIn("nomen_refs", columns...))
		if err != nil {
			return err
		}

		for _, v := range refs {
			_, err = stmt.Exec(v.RecordID, v.Ref)
			if err != nil {
				return err
			}
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
		fmt.Fprintf(os.Stderr, "\rImported %s nomen refs to db", humanize.Comma(int64(total)))
		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr)

	return nil
}
