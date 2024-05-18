package policy

import (
	"errors"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
)

// Channels represents a set of requirements that the initiator's node channels must satisfy.
type Channels struct {
	Number          *Range[uint32]      `yaml:"number,omitempty"`
	Capacity        *StatRange[int64]   `yaml:"capacity,omitempty"`
	ZeroBaseFees    *bool               `yaml:"zero_base_fees,omitempty"`
	BlockHeight     *StatRange[uint32]  `yaml:"block_height,omitempty"`
	TimeLockDelta   *StatRange[uint32]  `yaml:"time_lock_delta,omitempty"`
	MinHTLC         *StatRange[int64]   `yaml:"min_htlc,omitempty"`
	MaxHTLC         *StatRange[uint64]  `yaml:"max_htlc,omitempty"`
	LastUpdateDiff  *StatRange[uint32]  `yaml:"last_update_diff,omitempty"`
	Together        *Range[int]         `yaml:"together,omitempty"`
	FeeRates        *StatRange[int64]   `yaml:"fee_rates,omitempty"`
	BaseFees        *StatRange[int64]   `yaml:"base_fees,omitempty"`
	Disabled        *StatRange[float64] `yaml:"disabled,omitempty"`
	InboundFeeRates *StatRange[int32]   `yaml:"inbound_fees_rates,omitempty"`
	InboundBaseFees *StatRange[int32]   `yaml:"inbound_base_fees,omitempty"`
	Peers           *Peers              `yaml:"peers,omitempty"`
}

// Peers contains information about the initiator node channels peers.
//
// Fields must be duplicated to follow the YAML structure desired.
type Peers struct {
	FeeRates        *StatRange[int64]   `yaml:"fee_rates,omitempty"`
	BaseFees        *StatRange[int64]   `yaml:"base_fees,omitempty"`
	Disabled        *StatRange[float64] `yaml:"disabled,omitempty"`
	InboundFeeRates *StatRange[int32]   `yaml:"inbound_fees_rates,omitempty"`
	InboundBaseFees *StatRange[int32]   `yaml:"inbound_base_fees,omitempty"`
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

	if !checkStat(c.TimeLockDelta, peer, timeLockDeltaFunc()) {
		return errors.New("Time lock delta " + c.TimeLockDelta.Reason())
	}

	if !checkStat(c.MinHTLC, peer, minHTLCFunc()) {
		return errors.New("Channels minimum HTLC " + c.MinHTLC.Reason())
	}

	if !checkStat(c.MaxHTLC, peer, maxHTLCFunc()) {
		return errors.New("Channels maximum HTLC " + c.MaxHTLC.Reason())
	}

	if !checkStat(c.LastUpdateDiff, peer, lastUpdateFunc(time.Now().Unix())) {
		return errors.New("Channels last update " + c.LastUpdateDiff.Reason())
	}

	if !c.checkTogether(nodePublicKey, peer) {
		return errors.New("Channels together " + c.Together.Reason())
	}

	if !checkStat(c.FeeRates, peer, feeRatesFunc(true)) {
		return errors.New("Channels fee rates " + c.FeeRates.Reason())
	}

	if !checkStat(c.BaseFees, peer, baseFeesFunc(true)) {
		return errors.New("Channels base fees " + c.BaseFees.Reason())
	}

	if !checkStat(c.InboundFeeRates, peer, inboundFeeRatesFunc(true)) {
		return errors.New("Channels inbound fee rates " + c.InboundFeeRates.Reason())
	}

	if !checkStat(c.InboundBaseFees, peer, inboundBaseFeesFunc(true)) {
		return errors.New("Channels inbound base fees " + c.InboundBaseFees.Reason())
	}

	if !c.checkDisabled(peer) {
		return errors.New("Disabled channels " + c.Disabled.Reason())
	}

	if c.Peers == nil {
		return nil
	}

	if !checkStat(c.Peers.FeeRates, peer, feeRatesFunc(false)) {
		return errors.New("Peers fee rates " + c.Peers.FeeRates.Reason())
	}

	if !checkStat(c.Peers.BaseFees, peer, baseFeesFunc(false)) {
		return errors.New("Peers base fees " + c.Peers.BaseFees.Reason())
	}

	if !checkStat(c.Peers.InboundFeeRates, peer, inboundFeeRatesFunc(false)) {
		return errors.New("Peers inbound fee rates " + c.Peers.InboundFeeRates.Reason())
	}

	if !checkStat(c.Peers.InboundBaseFees, peer, inboundBaseFeesFunc(false)) {
		return errors.New("Peers inbound base fees " + c.Peers.InboundBaseFees.Reason())
	}

	if !c.checkPeersDisabled(peer) {
		return errors.New("Peers disabled channels " + c.Peers.Disabled.Reason())
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

func (c *Channels) checkDisabled(peer *lnrpc.NodeInfo) bool {
	if c.Disabled == nil {
		return true
	}

	disabledChannels := make([]float64, len(peer.Channels))
	for i, channel := range peer.Channels {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)

		if policy.Disabled {
			disabledChannels[i] = 1
		}
	}

	return c.Disabled.Contains(disabledChannels)
}

func (c *Channels) checkPeersDisabled(peer *lnrpc.NodeInfo) bool {
	if c.Peers.Disabled == nil {
		return true
	}

	disabledChannels := make([]float64, len(peer.Channels))
	for i, channel := range peer.Channels {
		policy := getNodePolicy(peer.Node.PubKey, channel, false)

		if policy.Disabled {
			disabledChannels[i] = 1
		}
	}

	return c.Peers.Disabled.Contains(disabledChannels)
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

func capacityFunc(_ *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int64 {
	return channel.Capacity
}

func blockHeightFunc(_ *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) uint32 {
	return uint32(channel.ChannelId >> 40)
}

func timeLockDeltaFunc() channelFunc[uint32] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) uint32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.TimeLockDelta
	}
}

func minHTLCFunc() channelFunc[int64] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.MinHtlc
	}
}

func maxHTLCFunc() channelFunc[uint64] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) uint64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return policy.MaxHtlcMsat / 1000
	}
}

func lastUpdateFunc(now int64) channelFunc[uint32] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) uint32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, true)
		return uint32(now) - policy.LastUpdate
	}
}

func feeRatesFunc(outgoing bool) channelFunc[int64] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.FeeRateMilliMsat / 1000
	}
}

func baseFeesFunc(outgoing bool) channelFunc[int64] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int64 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.FeeBaseMsat / 1000
	}
}

func inboundFeeRatesFunc(outgoing bool) channelFunc[int32] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.InboundFeeRateMilliMsat / 1000
	}
}

func inboundBaseFeesFunc(outgoing bool) channelFunc[int32] {
	return func(peer *lnrpc.NodeInfo, channel *lnrpc.ChannelEdge) int32 {
		policy := getNodePolicy(peer.Node.PubKey, channel, outgoing)
		return policy.InboundFeeBaseMsat / 1000
	}
}
