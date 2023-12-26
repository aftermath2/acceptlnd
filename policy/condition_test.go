package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	node := &lnrpc.GetInfoResponse{IdentityPubkey: "node_public_key"}
	peerPublicKey := "peer_public_key"
	defaultReq := &lnrpc.ChannelAcceptRequest{}
	defaultPeer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			PubKey: peerPublicKey,
		},
	}
	tru := true
	max := uint64(1)

	cases := []struct {
		conditions *Conditions
		req        *lnrpc.ChannelAcceptRequest
		peer       *lnrpc.NodeInfo
		desc       string
		expected   bool
	}{
		{
			desc:     "Nil conditions",
			req:      defaultReq,
			peer:     defaultPeer,
			expected: true,
		},
		{
			desc:       "Empty conditions",
			conditions: &Conditions{},
			req:        defaultReq,
			peer:       defaultPeer,
			expected:   true,
		},
		{
			desc: "Is",
			conditions: &Conditions{
				Is: &[]string{peerPublicKey},
			},
			req:      defaultReq,
			peer:     defaultPeer,
			expected: true,
		},
		{
			desc: "Is not",
			conditions: &Conditions{
				IsNot: &[]string{peerPublicKey},
			},
			req:      defaultReq,
			peer:     defaultPeer,
			expected: false,
		},
		{
			desc: "Is private",
			conditions: &Conditions{
				IsPrivate: &tru,
			},
			req: &lnrpc.ChannelAcceptRequest{
				ChannelFlags: uint32(lnwire.FFAnnounceChannel),
			},
			peer:     defaultPeer,
			expected: false,
		},
		{
			desc: "Wants zero conf",
			conditions: &Conditions{
				WantsZeroConf: &tru,
			},
			req: &lnrpc.ChannelAcceptRequest{
				WantsZeroConf: false,
			},
			peer:     defaultPeer,
			expected: false,
		},
		{
			desc: "Request",
			conditions: &Conditions{
				Request: &Request{
					ChannelCapacity: &Range[uint64]{
						Max: &max,
					},
				},
			},
			req: &lnrpc.ChannelAcceptRequest{
				FundingAmt: 10_000,
			},
			peer:     defaultPeer,
			expected: false,
		},

		{
			desc: "Node",
			conditions: &Conditions{
				Node: &Node{
					Hybrid: &tru,
				},
			},
			req: defaultReq,
			peer: &lnrpc.NodeInfo{
				Node: &lnrpc.LightningNode{
					PubKey: peerPublicKey,
					Addresses: []*lnrpc.NodeAddress{
						{Network: "tcp", Addr: "127.0.0.1:9735"},
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := tc.conditions.Match(tc.req, node, tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConditionsCheckIs(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		list      *[]string
		desc      string
		publicKey string
		expected  bool
	}{
		{
			desc:      "Is",
			publicKey: publicKey,
			list:      &[]string{publicKey},
			expected:  true,
		},
		{
			desc:      "Isn't",
			publicKey: "not key",
			list:      &[]string{publicKey},
			expected:  false,
		},
		{
			desc:     "Empty list",
			list:     &[]string{},
			expected: false,
		},
		{
			desc:     "Nil list",
			list:     nil,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				Is: tc.list,
			}

			actual := conditions.checkIs(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConditionsCheckIsNot(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		list      *[]string
		desc      string
		publicKey string
		expected  bool
	}{
		{
			desc:      "In list",
			publicKey: publicKey,
			list:      &[]string{publicKey},
			expected:  false,
		},
		{
			desc:      "Not in list",
			publicKey: "not key",
			list:      &[]string{publicKey},
			expected:  true,
		},
		{
			desc:     "Empty list",
			list:     &[]string{},
			expected: true,
		},
		{
			desc:     "Nil list",
			list:     nil,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				IsNot: tc.list,
			}

			actual := conditions.checkIsNot(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConditionsCheckIsPrivate(t *testing.T) {
	cases := []struct {
		desc      string
		isPrivate bool
		private   bool
		expected  bool
	}{
		{
			desc:      "Match",
			isPrivate: true,
			private:   true,
			expected:  true,
		},
		{
			desc:      "No match",
			isPrivate: false,
			private:   true,
			expected:  false,
		},
		{
			desc:      "No match 2",
			isPrivate: true,
			private:   false,
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				IsPrivate: &tc.isPrivate,
			}

			actual := conditions.checkIsPrivate(tc.private)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		conditions := Conditions{}
		actual := conditions.checkIsPrivate(true)
		assert.True(t, actual)
	})
}

func TestConditionsCheckWantsZeroConf(t *testing.T) {
	cases := []struct {
		desc          string
		wantsZeroConf bool
		wantZeroConf  bool
		expected      bool
	}{
		{
			desc:          "Match",
			wantsZeroConf: true,
			wantZeroConf:  true,
			expected:      true,
		},
		{
			desc:          "No match",
			wantsZeroConf: false,
			wantZeroConf:  true,
			expected:      false,
		},
		{
			desc:          "No match 2",
			wantsZeroConf: true,
			wantZeroConf:  false,
			expected:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				WantsZeroConf: &tc.wantsZeroConf,
			}

			actual := conditions.checkWantsZeroConf(tc.wantZeroConf)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		conditions := Conditions{}
		actual := conditions.checkWantsZeroConf(true)
		assert.True(t, actual)
	})
}
