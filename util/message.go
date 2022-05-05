package util

import "github.com/francoispqt/gojay"

type Message struct {
	Hour   string
	Minute string
	Second string
	MS     string

	ClipLength string
	ClipName   string
	Message    string
	Invert     bool
	AlertTime  int
}

func (m *Message) MarshalJSONObject(enc *gojay.Encoder) {
	enc.StringKey("hour", m.Hour)
	enc.StringKey("minute", m.Minute)
	enc.StringKey("second", m.Second)
	enc.StringKey("ms", m.MS)
	enc.StringKey("cliplength", m.ClipLength)
	enc.StringKey("clipname", m.ClipName)
	enc.StringKey("message", m.Message)
	enc.BoolKey("invert", m.Invert)
	enc.IntKey("alerttime", m.AlertTime)
}

func (m *Message) IsNil() bool {
	return m == nil
}
