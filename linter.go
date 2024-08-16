package csvlint

import (
	"encoding/csv"
	"fmt"
	"io"
)

// CSVError returns information about an invalid record in a CSV file
type CSVError struct {
	// Record is the invalid record. This will be nil when we were unable to parse a record.
	Record []string
	// Num is the record number of this record.
	Num int
	err error
}

// Error implements the error interface
// TODO: Option to emit malformed records
func (e CSVError) Error() string {
	return fmt.Sprintf("Record #%d: %s", e.Num, e.err.Error())
}

// Validate tests whether or not a CSV lints according to RFC 4180.
// The lazyquotes option will attempt to parse lines that aren't quoted properly.
func Validate(reader io.Reader, delimiter rune, comment rune, lazyquotes bool) ([]CSVError, bool, int, error) {
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1
	r.LazyQuotes = lazyquotes
	r.Comma = delimiter
	r.Comment = comment

	r.ReuseRecord = true // Reuse the same slice for each record for performance

	var header []string
	errors := []CSVError{}
	records := 0
	for {
		record, err := r.Read()
		if header != nil {
			records++
		}
		if err != nil {
			if err == io.EOF {
				break
			}

			parsedErr, ok := err.(*csv.ParseError)
			if !ok {
				return errors, true, records, err
			}

			errors = append(errors, CSVError{
				Record: nil,
				Num:    records,
				err:    parsedErr.Err,
			})

			return errors, true, records, nil
		}
		if header == nil {
			header = record
			continue
		} else if len(record) != len(header) {
			errors = append(errors, CSVError{
				Record: record,
				Num:    records,
				err:    csv.ErrFieldCount,
			})
		}
	}

	return errors, false, records, nil
}
