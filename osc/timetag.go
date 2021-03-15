package osc

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	// MinValue is the minimum value of an OSC Time Tag.
	MinValue = uint64(1)
)

// Timetag represents an OSC Time Tag.
// An OSC Time Tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
type Timetag struct {
	timeTag uint64 // The actual time tag
	time    time.Time
}

// NewTimetag returns a new OSC time tag object.
func NewTimetag(timeStamp time.Time) *Timetag {
	return &Timetag{
		time:    timeStamp,
		timeTag: timeToTimetag(timeStamp)}
}

// NewTimetagFromTimetag creates a new Timetag from the given `timetag`.
func NewTimetagFromTimetag(timetag uint64) *Timetag {
	return &Timetag{
		time:    timetagToTime(timetag),
		timeTag: timetag}
}

// Time returns the time.
func (t *Timetag) Time() time.Time {
	return t.time
}

// FractionalSecond returns the last 32 bits of the OSC time tag. Specifies the
// fractional part of a second.
func (t *Timetag) FractionalSecond() uint32 {
	return uint32(t.timeTag << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the
// midnight 1900) from the OSC time tag.
func (t *Timetag) SecondsSinceEpoch() uint32 {
	return uint32(t.timeTag >> 32)
}

// TimeTag returns the time tag value
func (t *Timetag) TimeTag() uint64 {
	return t.timeTag
}

// MarshalBinary converts the OSC time tag to a byte array.
func (t *Timetag) MarshalBinary() ([]byte, error) {
	data := new(bytes.Buffer)
	if err := binary.Write(data, binary.BigEndian, t.timeTag); err != nil {
		return []byte{}, err
	}
	return data.Bytes(), nil
}

// SetTime sets the value of the OSC time tag.
func (t *Timetag) SetTime(time time.Time) {
	t.time = time
	t.timeTag = timeToTimetag(time)
}

// ExpiresIn calculates the number of seconds until the current time is the
// same as the value of the time tag. It returns zero if the value of the
// time tag is in the past.
func (t *Timetag) ExpiresIn() time.Duration {
	if t.timeTag <= 1 {
		return 0
	}

	tt := timetagToTime(t.timeTag)
	seconds := tt.Sub(time.Now())

	if seconds <= 0 {
		return 0
	}

	return seconds
}

// timeToTimetag converts the given time to an OSC time tag.
//
// An OSC time tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
//
// The time tag value consisting of 63 zero bits followed by a one in the least
// significant bit is a special case meaning "immediately."
func timeToTimetag(time time.Time) (timetag uint64) {
	timetag = uint64((secondsFrom1900To1970 + time.Unix()) << 32)
	return timetag + uint64(uint32(time.Nanosecond()))
}

// timetagToTime converts the given timetag to a time object.
func timetagToTime(timetag uint64) (t time.Time) {
	return time.Unix(int64((timetag>>32)-secondsFrom1900To1970), int64(timetag&0xffffffff))
}
