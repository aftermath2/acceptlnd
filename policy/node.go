package policy

import (
	"errors"
	"strings"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Node represents a set of requirements the node requesting to open a channel must satisfy.
type Node struct {
	Capacity     *Range[int64]       `yaml:"capacity,omitempty"`
	Hybrid       *bool               `yaml:"hybrid,omitempty"`
	FeatureFlags *[]lnrpc.FeatureBit `yaml:"feature_flags,omitempty"`
	Channels     *Channels           `yaml:"channels,omitempty"`
}

func (n *Node) evaluate(nodePubKey string, peerNode *lnrpc.NodeInfo) error {
	if n == nil {
		return nil
	}

	if !check(n.Capacity, peerNode.TotalCapacity) {
		return errors.New("Node capacity " + n.Capacity.Reason())
	}

	if !n.checkHybrid(peerNode.Node.Addresses) {
		return errors.New("Node doesn't have both clearnet and tor addresses")
	}

	if !n.checkFeatureFlags(peerNode.Node.Features) {
		return errors.New("Node doesn't have the desired feature flags")
	}

	return n.Channels.evaluate(nodePubKey, peerNode)
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
