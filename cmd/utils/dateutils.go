package utils

import (
	"time"
)

// Converts a Time instance to the local time zone
func ToLocalTime(when time.Time) time.Time {
	localLoc, err := time.LoadLocation("Local")
	if err != nil {
		return when
	} else {
		return when.In(localLoc)
	}
}

// Formats a time with the local time zone including date, time, am/pm, and zone
func FormatLocalDateTimeAmPmZone(when time.Time) string {
	return ToLocalTime(when).Format("01/02/06 3:04pm MST")
}

// Formats a time with the local time zone including date
func FormatLocalDate(when time.Time) string {
	return ToLocalTime(when).Format("01/02/06")
}

// Formats a time with the local time zone including time and am/pm
func FormatLocalTimeAmPmSecs(when time.Time) string {
	return ToLocalTime(when).Format("3:04:05pm")
}
