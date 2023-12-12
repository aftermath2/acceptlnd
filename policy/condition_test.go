package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	nodePublicKey := "node_public_key"
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
			desc: "Whitelist",
			conditions: &Conditions{
				Whitelist: &[]string{peerPublicKey},
			},
			req:      defaultReq,
			peer:     defaultPeer,
			expected: true,
		},
		{
			desc: "Blacklist",
			conditions: &Conditions{
				Blacklist: &[]string{peerPublicKey},
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
			actual := tc.conditions.Match(tc.req, nodePublicKey, tc.peer)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConditionsCheckWhitelist(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		desc      string
		publicKey string
		whitelist []string
		expected  bool
	}{
		{
			desc:      "Whitelisted",
			publicKey: publicKey,
			whitelist: []string{publicKey},
			expected:  true,
		},
		{
			desc:      "Not whitelisted",
			publicKey: "not key",
			whitelist: []string{publicKey},
			expected:  false,
		},
		{
			desc:      "Empty whitelist",
			whitelist: []string{},
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				Whitelist: &tc.whitelist,
			}

			actual := conditions.checkWhitelist(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		conditions := Conditions{}
		assert.False(t, conditions.checkWhitelist(""))
	})
}

func TestConditionsCheckBlacklist(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		desc      string
		publicKey string
		blacklist []string
		expected  bool
	}{
		{
			desc:      "Blacklisted",
			publicKey: publicKey,
			blacklist: []string{publicKey},
			expected:  false,
		},
		{
			desc:      "Not blacklisted",
			publicKey: "not key",
			blacklist: []string{publicKey},
			expected:  true,
		},
		{
			desc:      "Empty blacklist",
			blacklist: []string{},
			expected:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conditions := Conditions{
				Blacklist: &tc.blacklist,
			}

			actual := conditions.checkBlacklist(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		conditions := Conditions{}
		assert.True(t, conditions.checkBlacklist(""))
	})
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
