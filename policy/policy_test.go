package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluatePolicy(t *testing.T) {
	node := &lnrpc.GetInfoResponse{IdentityPubkey: "node_public_key"}
	peerPublicKey := "peer_public_key"
	defaultReq := &lnrpc.ChannelAcceptRequest{}
	defaultPeer := &lnrpc.NodeInfo{
		Node: &lnrpc.LightningNode{
			PubKey: peerPublicKey,
		},
	}
	tru := true
	fals := false
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
					Is: &[]string{peerPublicKey},
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
					IsNot: &[]string{peerPublicKey},
				},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: false,
		},
		{
			desc: "Allow list",
			policy: Policy{
				AllowList: &[]string{"other_public_key"},
			},
			req:  defaultReq,
			peer: defaultPeer,
			fail: true,
		},
		{
			desc: "Block list",
			policy: Policy{
				BlockList: &[]string{peerPublicKey},
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
			desc: "Accept wants zero conf",
			policy: Policy{
				AcceptZeroConfChannels: &fals,
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
			err := tc.policy.Evaluate(tc.req, node, tc.peer)
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

func TestCheckAllowList(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		list      *[]string
		desc      string
		publicKey string
		expected  bool
	}{
		{
			desc:      "Allowed",
			publicKey: publicKey,
			list:      &[]string{publicKey},
			expected:  true,
		},
		{
			desc:      "Not allowed",
			publicKey: "not key",
			list:      &[]string{publicKey},
			expected:  false,
		},
		{
			desc:     "Empty list",
			list:     nil,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				AllowList: tc.list,
			}

			actual := policy.checkAllowList(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCheckBlockList(t *testing.T) {
	publicKey := "key"

	cases := []struct {
		list      *[]string
		desc      string
		publicKey string
		expected  bool
	}{
		{
			desc:      "Blocked",
			publicKey: publicKey,
			list:      &[]string{publicKey},
			expected:  false,
		},
		{
			desc:      "Not blocked",
			publicKey: "not key",
			list:      &[]string{publicKey},
			expected:  true,
		},
		{
			desc:     "Nil",
			list:     nil,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				BlockList: tc.list,
			}

			actual := policy.checkBlockList(tc.publicKey)
			assert.Equal(t, tc.expected, actual)
		})
	}
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
	publicKey := "public_key"

	cases := []struct {
		zeroConfList   *[]string
		desc           string
		publicKey      string
		acceptZeroConf bool
		wantsZeroConf  bool
		expected       bool
	}{
		{
			desc:          "No zero conf",
			wantsZeroConf: false,
			expected:      true,
		},
		{
			desc:           "Accept all",
			acceptZeroConf: true,
			wantsZeroConf:  true,
			expected:       true,
		},
		{
			desc:           "Accept in list",
			publicKey:      publicKey,
			zeroConfList:   &[]string{publicKey},
			acceptZeroConf: true,
			wantsZeroConf:  true,
			expected:       true,
		},
		{
			desc:           "Reject all",
			acceptZeroConf: false,
			wantsZeroConf:  true,
			expected:       false,
		},
		{
			desc:           "Reject even if in list",
			publicKey:      publicKey,
			zeroConfList:   &[]string{publicKey},
			acceptZeroConf: false,
			wantsZeroConf:  true,
			expected:       false,
		},
		{
			desc:           "Reject not in list",
			publicKey:      publicKey,
			zeroConfList:   &[]string{"other_public_key"},
			acceptZeroConf: true,
			wantsZeroConf:  true,
			expected:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			policy := Policy{
				AcceptZeroConfChannels: &tc.acceptZeroConf,
				ZeroConfList:           tc.zeroConfList,
			}

			actual := policy.checkZeroConf(tc.publicKey, tc.wantsZeroConf)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
