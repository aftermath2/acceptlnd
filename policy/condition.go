package policy

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
)

// Conditions represents a set of requirements that must be met to apply a policy.
type Conditions struct {
	IsPrivate     *bool     `yaml:"is_private,omitempty"`
	WantsZeroConf *bool     `yaml:"wants_zero_conf,omitempty"`
	Whitelist     *[]string `yaml:"whitelist,omitempty"`
	Blacklist     *[]string `yaml:"blacklist,omitempty"`
	Request       *Request  `yaml:"request,omitempty"`
	Node          *Node     `yaml:"node,omitempty"`
}

// Match returns true if all the conditions Match.
func (c *Conditions) Match(
	req *lnrpc.ChannelAcceptRequest,
	nodePubKey string,
	peerNode *lnrpc.NodeInfo,
) bool {
	if c == nil {
		return true
	}

	if c.checkWhitelist(peerNode.Node.PubKey) {
		return true
	}

	if !c.checkBlacklist(peerNode.Node.PubKey) {
		return false
	}

	if !c.checkIsPrivate(req.ChannelFlags != uint32(lnwire.FFAnnounceChannel)) {
		return false
	}

	if !c.checkWantsZeroConf(req.WantsZeroConf) {
		return false
	}

	if err := c.Request.evaluate(req); err != nil {
		return false
	}

	if err := c.Node.evaluate(nodePubKey, peerNode); err != nil {
		return false
	}

	return true
}

func (c *Conditions) checkWhitelist(publicKey string) bool {
	if c.Whitelist == nil {
		return false
	}

	for _, pubKey := range *c.Whitelist {
		if publicKey == pubKey {
			return true
		}
	}
	return false
}

func (c *Conditions) checkBlacklist(publicKey string) bool {
	if c.Blacklist == nil {
		return true
	}

	for _, pubKey := range *c.Blacklist {
		if publicKey == pubKey {
			return false
		}
	}
	return true
}

func (c *Conditions) checkIsPrivate(private bool) bool {
	if c.IsPrivate == nil {
		return true
	}
	return private == *c.IsPrivate
}

func (c *Conditions) checkWantsZeroConf(wantsZeroConf bool) bool {
	if c.WantsZeroConf == nil {
		return true
	}
	return wantsZeroConf == *c.WantsZeroConf
}
