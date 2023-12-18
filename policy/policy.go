// Package policy evaluates the set of conditions and requirements set by the node operator that a
// channel opening request must satisfy.
package policy

import (
	"errors"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
)

// Policy represents a set of requirements that a channel opening request must satisfy. They are
// enforced only if the conditions are met or do not exist.
type Policy struct {
	Conditions             *Conditions `yaml:"conditions,omitempty"`
	Request                *Request    `yaml:"request,omitempty"`
	Node                   *Node       `yaml:"node,omitempty"`
	Whitelist              *[]string   `yaml:"whitelist,omitempty"`
	Blacklist              *[]string   `yaml:"blacklist,omitempty"`
	ZeroConfList           *[]string   `yaml:"zero_conf_list,omitempty"`
	RejectAll              *bool       `yaml:"reject_all,omitempty"`
	RejectPrivateChannels  *bool       `yaml:"reject_private_channels,omitempty"`
	AcceptZeroConfChannels *bool       `yaml:"accept_zero_conf_channels,omitempty"`
}

// Evaluate set of policies.
func (p *Policy) Evaluate(
	req *lnrpc.ChannelAcceptRequest,
	node *lnrpc.GetInfoResponse,
	peer *lnrpc.NodeInfo,
) error {
	if p.Conditions != nil && !p.Conditions.Match(req, node, peer) {
		return nil
	}

	if !p.checkRejectAll() {
		return errors.New("No new channels are accepted")
	}

	if !p.checkWhitelist(peer.Node.PubKey) {
		return errors.New("Node is not whitelisted")
	}

	if !p.checkBlacklist(peer.Node.PubKey) {
		return errors.New("Node is blacklisted")
	}

	if !p.checkPrivate(req.ChannelFlags != uint32(lnwire.FFAnnounceChannel)) {
		return errors.New("Private channels are not accepted")
	}

	if !p.checkZeroConf(peer.Node.PubKey, req.WantsZeroConf) {
		return errors.New("Zero conf channels are not accepted")
	}

	if err := p.Request.evaluate(req); err != nil {
		return err
	}

	return p.Node.evaluate(node, peer)
}

func (p *Policy) checkRejectAll() bool {
	if p.RejectAll == nil {
		return true
	}
	return !*p.RejectAll
}

func (p *Policy) checkWhitelist(publicKey string) bool {
	if p.Whitelist == nil {
		return true
	}

	for _, pubKey := range *p.Whitelist {
		if publicKey == pubKey {
			return true
		}
	}
	return false
}

func (p *Policy) checkBlacklist(publicKey string) bool {
	if p.Blacklist == nil {
		return true
	}

	for _, pubKey := range *p.Blacklist {
		if publicKey == pubKey {
			return false
		}
	}
	return true
}

func (p *Policy) checkPrivate(private bool) bool {
	if p.RejectPrivateChannels == nil || !private {
		return true
	}
	return private && !*p.RejectPrivateChannels
}

func (p *Policy) checkZeroConf(publicKey string, wantsZeroConf bool) bool {
	if !wantsZeroConf {
		return true
	}

	if p.AcceptZeroConfChannels == nil || !*p.AcceptZeroConfChannels {
		return false
	}

	if p.ZeroConfList != nil {
		for _, pubKey := range *p.ZeroConfList {
			if publicKey == pubKey {
				return true
			}
		}

		return false
	}

	return true
}
