package policy

import (
	"errors"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Channels represents a set of requirements that the initiator's node channels must satisfy.
type Channels struct {
	Number                  *Range[uint32]      `yaml:"number,omitempty"`
	Capacity                *StatRange[int64]   `yaml:"capacity,omitempty"`
	ZeroBaseFees            *bool               `yaml:"zero_base_fees,omitempty"`
	BlockHeight             *StatRange[uint32]  `yaml:"block_height,omitempty"`
	TimeLockDelta           *StatRange[uint32]  `yaml:"time_lock_delta,omitempty"`
	MinHTLC                 *StatRange[int64]   `yaml:"min_htlc,omitempty"`
	MaxHTLC                 *StatRange[uint64]  `yaml:"max_htlc,omitempty"`
	LastUpdateDiff          *StatRange[uint32]  `yaml:"last_update_diff,omitempty"`
	Together                *Range[int]         `yaml:"together,omitempty"`
	IncomingFeeRates        *StatRange[int64]   `yaml:"incoming_fee_rates,omitempty"`
	OutgoingFeeRates        *StatRange[int64]   `yaml:"outgoing_fee_rates,omitempty"`
	IncomingBaseFees        *StatRange[int64]   `yaml:"incoming_base_fees,omitempty"`
	OutgoingBaseFees        *StatRange[int64]   `yaml:"outgoing_base_fees,omitempty"`
	IncomingDisabled        *StatRange[float64] `yaml:"incoming_disabled,omitempty"`
	OutgoingDisabled        *StatRange[float64] `yaml:"outgoing_disabled,omitempty"`
	IncomingInboundFeeRates *StatRange[int32]   `yaml:"incoming_inbound_fees_rates,omitempty"`
	OutgoingInboundFeeRates *StatRange[int32]   `yaml:"outgoing_inbound_fees_rates,omitempty"`
	IncomingInboundBaseFees *StatRange[int32]   `yaml:"incoming_inbound_base_fees,omitempty"`
	OutgoingInboundBaseFees *StatRange[int32]   `yaml:"outgoing_inbound_base_fees,omitempty"`
}

func (c *Channels) evaluate(nodePublicKey string, peer *lnrpc.NodeInfo) error {
	if c == nil {
		return nil
	}

	if !check(c.Number, peer.NumChannels) {
		return errors.New("Node number of channels " + c.Number.Reason())
	}

	if !checkStat(c.Capacity, peer, capacityFunc) {
		return errors.New("Capacity " + c.Capacity.Reason())
	}

	if !c.checkZeroBaseFees(peer) {
		return errors.New("Node has channels with base fees higher than zero")
	}

	if !checkStat(c.BlockHeight, peer, blockHeightFunc) {
		return errors.New("Block height " + c.BlockHeight.Reason())
	}

	if !checkStat(c.TimeLockDelta, peer, timeLockDeltaFunc(peer)) {
		return errors.New("Time lock delta " + c.TimeLockDelta.Reason())
	}

	if !checkStat(c.MinHTLC, peer, minHTLCFunc(peer)) {
		return errors.New("Channels minimum HTLC " + c.MinHTLC.Reason())
	}

	if !checkStat(c.MaxHTLC, peer, maxHTLCFunc(peer)) {
		return errors.New("Channels maximum HTLC " + c.MaxHTLC.Reason())
	}

	if !checkStat(c.LastUpdateDiff, peer, lastUpdateFunc(peer, time.Now().Unix())) {
		return errors.New("Channels last update " + c.LastUpdateDiff.Reason())
	}

	if !c.checkTogether(nodePublicKey, peer) {
		return errors.New("Channels together " + c.Together.Reason())
	}

	if !checkStat(c.IncomingFeeRates, peer, feeRatesFunc(peer, false)) {
		return errors.New("Incoming fee rates " + c.IncomingFeeRates.Reason())
	}

	if !checkStat(c.OutgoingFeeRates, peer, feeRatesFunc(peer, true)) {
		return errors.New("Outgoing fee rates " + c.OutgoingFeeRates.Reason())
	}

	if !checkStat(c.IncomingBaseFees, peer, baseFeesFunc(peer, false)) {
		return errors.New("Incoming base fees " + c.IncomingBaseFees.Reason())
	}

	if !checkStat(c.OutgoingBaseFees, peer, baseFeesFunc(peer, true)) {
		return errors.New("Outgoing base fees " + c.OutgoingBaseFees.Reason())
	}

	if !checkStat(c.IncomingInboundFeeRates, peer, inboundFeeRatesFunc(peer, false)) {
		return errors.New("Incoming inbound fee rates " + c.IncomingInboundFeeRates.Reason())
	}

	if !checkStat(c.OutgoingInboundFeeRates, peer, inboundFeeRatesFunc(peer, true)) {
		return errors.New("Outgoing inbound fee rates " + c.OutgoingInboundFeeRates.Reason())
	}

	if !checkStat(c.IncomingInboundBaseFees, peer, inboundBaseFeesFunc(peer, false)) {
		return errors.New("Incoming inbound base fees " + c.IncomingInboundBaseFees.Reason())
	}

	if !checkStat(c.OutgoingInboundBaseFees, peer, inboundBaseFeesFunc(peer, true)) {
		return errors.New("Outgoing inbound base fees " + c.OutgoingInboundBaseFees.Reason())
	}

	if !c.checkIncomingDisabled(peer) {
		return errors.New("Incoming disabled channels " + c.IncomingDisabled.Reason())
	}

	if !c.checkOutgoingDisabled(peer) {
		return errors.New("Outgoing disabled channels " + c.OutgoingDisabled.Reason())
	}

	return nil
}

func (c *Channels) checkZeroBaseFees(peer *lnrpc.NodeInfo) bool {
	if c.ZeroBaseFees == nil {
		return true
	}

	for _, channel := range peer.Channels {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		if policy.FeeBaseMsat != 0 {
			return false
		}
	}
	return true
}

func (c *Channels) checkTogether(nodePublicKey string, peer *lnrpc.NodeInfo) bool {
	if c.Together == nil {
		return true
	}

	count := 0
	for _, channel := range peer.Channels {
		if (nodePublicKey == channel.Node1Pub && peer.Node.PubKey == channel.Node2Pub) ||
			(nodePublicKey == channel.Node2Pub && peer.Node.PubKey == channel.Node1Pub) {
			count++
		}
	}

	return c.Together.Contains(count)
}

func (c *Channels) checkIncomingDisabled(peer *lnrpc.NodeInfo) bool {
	if c.IncomingDisabled == nil {
		return true
	}

	disabledChannels := make([]float64, len(peer.Channels))
	for i, channel := range peer.Channels {
		policy := getNodePolicy(peer.Node.PubKey, channel, false)

		if policy.Disabled {
			disabledChannels[i] = 1
		}
	}

	return c.IncomingDisabled.Contains(disabledChannels)
}

func (c *Channels) checkOutgoingDisabled(peer *lnrpc.NodeInfo) bool {
	if c.OutgoingDisabled == nil {
		return true
	}

	disabledChannels := make([]float64, len(peer.Channels))
	for i, channel := range peer.Channels {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)

		if policy.Disabled {
			disabledChannels[i] = 1
		}
	}

	return c.OutgoingDisabled.Contains(disabledChannels)
}

func getNodePolicy(peerPublicKey string, channel *lnrpc.ChannelEdge, outgoing bool) *lnrpc.RoutingPolicy {
	if outgoing {
		if peerPublicKey == channel.Node1Pub {
			return channel.Node1Policy
		}

		return channel.Node2Policy
	}

	if peerPublicKey == channel.Node2Pub {
		return channel.Node1Policy
	}

	return channel.Node2Policy
}

func capacityFunc(channel *lnrpc.ChannelEdge) int64 {
	return channel.Capacity
}

func blockHeightFunc(channel *lnrpc.ChannelEdge) uint32 {
	return uint32(channel.ChannelId >> 40)
}

func timeLockDeltaFunc(peer *lnrpc.NodeInfo) channelFunc[uint32] {
	return func(channel *lnrpc.ChannelEdge) uint32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.TimeLockDelta
	}
}

func minHTLCFunc(peer *lnrpc.NodeInfo) channelFunc[int64] {
	return func(channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.MinHtlc
	}
}

func maxHTLCFunc(peer *lnrpc.NodeInfo) channelFunc[uint64] {
	return func(channel *lnrpc.ChannelEdge) uint64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.MaxHtlcMsat / 1000
	}
}

func lastUpdateFunc(peer *lnrpc.NodeInfo, now int64) channelFunc[uint32] {
	return func(channel *lnrpc.ChannelEdge) uint32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return uint32(now) - policy.LastUpdate
	}
}

func feeRatesFunc(peer *lnrpc.NodeInfo, outgoing bool) channelFunc[int64] {
	return func(channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.FeeRateMilliMsat / 1000
	}
}

func baseFeesFunc(peer *lnrpc.NodeInfo, outgoing bool) channelFunc[int64] {
	return func(channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.FeeBaseMsat / 1000
	}
}

func inboundFeeRatesFunc(peer *lnrpc.NodeInfo, outgoing bool) channelFunc[int32] {
	return func(channel *lnrpc.ChannelEdge) int32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.InboundFeeRateMilliMsat / 1000
	}
}

func inboundBaseFeesFunc(peer *lnrpc.NodeInfo, outgoing bool) channelFunc[int32] {
	return func(channel *lnrpc.ChannelEdge) int32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.InboundFeeBaseMsat / 1000
	}
}
