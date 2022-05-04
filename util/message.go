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
}

func (m *Message) MarshalJSONObject(enc *gojay.Encoder) {
	enc.StringKey("hour", m.Hour)
	enc.StringKey("minute", m.Minute)
	enc.StringKey("second", m.Second)
	enc.StringKey("ms", m.MS)
	enc.StringKey("clip_length", m.ClipLength)
	enc.StringKey("clip_name", m.ClipName)
	enc.StringKey("message", m.Message)
	enc.BoolKey("invert", m.Invert)
}

func (m *Message) IsNil() bool {
	return m == nil
}
