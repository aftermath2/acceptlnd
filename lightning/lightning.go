// Package lightning connects to the lightning network daemon and exposes an interface with the
// methods available to use.
package lightning

import (
	"context"
	"os"
	"time"

	"github.com/aftermath2/acceptlnd/config"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

// Client represents a lightning node client.
type Client interface {
	ChannelAcceptor(ctx context.Context, opts ...grpc.CallOption) (lnrpc.Lightning_ChannelAcceptorClient, error)
	GetInfo(ctx context.Context, in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error)
	GetNodeInfo(ctx context.Context, in *lnrpc.NodeInfoRequest, opts ...grpc.CallOption) (*lnrpc.NodeInfo, error)
}

// NewClient returns a new lightning client.
func NewClient(config config.Config) (Client, error) {
	opts, err := loadGRPCOpts(config)
	if err != nil {
		return nil, errors.Wrap(err, "loading grpc options")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.RPCAddress, opts...)
	if err != nil {
		return nil, err
	}

	return lnrpc.NewLightningClient(conn), nil
}

func loadGRPCOpts(config config.Config) ([]grpc.DialOption, error) {
	tlsCert, err := credentials.NewClientTLSFromFile(config.CertificatePath, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to read TLS certificate")
	}

	macBytes, err := os.ReadFile(config.MacaroonPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading macaroon file")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return nil, errors.Wrap(err, "unmarshaling macaroon")
	}

	macaroon, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return nil, errors.Wrap(err, "creating macaroon credential")
	}

	return []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(tlsCert),
		grpc.WithPerRPCCredentials(macaroon),
	}, nil
}
