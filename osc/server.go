package osc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	data = make([]byte, 65535)
	buf  = bytes.NewBuffer(data)
	n    int
	err  error
)

// Server represents an OSC server. The server listens on Address and Port for
// incoming OSC packets and bundles.
type Server struct {
	Addr        string
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
}

// ListenAndServe retrieves incoming OSC packets and dispatches the retrieved
// OSC packets.
func (s *Server) ListenAndServe() error {
	if s.Dispatcher == nil {
		s.Dispatcher = NewStandardDispatcher()
	}

	ln, err := net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	return s.Serve(ln)
}

// Serve retrieves incoming OSC packets from the given connection and dispatches
// retrieved OSC packets. If something goes wrong an error is returned.
func (s *Server) Serve(c net.PacketConn) error {
	var tempDelay time.Duration
	for {
		msg, err := s.readFromConnection(c)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		go s.Dispatcher.Dispatch(msg)
	}
}

// ReceivePacket listens for incoming OSC packets and returns the packet if one is received.
func (s *Server) ReceivePacket(c net.PacketConn) (Packet, error) {
	return s.readFromConnection(c)
}

// readFromConnection retrieves OSC packets.
func (s *Server) readFromConnection(c net.PacketConn) (Packet, error) {
	if s.ReadTimeout != 0 {
		if err = c.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, err
		}
	}

	n, _, err = c.ReadFrom(data)
	if err != nil {
		return nil, err
	}

	buf.Reset()
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	var start int
	return readPacket(buf, &start, n)
}

// ParsePacket parses the given msg string and returns a Packet
func ParsePacket(msg string) (Packet, error) {
	var start int
	return readPacket(bytes.NewBufferString(msg), &start, len(msg))
}

// receivePacket receives an OSC packet from the given reader.
func readPacket(reader *bytes.Buffer, start *int, end int) (Packet, error) {
	var b byte
	b, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}

	err = reader.UnreadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '/': // An OSC Message starts with a '/'
		return readMessage(reader, start)
	case '#': // An OSC bundle starts with a '#'
		return readBundle(reader, start, end)
	}

	return nil, fmt.Errorf("readPacket: invalid packet")
}

// readBundle reads an Bundle from reader.
func readBundle(reader *bytes.Buffer, start *int, end int) (*Bundle, error) {
	// Read the '#bundle' OSC string
	var startTag string
	startTag, n, err = readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	if startTag != bundleTagString {
		return nil, fmt.Errorf("invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var timeTag uint64
	if err = binary.Read(reader, binary.BigEndian, &timeTag); err != nil {
		return nil, err
	}
	*start += 8

	// Create a new bundle
	bundle := &Bundle{Timetag: *NewTimetagFromTimetag(timeTag)}

	// Read until the end of the buffer
	for *start < end {
		// Read the size of the bundle element
		var length int32
		if err = binary.Read(reader, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		*start += 4

		var p Packet
		p, err = readPacket(reader, start, end)
		if err != nil {
			return nil, err
		}
		if err = bundle.Append(p); err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

// readMessage from `reader`.
func readMessage(reader *bytes.Buffer, start *int) (*Message, error) {
	// First, read the OSC address
	var addr string
	addr, n, err = readPaddedString(reader)
	if err != nil {
		return nil, err
	}
	*start += n

	// Read all arguments
	msg := &Message{Address: addr}
	if err = readArguments(msg, reader, start); err != nil {
		return nil, err
	}

	return msg, nil
}

// readArguments from `reader` and add them to the OSC message `msg`.
func readArguments(msg *Message, reader *bytes.Buffer, start *int) error {
	// Read the type tag string
	var typetags string
	typetags, n, err = readPaddedString(reader)
	if err != nil {
		return err
	}
	*start += n

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return errors.New("unsupported type tag string")
	}

	// Remove ',' from the type tag
	typetags = typetags[1:]

	for _, c := range typetags {
		switch c {
		default:
			return fmt.Errorf("unsupported type tag: %c", c)

		case 'i': // int32
			var i int32
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			*start += 4
			msg.Append(i)

		case 'h': // int64
			var i int64
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			*start += 8
			msg.Append(i)

		case 'f': // float32
			var f float32
			if err = binary.Read(reader, binary.BigEndian, &f); err != nil {
				return err
			}
			*start += 4
			msg.Append(f)

		case 'd': // float64/double
			var d float64
			if err = binary.Read(reader, binary.BigEndian, &d); err != nil {
				return err
			}
			*start += 8
			msg.Append(d)

		case 's': // string
			// TODO: fix reading string value
			var s string
			if s, _, err = readPaddedString(reader); err != nil {
				return err
			}
			*start += len(s) + padBytesNeeded(len(s))
			msg.Append(s)

		case 'b': // blob
			var buf []byte
			var n int
			if buf, n, err = readBlob(reader); err != nil {
				return err
			}
			*start += n
			msg.Append(buf)

		case 't': // OSC time tag
			var tt uint64
			if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
				return nil
			}
			*start += 8
			msg.Append(NewTimetagFromTimetag(tt))

		case 'N': // nil
			msg.Append(nil)

		case 'T': // true
			msg.Append(true)

		case 'F': // false
			msg.Append(false)
		}
	}

	return nil
}
