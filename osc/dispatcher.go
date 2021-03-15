// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	secondsFrom1900To1970 = 2208988800
	bundleTagString       = "#bundle"
)

// Dispatcher is an interface for an OSC message dispatcher. A dispatcher is
// responsible for dispatching received OSC messages.
type Dispatcher interface {
	Dispatch(packet Packet)
}

// Handler is an interface for message handlers. Every handler implementation
// for an OSC message must implement this interface.
type Handler interface {
	HandleMessage(msg *Message)
}

// HandlerFunc implements the Handler interface. Type definition for an OSC
// handler function.
type HandlerFunc func(msg *Message)

// HandleMessage calls itself with the given OSC Message. Implements the
// Handler interface.
func (f HandlerFunc) HandleMessage(msg *Message) {
	f(msg)
}

// StandardDispatcher is a dispatcher for OSC packets. It handles the dispatching of
// received OSC packets to Handlers for their given address.
type StandardDispatcher struct {
	handlers       map[string]Handler
	defaultHandler Handler
}

// NewStandardDispatcher returns an StandardDispatcher.
func NewStandardDispatcher() *StandardDispatcher {
	return &StandardDispatcher{handlers: make(map[string]Handler)}
}

// AddMsgHandler adds a new message handler for the given OSC address.
func (s *StandardDispatcher) AddMsgHandler(addr string, handler HandlerFunc) error {
	if addr == "*" {
		s.defaultHandler = handler
		return nil
	}
	for _, chr := range "*?,[]{}# " {
		if strings.Contains(addr, fmt.Sprintf("%c", chr)) {
			return errors.New("OSC Address string may not contain any characters in \"*?,[]{}#")
		}
	}

	if addressExists(addr, s.handlers) {
		return errors.New("OSC address exists already")
	}

	s.handlers[addr] = handler
	return nil
}

// Dispatch dispatches OSC packets. Implements the Dispatcher interface.
func (s *StandardDispatcher) Dispatch(packet Packet) {
	switch p := packet.(type) {
	default:
		return

	case *Message:
		for addr, handler := range s.handlers {
			if p.Match(addr) {
				handler.HandleMessage(p)
			}
		}
		if s.defaultHandler != nil {
			s.defaultHandler.HandleMessage(p)
		}

	case *Bundle:
		timer := time.NewTimer(p.Timetag.ExpiresIn())

		go func() {
			<-timer.C
			for _, message := range p.Messages {
				for address, handler := range s.handlers {
					if message.Match(address) {
						handler.HandleMessage(message)
					}
				}
				if s.defaultHandler != nil {
					s.defaultHandler.HandleMessage(message)
				}
			}

			// Process all bundles
			for _, b := range p.Bundles {
				s.Dispatch(b)
			}
		}()
	}
}
