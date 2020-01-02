package rpc

import (
	"github.com/gnames/bhlindex/protob"
	"google.golang.org/grpc"
)

type ClientRPC struct {
	Host   string
	Conn   *grpc.ClientConn
	Client protob.BHLIndexClient
}

func (c *ClientRPC) Connect() error {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(c.Host, opts...)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.Client = protob.NewBHLIndexClient(c.Conn)
	return nil
}
