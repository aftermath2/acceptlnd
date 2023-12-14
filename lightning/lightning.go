// Package lightning connects to the lightning network daemon and exposes an interface with the
// methods available to use.
package lightning

import (
	"context"
	"log/slog"
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
		return nil, errors.Wrap(err, "loading GRPC options")
	}

	if config.RPCTimeout == nil {
		defaultTimeout := 60 * time.Second
		config.RPCTimeout = &defaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), *config.RPCTimeout)
	defer cancel()

	if *config.RPCTimeout == 0 {
		ctx = context.Background()
	}

	slog.Info("Connecting to LND",
		slog.String("address", config.RPCAddress),
		slog.String("timeout", config.RPCTimeout.String()),
	)
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
