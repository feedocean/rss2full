package syndfeed

import (
	"errors"
	"strings"
	"time"
)

var dateFormats = []string{
	time.RFC822,  // RSS
	time.RFC822Z, // RSS
	time.RFC3339, // Atom
	time.UnixDate,
	time.RubyDate,
	time.RFC850,
	time.RFC1123Z,
	time.RFC1123,
	time.ANSIC,
}

func parseDateString(v string) (t time.Time, err error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return t, errors.New("invlaid date string")
	}
	for _, f := range dateFormats {
		if t, err = time.Parse(f, v); err == nil {
			return
		}
	}
	return t, errors.New("invalid date format")
}
