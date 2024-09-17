package tcp

import (
	"context"
	"net"
)

// handler function
type HanldeFunc func(ctx context.Context, conn net.Conn)

// application server over tcp
type Handler interface {
	Hanlde(ctx context.Context, conn net.Conn)
	Close() error
}
