package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateChannels(t *testing.T) {
	nodePublicKey := "node_public_key"
	peerPublicKey := "peer_public_key"
	maxu32 := uint32(1)
	max32 := int32(1)
	max64 := int64(1)
	maxu64 := uint64(1)
	max := 1
	maxFloat := float64(0.5)
	tru := true

	cases := []struct {
		channels *Channels
		peer     *lnrpc.NodeInfo
		desc     string
		fail     bool
	}{
		{
			desc:     "Nil channels",
			channels: nil,
		},
		{
			desc:     "Nil peers",
			channels: &Channels{Peers: nil},
			peer:     &lnrpc.NodeInfo{},
		},
		{
			desc:     "Empty channels and peers",
			channels: &Channels{Peers: &Peers{}},
			peer:     &lnrpc.NodeInfo{},
		},
		{
			desc: "Number of channels",
			channels: &Channels{
				Number: &Range[uint32]{
					Max: &maxu32,
				},
			},
			peer: &lnrpc.NodeInfo{
				NumChannels: 2,
			},
			fail: true,
		},
		{
			desc: "Capacity",
			channels: &Channels{
				Capacity: &StatRange[int64]{
					Max: &max64,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Capacity: 1_000_000,
						Node1Pub: peerPublicKey,
					},
				},
			},
			fail: true,
		},
		{
			desc: "Zero base fees",
			channels: &Channels{
				ZeroBaseFees: &tru,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							FeeBaseMsat: 1000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Block height",
			channels: &Channels{
				BlockHeight: &StatRange[uint32]{
					Max: &maxu32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:  peerPublicKey,
						ChannelId: 623702369048395776,
					},
				},
			},
			fail: true,
		},
		{
			desc: "Time lock delta",
			channels: &Channels{
				TimeLockDelta: &StatRange[uint32]{
					Max: &maxu32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							TimeLockDelta: 90,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Minimum HTLC",
			channels: &Channels{
				MinHTLC: &StatRange[int64]{
					Max: &max64,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							MinHtlc: 2,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Maximum HTLC",
			channels: &Channels{
				MaxHTLC: &StatRange[uint64]{
					Max: &maxu64,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							MaxHtlcMsat: 2000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Last update difference",
			channels: &Channels{
				LastUpdateDiff: &StatRange[uint32]{
					Max: &maxu32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							LastUpdate: 14_230_110,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Together",
			channels: &Channels{
				Together: &Range[int]{
					Max: &max,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Pub: nodePublicKey,
					},
					{
						Node1Pub: peerPublicKey,
						Node2Pub: nodePublicKey,
					},
				},
			},
			fail: true,
		},
		{
			desc: "Peers fee rates",
			channels: &Channels{
				Peers: &Peers{
					FeeRates: &StatRange[int64]{
						Max: &max64,
					},
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Policy: &lnrpc.RoutingPolicy{
							FeeRateMilliMsat: 10000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Fee rates",
			channels: &Channels{
				FeeRates: &StatRange[int64]{
					Max: &max64,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							FeeRateMilliMsat: 10000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Peers base fees",
			channels: &Channels{
				Peers: &Peers{
					BaseFees: &StatRange[int64]{
						Max: &max64,
					},
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Policy: &lnrpc.RoutingPolicy{
							FeeBaseMsat: 2000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Base fees",
			channels: &Channels{
				BaseFees: &StatRange[int64]{
					Max: &max64,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							FeeBaseMsat: 2000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Peers disabled",
			channels: &Channels{
				Peers: &Peers{
					Disabled: &StatRange[float64]{
						Max: &maxFloat,
					},
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Policy: &lnrpc.RoutingPolicy{
							Disabled: true,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Disabled",
			channels: &Channels{
				Disabled: &StatRange[float64]{
					Max: &maxFloat,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							Disabled: true,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Peers inbound fee rates",
			channels: &Channels{
				Peers: &Peers{
					InboundFeeRates: &StatRange[int32]{
						Max: &max32,
					},
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Policy: &lnrpc.RoutingPolicy{
							InboundFeeRateMilliMsat: 10000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Inbound fee rates",
			channels: &Channels{
				InboundFeeRates: &StatRange[int32]{
					Max: &max32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							InboundFeeRateMilliMsat: 10000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Peers inbound base fees",
			channels: &Channels{
				Peers: &Peers{
					InboundBaseFees: &StatRange[int32]{
						Max: &max32,
					},
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node2Policy: &lnrpc.RoutingPolicy{
							InboundFeeBaseMsat: 2000,
						},
					},
				},
			},
			fail: true,
		},
		{
			desc: "Inbound base fees",
			channels: &Channels{
				InboundBaseFees: &StatRange[int32]{
					Max: &max32,
				},
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							InboundFeeBaseMsat: 2000,
						},
					},
				},
			},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.channels.evaluate(nodePublicKey, tc.peer)
			if tc.fail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCheckCapacity(t *testing.T) {
	min := int64(100_000)
	max := int64(1_000_000)

	cases := []struct {
		capacity *StatRange[int64]
		desc     string
		channels []*lnrpc.ChannelEdge
		expected bool
	}{
		{
			desc: "Contains",
			capacity: &StatRange[int64]{
				Operation: Mean,
				Min:       &min,
				Max:       &max,
			},
			channels: []*lnrpc.ChannelEdge{
				{Capacity: 10_000},
				{Capacity: 250_000},
			},
			expected: true,
		},
		{
			desc: "Does not contain",
			capacity: &StatRange[int64]{
				Operation: Mean,
				Min:       &min,
				Max:       &max,
			},
			channels: []*lnrpc.ChannelEdge{
				{Capacity: 50_000},
				{Capacity: 25_000},
			},
			expected: false,
		},
		{
			desc:     "Nil",
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			channels := Channels{
				Capacity: tc.capacity,
			}

			actual := checkStat(
				channels.Capacity,
				&lnrpc.NodeInfo{Channels: tc.channels},
				capacityFunc,
			)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCheckZeroBaseFees(t *testing.T) {
	publicKey := "public_key"

	cases := []struct {
		peer         *lnrpc.NodeInfo
		desc         string
		zeroBaseFees bool
		expected     bool
	}{
		{
			desc:         "Zero base fee channels",
			zeroBaseFees: true,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{PubKey: publicKey},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: publicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							FeeBaseMsat: 0,
						},
					},
					{
						Node1Pub: publicKey,
						Node1Policy: &lnrpc.RoutingPolicy{
							FeeBaseMsat: 0,
						},
					},
				},
			},
			expected: true,
		},
		{
			desc:         "Non zero base fee channels",
			zeroBaseFees: true,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{PubKey: publicKey},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:    publicKey,
						Node1Policy: &lnrpc.RoutingPolicy{FeeBaseMsat: 0},
					},
					{
						Node1Pub:    publicKey,
						Node1Policy: &lnrpc.RoutingPolicy{FeeBaseMsat: 1},
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			channels := Channels{
				ZeroBaseFees: &tc.zeroBaseFees,
			}

			actual := channels.checkZeroBaseFees(tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		channels := Channels{}
		actual := channels.checkZeroBaseFees(nil)
		assert.True(t, actual)
	})
}

func TestCheckTogether(t *testing.T) {
	nodePublicKey := "node_public_key"
	peerPublicKey := "peer_public_key"
	min, max := 1, 3

	cases := []struct {
		peer          *lnrpc.NodeInfo
		together      *Range[int]
		desc          string
		nodePublicKey string
		expected      bool
	}{
		{
			desc:          "Channels together",
			nodePublicKey: nodePublicKey,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: nodePublicKey,
						Node2Pub: peerPublicKey,
					},
					{
						Node1Pub: peerPublicKey,
						Node2Pub: nodePublicKey,
					},
				},
			},
			together: &Range[int]{
				Min: &min,
				Max: &max,
			},
			expected: true,
		},
		{
			desc:          "Not enough channels together",
			nodePublicKey: nodePublicKey,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: nodePublicKey,
						Node2Pub: peerPublicKey,
					},
					{
						Node1Pub: peerPublicKey,
						Node2Pub: nodePublicKey,
					},
				},
			},
			together: &Range[int]{
				Min: &max,
			},
			expected: false,
		},
		{
			desc:          "No channels together",
			nodePublicKey: nodePublicKey,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub: nodePublicKey + "d21u",
						Node2Pub: peerPublicKey,
					},
					{
						Node1Pub: peerPublicKey + "d21u",
						Node2Pub: nodePublicKey,
					},
				},
			},
			together: &Range[int]{
				Min: &min,
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			channels := Channels{
				Together: tc.together,
			}

			actual := channels.checkTogether(tc.nodePublicKey, tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		channels := Channels{}
		assert.True(t, channels.checkTogether("", nil))
	})
}

func TestCheckPeersDisabled(t *testing.T) {
	peerPublicKey := "peer_public_key"
	value := 0.6

	cases := []struct {
		peer          *lnrpc.NodeInfo
		peersDisabled *StatRange[float64]
		desc          string
		expected      bool
	}{
		{
			desc: "Maximum disabled channels rate met",
			peersDisabled: &StatRange[float64]{
				Operation: Mean,
				Max:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
				},
			},
			expected: true,
		},
		{
			desc: "Maximum disabled channels rate not met",
			peersDisabled: &StatRange[float64]{
				Operation: Mean,
				Max:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
				},
			},
			expected: false,
		},
		{
			desc: "Minimum disabled channels rate met",
			peersDisabled: &StatRange[float64]{
				Operation: Mean,
				Min:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
				},
			},
			expected: true,
		},
		{
			desc: "Minimum disabled channels rate not met",
			peersDisabled: &StatRange[float64]{
				Operation: Mean,
				Min:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: true}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
					{Node1Pub: peerPublicKey, Node2Policy: &lnrpc.RoutingPolicy{Disabled: false}},
				},
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			channels := Channels{
				Peers: &Peers{
					Disabled: tc.peersDisabled,
				},
			}

			actual := channels.checkPeersDisabled(tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		channels := Channels{Peers: &Peers{}}
		assert.True(t, channels.checkPeersDisabled(nil))
	})
}

func TestCheckDisabled(t *testing.T) {
	value := 0.6
	peerPublicKey := "peer_public_key"

	cases := []struct {
		peer     *lnrpc.NodeInfo
		disabled *StatRange[float64]
		desc     string
		expected bool
	}{
		{
			desc: "Maximum disabled channels rate met",
			disabled: &StatRange[float64]{
				Operation: Mean,
				Max:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
				},
			},
			expected: true,
		},
		{
			desc: "Maximum disabled channels rate not met",
			disabled: &StatRange[float64]{
				Operation: Mean,
				Max:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
				},
			},
			expected: false,
		},
		{
			desc: "Minimum disabled channels rate met",
			disabled: &StatRange[float64]{
				Operation: Mean,
				Min:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
				},
			},
			expected: true,
		},
		{
			desc: "Minimum disabled channels rate not met",
			disabled: &StatRange[float64]{
				Operation: Mean,
				Min:       &value,
			},
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
				},
				Channels: []*lnrpc.ChannelEdge{
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: true},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
					{
						Node1Pub:    peerPublicKey,
						Node1Policy: &lnrpc.RoutingPolicy{Disabled: false},
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			channels := Channels{
				Disabled: tc.disabled,
			}

			actual := channels.checkDisabled(tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		channels := Channels{}
		assert.True(t, channels.checkDisabled(nil))
	})
}

func TestGetNodePolicy(t *testing.T) {
	publicKey := "public_key"
	expectedPolicy := &lnrpc.RoutingPolicy{
		TimeLockDelta: 1,
	}
	otherPolicy := &lnrpc.RoutingPolicy{
		TimeLockDelta: 5,
	}

	cases := []struct {
		peerPublicKey string
		channel       *lnrpc.ChannelEdge
		expected      *lnrpc.RoutingPolicy
		desc          string
		outgoing      bool
	}{
		{
			desc:          "Get peers node policy",
			peerPublicKey: publicKey,
			channel: &lnrpc.ChannelEdge{
				Node1Policy: expectedPolicy,
				Node2Pub:    publicKey,
				Node2Policy: otherPolicy,
			},
			outgoing: false,
		},
		{
			desc:          "Get peers node policy 2",
			peerPublicKey: publicKey,
			channel: &lnrpc.ChannelEdge{
				Node1Pub:    publicKey,
				Node1Policy: otherPolicy,
				Node2Policy: expectedPolicy,
			},
			outgoing: false,
		},
		{
			desc:          "Get node policy",
			peerPublicKey: publicKey,
			channel: &lnrpc.ChannelEdge{
				Node1Pub:    publicKey,
				Node1Policy: expectedPolicy,
				Node2Policy: otherPolicy,
			},
			outgoing: true,
		},
		{
			desc:          "Get node policy 2",
			peerPublicKey: publicKey,
			channel: &lnrpc.ChannelEdge{
				Node1Policy: otherPolicy,
				Node2Pub:    publicKey,
				Node2Policy: expectedPolicy,
			},
			outgoing: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := getNodePolicy(tc.peerPublicKey, tc.channel, tc.outgoing)
			assert.Equal(t, expectedPolicy, actual)
		})
	}
}

func TestBlockHeightFunc(t *testing.T) {
	channel := &lnrpc.ChannelEdge{
		ChannelId: 623702369048395776,
	}
	expected := uint32(567254)
	actual := blockHeightFunc(nil, channel)
	assert.Equal(t, expected, actual)
}

func TestTimeLockDeltaFunc(t *testing.T) {
	publicKey := "public_key"
	expected := uint32(5)
	peer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{PubKey: publicKey},
	}
	channel := &lnrpc.ChannelEdge{
		Node1Pub:    publicKey,
		Node1Policy: &lnrpc.RoutingPolicy{TimeLockDelta: expected},
	}
	actual := timeLockDeltaFunc()(peer, channel)
	assert.Equal(t, expected, actual)
}

func TestMinHTLCFunc(t *testing.T) {
	publicKey := "public_key"
	expected := int64(1)
	peer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{PubKey: publicKey},
	}
	channel := &lnrpc.ChannelEdge{
		Node1Pub:    publicKey,
		Node1Policy: &lnrpc.RoutingPolicy{MinHtlc: expected},
	}
	actual := minHTLCFunc()(peer, channel)
	assert.Equal(t, expected, actual)
}

func TestMaxHTLCFunc(t *testing.T) {
	publicKey := "public_key"
	expected := uint64(90000000)
	peer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{PubKey: publicKey},
	}
	channel := &lnrpc.ChannelEdge{
		Node1Pub:    publicKey,
		Node1Policy: &lnrpc.RoutingPolicy{MaxHtlcMsat: expected * 1000},
	}
	actual := maxHTLCFunc()(peer, channel)
	assert.Equal(t, expected, actual)
}

func TestLastUpdateFunc(t *testing.T) {
	publicKey := "public_key"
	expected := uint32(500)
	peer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{PubKey: publicKey},
	}
	channel := &lnrpc.ChannelEdge{
		Node1Pub:    publicKey,
		Node1Policy: &lnrpc.RoutingPolicy{LastUpdate: expected},
	}
	actual := lastUpdateFunc(1000)(peer, channel)
	assert.Equal(t, expected, actual)
}

func TestFeeRatesFunc(t *testing.T) {
	t.Run("Peers", func(t *testing.T) {
		expected := int64(1)
		peer := &lnrpc.NodeInfo{Node: &lnrpc.LightningNode{}}
		channel := &lnrpc.ChannelEdge{
			Node2Pub:    "pub",
			Node2Policy: &lnrpc.RoutingPolicy{FeeRateMilliMsat: expected * 1000},
		}
		actual := feeRatesFunc(false)(peer, channel)
		assert.Equal(t, expected, actual)
	})

	t.Run("Outgoing", func(t *testing.T) {
		publicKey := "public_key"
		expected := int64(5)
		peer := &lnrpc.NodeInfo{
			Node: &lnrpc.LightningNode{PubKey: publicKey},
		}
		channel := &lnrpc.ChannelEdge{
			Node1Pub:    publicKey,
			Node1Policy: &lnrpc.RoutingPolicy{FeeRateMilliMsat: expected * 1000},
		}
		actual := feeRatesFunc(true)(peer, channel)
		assert.Equal(t, expected, actual)
	})
}

func TestBaseFeesFunc(t *testing.T) {
	t.Run("Peers", func(t *testing.T) {
		expected := int64(3)
		peer := &lnrpc.NodeInfo{Node: &lnrpc.LightningNode{}}
		channel := &lnrpc.ChannelEdge{
			Node2Pub:    "pub",
			Node2Policy: &lnrpc.RoutingPolicy{FeeBaseMsat: expected * 1000},
		}
		actual := baseFeesFunc(false)(peer, channel)
		assert.Equal(t, expected, actual)
	})

	t.Run("Outgoing", func(t *testing.T) {
		publicKey := "public_key"
		expected := int64(1)
		peer := &lnrpc.NodeInfo{
			Node: &lnrpc.LightningNode{PubKey: publicKey},
		}
		channel := &lnrpc.ChannelEdge{
			Node1Pub:    publicKey,
			Node1Policy: &lnrpc.RoutingPolicy{FeeBaseMsat: expected * 1000},
		}
		actual := baseFeesFunc(true)(peer, channel)
		assert.Equal(t, expected, actual)
	})
}

func TestInboundFeeRatesFunc(t *testing.T) {
	t.Run("Peers", func(t *testing.T) {
		expected := int32(1)
		peer := &lnrpc.NodeInfo{Node: &lnrpc.LightningNode{}}
		channel := &lnrpc.ChannelEdge{
			Node2Pub:    "pub",
			Node2Policy: &lnrpc.RoutingPolicy{InboundFeeRateMilliMsat: expected * 1000},
		}
		actual := inboundFeeRatesFunc(false)(peer, channel)
		assert.Equal(t, expected, actual)
	})

	t.Run("Outgoing", func(t *testing.T) {
		publicKey := "public_key"
		expected := int32(5)
		peer := &lnrpc.NodeInfo{
			Node: &lnrpc.LightningNode{PubKey: publicKey},
		}
		channel := &lnrpc.ChannelEdge{
			Node1Pub:    publicKey,
			Node1Policy: &lnrpc.RoutingPolicy{InboundFeeRateMilliMsat: expected * 1000},
		}
		actual := inboundFeeRatesFunc(true)(peer, channel)
		assert.Equal(t, expected, actual)
	})
}

func TestInboundBaseFeesFunc(t *testing.T) {
	t.Run("Peers", func(t *testing.T) {
		expected := int32(1)
		peer := &lnrpc.NodeInfo{Node: &lnrpc.LightningNode{}}
		channel := &lnrpc.ChannelEdge{
			Node2Pub:    "pub",
			Node2Policy: &lnrpc.RoutingPolicy{InboundFeeBaseMsat: expected * 1000},
		}
		actual := inboundBaseFeesFunc(false)(peer, channel)
		assert.Equal(t, expected, actual)
	})

	t.Run("Outgoing", func(t *testing.T) {
		publicKey := "public_key"
		expected := int32(1)
		peer := &lnrpc.NodeInfo{
			Node: &lnrpc.LightningNode{PubKey: publicKey},
		}
		channel := &lnrpc.ChannelEdge{
			Node1Pub:    publicKey,
			Node1Policy: &lnrpc.RoutingPolicy{InboundFeeBaseMsat: expected * 1000},
		}
		actual := inboundBaseFeesFunc(true)(peer, channel)
		assert.Equal(t, expected, actual)
	})
}
