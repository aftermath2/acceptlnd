package policy

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateRequest(t *testing.T) {
	max64 := uint64(1)
	max32 := uint32(1)

	cases := []struct {
		chanReq *lnrpc.ChannelAcceptRequest
		req     *Request
		desc    string
		fail    bool
	}{
		{
			desc: "Nil request",
			req:  nil,
			fail: false,
		},
		{
			desc:    "Empty request",
			req:     &Request{},
			chanReq: &lnrpc.ChannelAcceptRequest{},
			fail:    false,
		},
		{
			desc: "Channel capacity",
			req: &Request{
				ChannelCapacity: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				FundingAmt: 1000,
			},
			fail: true,
		},
		{
			desc: "Push amount",
			req: &Request{
				PushAmount: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				PushAmt: 1000,
			},
			fail: true,
		},
		{
			desc: "Channel reserve",
			req: &Request{
				ChannelReserve: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				ChannelReserve: 1000,
			},
			fail: true,
		},
		{
			desc: "CSV delay",
			req: &Request{
				CSVDelay: &Range[uint32]{
					Max: &max32,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				CsvDelay: 144,
			},
			fail: true,
		},
		{
			desc: "Max accepted HTLCs",
			req: &Request{
				MaxAcceptedHTLCs: &Range[uint32]{
					Max: &max32,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				MaxAcceptedHtlcs: 300,
			},
			fail: true,
		},
		{
			desc: "Min HTLC",
			req: &Request{
				MinHTLC: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				MinHtlc: 5,
			},
			fail: true,
		},
		{
			desc: "Max value in flight",
			req: &Request{
				MaxValueInFlight: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				MaxValueInFlight: 1000,
			},
			fail: true,
		},
		{
			desc: "Dust limit",
			req: &Request{
				DustLimit: &Range[uint64]{
					Max: &max64,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				DustLimit: 1000,
			},
			fail: true,
		},
		{
			desc: "Commitment type",
			req: &Request{
				CommitmentTypes: &[]lnrpc.CommitmentType{
					lnrpc.CommitmentType_ANCHORS,
				},
			},
			chanReq: &lnrpc.ChannelAcceptRequest{
				CommitmentType: lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE,
			},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.req.evaluate(tc.chanReq)
			if tc.fail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCheckCommitmentType(t *testing.T) {
	cases := []struct {
		desc            string
		commitmentTypes []lnrpc.CommitmentType
		commitmentType  lnrpc.CommitmentType
		expected        bool
	}{
		{
			desc:           "Accept",
			commitmentType: lnrpc.CommitmentType_ANCHORS,
			commitmentTypes: []lnrpc.CommitmentType{
				lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE,
				lnrpc.CommitmentType_ANCHORS,
			},
			expected: true,
		},
		{
			desc:           "Reject",
			commitmentType: lnrpc.CommitmentType_LEGACY,
			commitmentTypes: []lnrpc.CommitmentType{
				lnrpc.CommitmentType_SCRIPT_ENFORCED_LEASE,
				lnrpc.CommitmentType_ANCHORS,
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			r := Request{
				CommitmentTypes: &tc.commitmentTypes,
			}

			actual := r.checkCommitmentType(tc.commitmentType)
			assert.Equal(t, tc.expected, actual)
		})
	}

	t.Run("Nil", func(t *testing.T) {
		r := Request{}
		assert.True(t, r.checkCommitmentType(lnrpc.CommitmentType_ANCHORS))
	})
}
