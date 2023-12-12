package policy

import (
	"errors"
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Request represents the desired values in a channel request.
type Request struct {
	ChannelCapacity  *Range[uint64]          `yaml:"channel_capacity,omitempty"`
	ChannelReserve   *Range[uint64]          `yaml:"channel_reserve,omitempty"`
	CSVDelay         *Range[uint32]          `yaml:"csv_delay,omitempty"`
	PushAmount       *Range[uint64]          `yaml:"push_amount,omitempty"`
	MaxAcceptedHTLCs *Range[uint32]          `yaml:"max_accepted_htlcs,omitempty"`
	MinHTLC          *Range[uint64]          `yaml:"min_htlc,omitempty"`
	MaxValueInFlight *Range[uint64]          `yaml:"max_value_in_flight,omitempty"`
	DustLimit        *Range[uint64]          `yaml:"dust_limit,omitempty"`
	CommitmentTypes  *[]lnrpc.CommitmentType `yaml:"commitment_types,omitempty"`
}

func (r *Request) evaluate(req *lnrpc.ChannelAcceptRequest) error {
	if r == nil {
		return nil
	}

	if !check(r.ChannelCapacity, req.FundingAmt) {
		return errors.New("Channel capacity " + r.ChannelCapacity.Reason())
	}

	if !check(r.PushAmount, req.PushAmt) {
		return errors.New("Pushed amount lower than expected")
	}

	if !check(r.ChannelReserve, req.ChannelReserve) {
		return errors.New("Channel reserve " + r.ChannelReserve.Reason())
	}

	if !check(r.CSVDelay, req.CsvDelay) {
		return errors.New("Check sequence verify delay " + r.CSVDelay.Reason())
	}

	if !check(r.MaxAcceptedHTLCs, req.MaxAcceptedHtlcs) {
		return errors.New("Maximum accepted HTLCs " + r.MaxAcceptedHTLCs.Reason())
	}

	if !check(r.MinHTLC, req.MinHtlc) {
		return errors.New("Minimum HTLCs " + r.MinHTLC.Reason())
	}

	if !check(r.MaxValueInFlight, req.MaxValueInFlight) {
		return errors.New("Maximum value in flight " + r.MaxValueInFlight.Reason())
	}

	if !check(r.DustLimit, req.DustLimit) {
		return errors.New("Commitment transaction dust limit " + r.DustLimit.Reason())
	}

	if !r.checkCommitmentType(req.CommitmentType) {
		return fmt.Errorf("Commitment type is not in %s", *r.CommitmentTypes)
	}

	return nil
}

func (r *Request) checkCommitmentType(commitmentType lnrpc.CommitmentType) bool {
	if r.CommitmentTypes == nil {
		return true
	}
	for _, ct := range *r.CommitmentTypes {
		if ct == commitmentType {
			return true
		}
	}
	return false
}
