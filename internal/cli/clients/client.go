package clients

import (
	"github.com/theotruvelot/g0s/pkg/proto/auth"
	"github.com/theotruvelot/g0s/pkg/proto/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	AuthClient        auth.AuthServiceClient
	HealthcheckClient health.HealthServiceClient
	conn              *grpc.ClientConn
}

func NewClients(serverAddr string) (*Clients, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	return &Clients{
		AuthClient:        auth.NewAuthServiceClient(conn),
		HealthcheckClient: health.NewHealthServiceClient(conn),
		conn:              conn,
	}, nil
}

func (c *Clients) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
