package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateNode(t *testing.T) {
	node := &lnrpc.GetInfoResponse{IdentityPubkey: "node_public_key"}
	peerPublicKey := "peer_public_key"
	defaultPeer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			PubKey: peerPublicKey,
		},
	}
	tru := true
	max := int64(1)
	maxu32 := uint32(1)

	cases := []struct {
		node *Node
		peer *lnrpc.NodeInfo
		desc string
		fail bool
	}{
		{
			desc: "Nil node",
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "Empty node",
			node: &Node{},
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "Capacity",
			node: &Node{
				Capacity: &Range[int64]{
					Max: &max,
				},
			},
			peer: &lnrpc.NodeInfo{
				TotalCapacity: 100_000_000,
				Node:          defaultPeer.Node,
			},
			fail: true,
		},
		{
			desc: "Age",
			node: &Node{
				Age: &Range[uint32]{
					Max: &maxu32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Channels: []*lnrpc.ChannelEdge{
					{
						ChannelId: 623702369048395776,
					},
				},
			},
			fail: true,
		},
		{
			desc: "Hybrid",
			node: &Node{
				Hybrid: &tru,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
					Addresses: []*lnrpc.NodeAddress{
						{Network: "tcp", Addr: "127.0.0.1:9735"},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Feature flags",
			node: &Node{
				FeatureFlags: &[]lnrpc.FeatureBit{
					lnrpc.FeatureBit_AMP_REQ,
					lnrpc.FeatureBit_AMP_OPT,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
					Features: map[uint32]*lnrpc.Feature{
						uint32(lnrpc.FeatureBit_AMP_REQ): {IsKnown: true},
					},
				},
			},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.node.evaluate(node, tc.peer)
			if tc.fail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCheckAge(t *testing.T) {
	bestBlockHeight := uint32(820931)
	one := uint32(1)
	two := uint32(2)

	cases := []struct {
		desc     string
		age      *Range[uint32]
		channels []*lnrpc.ChannelEdge
		expected bool
	}{
		{
			desc: "Min",
			age: &Range[uint32]{
				Min: &one,
			},
			channels: []*lnrpc.ChannelEdge{
				{ChannelId: 902611085473873920}, // 820920
				{ChannelId: 902622080590151680}, // 820930
			},
			expected: true,
		},
		{
			desc: "Min no match",
			age: &Range[uint32]{
				Min: &two,
			},
			channels: []*lnrpc.ChannelEdge{
				{ChannelId: 902623180101779456}, // 820931
			},
			expected: false,
		},
		{
			desc: "Max",
			age: &Range[uint32]{
				Max: &one,
			},
			channels: []*lnrpc.ChannelEdge{
				{ChannelId: 902623180101779456}, // 820931
			},
			expected: true,
		},
		{
			desc: "Max no match",
			age: &Range[uint32]{
				Max: &one,
			},
			channels: []*lnrpc.ChannelEdge{
				{ChannelId: 902620981078523904}, // 820929
				{ChannelId: 902623180101779456}, // 820931
			},
			expected: false,
		},
		{
			desc: "Range",
			age: &Range[uint32]{
				Min: &one,
				Max: &two,
			},
			channels: []*lnrpc.ChannelEdge{
				{ChannelId: 902622080590151680}, // 820930
				{ChannelId: 902623180101779456}, // 820931
			},
			expected: true,
		},
		{
			desc: "No channels min",
			age: &Range[uint32]{
				Min: &one,
			},
			channels: nil,
			expected: false,
		},
		{
			desc: "No channels max",
			age: &Range[uint32]{
				Max: &one,
			},
			channels: nil,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			node := Node{
				Age: tc.age,
			}

			actual := node.checkAge(bestBlockHeight, tc.channels)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		node := Node{}
		actual := node.checkAge(bestBlockHeight, nil)
		assert.True(t, actual)
	})
}

func TestCheckHybrid(t *testing.T) {
	cases := []struct {
		desc      string
		addresses []*lnrpc.NodeAddress
		hybrid    bool
		expected  bool
	}{
		{
			desc: "Hybrid",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "url.onion:9735"},
				{Addr: "0.0.0.0:9735"},
			},
			hybrid:   true,
			expected: true,
		},
		{
			desc: "Hybrid (no tor)",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "0.0.0.0:9735"},
			},
			hybrid:   true,
			expected: false,
		},
		{
			desc: "Hybrid (no clearnet)",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "url.onion:9735"},
			},
			hybrid:   true,
			expected: false,
		},
		{
			desc: "Not hybrid",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "url.onion:9735"},
				{Addr: "0.0.0.0:9735"},
			},
			hybrid:   false,
			expected: false,
		},
		{
			desc: "Clearnet",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "0.0.0.0:9735"},
			},
			hybrid:   false,
			expected: true,
		},
		{
			desc: "Tor",
			addresses: []*lnrpc.NodeAddress{
				{Addr: "url.onion:9735"},
			},
			hybrid:   false,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			node := Node{
				Hybrid: &tc.hybrid,
			}

			actual := node.checkHybrid(tc.addresses)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("None", func(t *testing.T) {
		node := Node{
			Hybrid: nil,
		}

		actual := node.checkHybrid(nil)
		assert.True(t, actual)
	})
}

func TestCheckFeatureFlags(t *testing.T) {
	cases := []struct {
		featureFlags *[]lnrpc.FeatureBit
		features     map[uint32]*lnrpc.Feature
		desc         string
		expected     bool
	}{
		{
			desc: "Knows features",
			featureFlags: &[]lnrpc.FeatureBit{
				lnrpc.FeatureBit_AMP_REQ,
				lnrpc.FeatureBit_AMP_OPT,
			},
			features: map[uint32]*lnrpc.Feature{
				uint32(lnrpc.FeatureBit_AMP_REQ): {IsKnown: true},
				uint32(lnrpc.FeatureBit_AMP_OPT): {IsKnown: true},
			},
			expected: true,
		},
		{
			desc: "Knows only one",
			featureFlags: &[]lnrpc.FeatureBit{
				lnrpc.FeatureBit_AMP_REQ,
				lnrpc.FeatureBit_AMP_OPT,
			},
			features: map[uint32]*lnrpc.Feature{
				uint32(lnrpc.FeatureBit_AMP_OPT): {IsKnown: true},
			},
			expected: false,
		},
		{
			desc: "Unknown flags",
			featureFlags: &[]lnrpc.FeatureBit{
				lnrpc.FeatureBit_AMP_REQ,
				lnrpc.FeatureBit_AMP_OPT,
			},
			features: map[uint32]*lnrpc.Feature{
				uint32(lnrpc.FeatureBit_AMP_REQ): {IsKnown: false},
				uint32(lnrpc.FeatureBit_AMP_OPT): {IsKnown: false},
			},
			expected: false,
		},
		{
			desc:         "Empty flags",
			featureFlags: nil,
			expected:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			node := Node{
				FeatureFlags: tc.featureFlags,
			}

			actual := node.checkFeatureFlags(tc.features)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
