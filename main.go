package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/aftermath2/acceptlnd/config"
	"github.com/aftermath2/acceptlnd/lightning"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
)

func main() {
	configPath := flag.String("config", "acceptlnd.yml", "Path to the configuration file")
	debug := flag.Bool("debug", false, "Enable debug level logging")
	version := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	level := &slog.LevelVar{}
	if *debug {
		level.Set(slog.LevelDebug)
	}
	loggerOpts := &slog.HandlerOptions{
		AddSource: *debug,
		Level:     level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)

	config, err := config.Load(*configPath)
	if err != nil {
		fatal(err)
	}

	client, err := lightning.NewClient(config)
	if err != nil {
		fatal(err)
	}

	if err := handleChannelRequests(config, client); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}

// handleChannelRequests listens to the ChannnelAcceptor RPC stream and accepts/rejects requests.
func handleChannelRequests(config config.Config, client lightning.Client) error {
	ctx := context.Background()

	stream, err := client.ChannelAcceptor(ctx)
	if err != nil {
		return errors.Wrap(err, "subscribing to the channel acceptor stream")
	}

	slog.Info("Listening for channel requests")
	for {
		req, err := stream.Recv()
		if err != nil {
			return errors.Wrap(err, "receiving channel request")
		}
		slog.Debug("Channel opening request", slog.Any("request", req))

		resp, err := handleRequest(config, client, req)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Accept = true
		}

		if err := stream.Send(resp); err != nil {
			return errors.Wrap(err, "sending channel response")
		}

		logResponse(response{
			accepted:  resp.Accept,
			id:        hex.EncodeToString(req.PendingChanId),
			publicKey: hex.EncodeToString(req.NodePubkey),
			err:       resp.Error,
		})
	}
}

func handleRequest(
	config config.Config,
	client lightning.Client,
	req *lnrpc.ChannelAcceptRequest,
) (*lnrpc.ChannelAcceptResponse, error) {
	ctx := context.Background()
	resp := &lnrpc.ChannelAcceptResponse{Accept: false, PendingChanId: req.PendingChanId}

	node, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return resp, errors.Wrap(err, "getting node information")
	}

	getPeerInfoReq := &lnrpc.NodeInfoRequest{
		PubKey:          hex.EncodeToString(req.NodePubkey),
		IncludeChannels: true,
	}
	peer, err := client.GetNodeInfo(ctx, getPeerInfoReq)
	if err != nil {
		return resp, errors.New("Internal server error")
	}
	slog.Debug("Peer node information", slog.Any("node", peer))

	for _, policy := range config.Policies {
		if err := policy.Evaluate(req, node, peer); err != nil {
			return resp, err
		}

		if policy.MinAcceptDepth != nil {
			resp.MinAcceptDepth = *policy.MinAcceptDepth
		}
	}

	if req.WantsZeroConf && len(config.Policies) != 0 {
		// The initiator requested a zero conf channel and it was explicitly accepted, set the
		// fields required to open it
		resp.ZeroConf = true
		resp.MinAcceptDepth = 0
	}

	return resp, nil
}

type response struct {
	id        string
	publicKey string
	err       string
	accepted  bool
}

func logResponse(res response) {
	args := []any{
		slog.Bool("accepted", res.accepted),
		slog.String("id", res.id),
		slog.String("public_key", res.publicKey),
	}
	if !res.accepted {
		args = append(args, slog.String("error", res.err))
	}

	slog.Info("New request received", args...)
}

func printVersion() {
	bi, _ := debug.ReadBuildInfo()

	var commit string
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			commit = s.Value
			break
		}
	}

	fmt.Println("AcceptLND", bi.Main.Version, commit)
}
