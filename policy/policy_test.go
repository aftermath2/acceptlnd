package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluatePolicy(t *testing.T) {
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
		policy Policy
		req    *lnrpc.ChannelAcceptRequest
		peer   *lnrpc.NodeInfo
		desc   string
		fail   bool
	}{
		{
			desc:   "No policy",
			policy: Policy{},
			req:    defaultReq,
			peer:   defaultPeer,
			fail:   false,
		},
		{
			desc: "Conditions match",
			policy: Policy{
				Conditions: &Conditions{
					Whitelist: &[]string{peerPublicKey},
				},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "No conditions match",
			policy: Policy{
				Conditions: &Conditions{
					Blacklist: &[]string{peerPublicKey},
				},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "Whitelist",
			policy: Policy{
				Whitelist: &[]string{peerPublicKey},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "Blacklist",
			policy: Policy{
				Blacklist: &[]string{peerPublicKey},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Reject all",
			policy: Policy{
				RejectAll: &tru,
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Reject private channels",
			policy: Policy{
				RejectPrivateChannels: &tru,
			},
			req: &lnrpc.ChannelAcceptRequest{
				ChannelFlags: 0,
			},
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Reject wants zero conf",
			policy: Policy{
				RejectZeroConfChannels: &tru,
			},
			req: &lnrpc.ChannelAcceptRequest{
				WantsZeroConf: true,
			},
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Request",
			policy: Policy{
				Request: &Request{
					ChannelCapacity: &Range[uint64]{
						Max: &max,
					},
				},
			},
			req: &lnrpc.ChannelAcceptRequest{
				FundingAmt: 10_000,
			},
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Node",
			policy: Policy{
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
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.policy.Evaluate(tc.req, nodePublicKey, tc.peer)
			if tc.fail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCheckRejectAll(t *testing.T) {
	cases := []struct {
		desc      string
		rejectAll bool
		expected  bool
	}{
		{
			desc:      "Reject all",
			rejectAll: true,
			expected:  false,
		},
		{
			desc:      "Do not reject all",
			rejectAll: false,
			expected:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				RejectAll: &tc.rejectAll,
			}

			actual := policy.checkRejectAll()
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		policy := Policy{}
		assert.True(t, policy.checkRejectAll())
	})
}

func TestCheckWhitelist(t *testing.T) {
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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				Whitelist: &tc.whitelist,
			}

			actual := policy.checkWhitelist(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		policy := Policy{}
		assert.False(t, policy.checkWhitelist(""))
	})
}

func TestCheckBlacklist(t *testing.T) {
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
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				Blacklist: &tc.blacklist,
			}

			actual := policy.checkBlacklist(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		policy := Policy{}
		assert.True(t, policy.checkBlacklist(""))
	})
}

func TestCheckPrivate(t *testing.T) {
	cases := []struct {
		desc          string
		rejectPrivate bool
		private       bool
		expected      bool
	}{
		{
			desc:          "Reject",
			rejectPrivate: true,
			private:       true,
			expected:      false,
		},
		{
			desc:          "Reject 2",
			rejectPrivate: true,
			private:       false,
			expected:      true,
		},
		{
			desc:          "Accept",
			rejectPrivate: false,
			private:       true,
			expected:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				RejectPrivateChannels: &tc.rejectPrivate,
			}

			actual := policy.checkPrivate(tc.private)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Empty", func(t *testing.T) {
		policy := Policy{}
		actual := policy.checkPrivate(true)
		assert.True(t, actual)
	})
}

func TestCheckZeroConf(t *testing.T) {
	cases := []struct {
		desc           string
		rejectZeroConf bool
		wantsZeroConf  bool
		expected       bool
	}{
		{
			desc:           "Reject",
			rejectZeroConf: true,
			wantsZeroConf:  true,
			expected:       false,
		},
		{
			desc:           "Accept",
			rejectZeroConf: true,
			wantsZeroConf:  false,
			expected:       true,
		},
		{
			desc:           "Accept 2",
			rejectZeroConf: false,
			wantsZeroConf:  true,
			expected:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				RejectZeroConfChannels: &tc.rejectZeroConf,
			}

			actual := policy.checkZeroConf(tc.wantsZeroConf)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Empty", func(t *testing.T) {
		policy := Policy{}
		actual := policy.checkZeroConf(true)
		assert.True(t, actual)
	})
}
