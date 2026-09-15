package timeutil

import (
	"log"
	"strings"
	"time"
)

var loc *time.Location

func SetLocation(name string) {
	var err error
	loc, err = time.LoadLocation(name)
	if err != nil {
		log.Printf("timeutil: failed to load %s, falling back to UTC: %v", name, err)
		loc = time.UTC
	}
	log.Printf("timeutil: location set to %s", loc.String())
}

func init() {
	loc = time.UTC
}

func Now() time.Time {
	return time.Now().In(loc)
}

func Today() string {
	return Now().Format("2006-01-02")
}

func TimeStr() string {
	return Now().Format("15:04")
}

func Location() *time.Location {
	return loc
}

// FormatEU converts an ISO date string (YYYY-MM-DD) to European format (DD.MM.YYYY).
// If the input is already in DD.MM.YYYY or cannot be parsed, it is returned as-is.
func FormatEU(isoDate string) string {
	if isoDate == "" {
		return ""
	}
	// Already in DD.MM.YYYY format?
	if strings.Contains(isoDate, ".") {
		return isoDate
	}
	t, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		return isoDate
	}
	return t.Format("02.01.2006")
}
