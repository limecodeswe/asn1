package asn1

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ASN1UTCTime represents an ASN.1 UTCTime value
type ASN1UTCTime struct {
	time time.Time
}

// NewUTCTime creates a new UTCTime with the given time
func NewUTCTime(t time.Time) *ASN1UTCTime {
	return &ASN1UTCTime{
		time: t.UTC(),
	}
}

// NewUTCTimeNow creates a new UTCTime with the current time
func NewUTCTimeNow() *ASN1UTCTime {
	return &ASN1UTCTime{
		time: time.Now().UTC(),
	}
}

// Time returns the time value
func (u *ASN1UTCTime) Time() time.Time {
	return u.time
}

// SetTime sets the time value
func (u *ASN1UTCTime) SetTime(t time.Time) {
	u.time = t.UTC()
}

// Tag returns the ASN.1 tag for UTCTime
func (u *ASN1UTCTime) Tag() Tag {
	return NewUniversalTag(TagUTCTime, false)
}

// Encode returns the BER encoding of the UTCTime
func (u *ASN1UTCTime) Encode() ([]byte, error) {
	// UTCTime format: YYMMDDHHMMSSZ or YYMMDDHHMMSS+HHMM or YYMMDDHHMMSS-HHMM
	// We'll use the Z (UTC) format for simplicity
	timeStr := u.time.Format("060102150405Z")
	return EncodeTLV(u.Tag(), []byte(timeStr))
}

// String returns a string representation of the UTCTime
func (u *ASN1UTCTime) String() string {
	return fmt.Sprintf("UTCTime{%s}", u.time.Format(time.RFC3339))
}

// TaggedString returns a string representation with tag information
func (u *ASN1UTCTime) TaggedString() string {
	return fmt.Sprintf("%s UTCTime: %s", u.Tag().TagString(), u.time.Format(time.RFC3339))
}

// ASN1GeneralizedTime represents an ASN.1 GeneralizedTime value
type ASN1GeneralizedTime struct {
	time time.Time
}

// NewGeneralizedTime creates a new GeneralizedTime with the given time
func NewGeneralizedTime(t time.Time) *ASN1GeneralizedTime {
	return &ASN1GeneralizedTime{
		time: t.UTC(),
	}
}

// NewGeneralizedTimeNow creates a new GeneralizedTime with the current time
func NewGeneralizedTimeNow() *ASN1GeneralizedTime {
	return &ASN1GeneralizedTime{
		time: time.Now().UTC(),
	}
}

// Time returns the time value
func (g *ASN1GeneralizedTime) Time() time.Time {
	return g.time
}

// SetTime sets the time value
func (g *ASN1GeneralizedTime) SetTime(t time.Time) {
	g.time = t.UTC()
}

// Tag returns the ASN.1 tag for GeneralizedTime
func (g *ASN1GeneralizedTime) Tag() Tag {
	return NewUniversalTag(TagGeneralizedTime, false)
}

// Encode returns the BER encoding of the GeneralizedTime
func (g *ASN1GeneralizedTime) Encode() ([]byte, error) {
	// GeneralizedTime format: YYYYMMDDHHMMSSZ or YYYYMMDDHHMMSS+HHMM or YYYYMMDDHHMMSS-HHMM
	// We'll use the Z (UTC) format for simplicity
	timeStr := g.time.Format("20060102150405Z")
	return EncodeTLV(g.Tag(), []byte(timeStr))
}

// String returns a string representation of the GeneralizedTime
func (g *ASN1GeneralizedTime) String() string {
	return fmt.Sprintf("GeneralizedTime{%s}", g.time.Format(time.RFC3339))
}

// TaggedString returns a string representation with tag information
func (g *ASN1GeneralizedTime) TaggedString() string {
	return fmt.Sprintf("%s GeneralizedTime: %s", g.Tag().TagString(), g.time.Format(time.RFC3339))
}

// DecodeUTCTime decodes a UTCTime from BER data
func DecodeUTCTime(data []byte) (*ASN1UTCTime, int, error) {
	value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode TLV: %w", err)
	}

	expectedTag := NewUniversalTag(TagUTCTime, false)
	if value.Tag() != expectedTag {
		return nil, 0, fmt.Errorf("expected UTCTime tag %+v, got %+v", expectedTag, value.Tag())
	}

	timeValue, err := parseUTCTime(string(value.Value()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse UTCTime: %w", err)
	}

	return NewUTCTime(timeValue), consumed, nil
}

// DecodeGeneralizedTime decodes a GeneralizedTime from BER data
func DecodeGeneralizedTime(data []byte) (*ASN1GeneralizedTime, int, error) {
	value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode TLV: %w", err)
	}

	expectedTag := NewUniversalTag(TagGeneralizedTime, false)
	if value.Tag() != expectedTag {
		return nil, 0, fmt.Errorf("expected GeneralizedTime tag %+v, got %+v", expectedTag, value.Tag())
	}

	timeValue, err := parseGeneralizedTime(string(value.Value()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse GeneralizedTime: %w", err)
	}

	return NewGeneralizedTime(timeValue), consumed, nil
}

// parseUTCTime parses a UTCTime string
func parseUTCTime(timeStr string) (time.Time, error) {
	// Remove any whitespace
	timeStr = strings.TrimSpace(timeStr)
	
	if len(timeStr) < 11 {
		return time.Time{}, fmt.Errorf("UTCTime string too short: %s", timeStr)
	}

	// Extract components
	yearStr := timeStr[0:2]
	monthStr := timeStr[2:4]
	dayStr := timeStr[4:6]
	hourStr := timeStr[6:8]
	minStr := timeStr[8:10]
	secStr := timeStr[10:12]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year: %s", yearStr)
	}
	
	// UTCTime uses 2-digit years, with 50-99 meaning 1950-1999, and 00-49 meaning 2000-2049
	if year >= 50 {
		year += 1900
	} else {
		year += 2000
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %s", dayStr)
	}

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", hourStr)
	}

	min, err := strconv.Atoi(minStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", minStr)
	}

	sec, err := strconv.Atoi(secStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid second: %s", secStr)
	}

	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC), nil
}

// parseGeneralizedTime parses a GeneralizedTime string
func parseGeneralizedTime(timeStr string) (time.Time, error) {
	// Remove any whitespace
	timeStr = strings.TrimSpace(timeStr)
	
	if len(timeStr) < 15 {
		return time.Time{}, fmt.Errorf("GeneralizedTime string too short: %s", timeStr)
	}

	// Extract components
	yearStr := timeStr[0:4]
	monthStr := timeStr[4:6]
	dayStr := timeStr[6:8]
	hourStr := timeStr[8:10]
	minStr := timeStr[10:12]
	secStr := timeStr[12:14]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year: %s", yearStr)
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %s", dayStr)
	}

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", hourStr)
	}

	min, err := strconv.Atoi(minStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", minStr)
	}

	sec, err := strconv.Atoi(secStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid second: %s", secStr)
	}

	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC), nil
}