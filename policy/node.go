package policy

import (
	"errors"
	"math"
	"strings"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Node represents a set of requirements the node requesting to open a channel must satisfy.
type Node struct {
	Age          *Range[uint32]      `yaml:"age,omitempty"`
	Capacity     *Range[int64]       `yaml:"capacity,omitempty"`
	Hybrid       *bool               `yaml:"hybrid,omitempty"`
	FeatureFlags *[]lnrpc.FeatureBit `yaml:"feature_flags,omitempty"`
	Channels     *Channels           `yaml:"channels,omitempty"`
}

func (n *Node) evaluate(node *lnrpc.GetInfoResponse, peer *lnrpc.NodeInfo) error {
	if n == nil {
		return nil
	}

	if !n.checkAge(node.BlockHeight, peer.Channels) {
		return errors.New("Node age " + n.Age.Reason())
	}

	if !check(n.Capacity, peer.TotalCapacity) {
		return errors.New("Node capacity " + n.Capacity.Reason())
	}

	if !n.checkHybrid(peer.Node.Addresses) {
		return errors.New("Node doesn't have both clearnet and tor addresses")
	}

	if !n.checkFeatureFlags(peer.Node.Features) {
		return errors.New("Node doesn't have the desired feature flags")
	}

	return n.Channels.evaluate(node.IdentityPubkey, peer)
}

func (n *Node) checkAge(bestBlockHeight uint32, channels []*lnrpc.ChannelEdge) bool {
	if n.Age == nil {
		return true
	}

	if len(channels) == 0 {
		return n.Age.Contains(0)
	}

	oldestChannel := uint32(math.MaxInt32)
	for _, channel := range channels {
		blockHeight := uint32(channel.ChannelId >> 40)
		if blockHeight < oldestChannel {
			oldestChannel = blockHeight
		}
	}

	age := (bestBlockHeight - oldestChannel) + 1
	return n.Age.Contains(age)
}

func (n *Node) checkHybrid(addresses []*lnrpc.NodeAddress) bool {
	if n.Hybrid == nil {
		return true
	}
	hasClearnet := false
	hasTor := false

	for _, address := range addresses {
		host, _, _ := strings.Cut(address.Addr, ":")
		if strings.HasSuffix(host, ".onion") {
			hasTor = true
			continue
		}
		hasClearnet = true
	}

	if hasClearnet && hasTor {
		return *n.Hybrid
	}

	return !*n.Hybrid
}

func (n *Node) checkFeatureFlags(features map[uint32]*lnrpc.Feature) bool {
	if n.FeatureFlags == nil {
		return true
	}

	for _, flag := range *n.FeatureFlags {
		if feature, ok := features[uint32(flag)]; !ok || !feature.IsKnown {
			return false
		}
	}

	return true
}
