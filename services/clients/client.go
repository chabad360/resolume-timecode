package clients

import "context"

type Client interface {
	Start(context.Context) error
	New() error
	SetAddress(string, string) error
}
