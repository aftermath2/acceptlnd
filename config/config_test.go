package config

import (
	"testing"

	"github.com/aftermath2/acceptlnd/policy"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		desc string
		path string
		fail bool
	}{
		{
			desc: "Valid",
			path: "./testdata/config.yml",
		},
		{
			desc: "Invalid value type",
			path: "./testdata/invalid_config.yml",
			fail: true,
		},
		{
			desc: "Invalid value",
			path: "./testdata/invalid_config2.yml",
			fail: true,
		},
		{
			desc: "Non existent",
			path: "",
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := Load(tc.path)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tru := true

	cases := []struct {
		desc   string
		config Config
		fail   bool
	}{
		{
			desc: "Valid",
			config: Config{
				RPCAddress:      "127.0.0.1:10001",
				CertificatePath: "./testdata/tls.mock",
				MacaroonPath:    "./testdata/acceptlnd.mock",
				Policies: []*policy.Policy{
					{
						RejectPrivateChannels: &tru,
						Node: &policy.Node{
							Hybrid: &tru,
						},
						Request: &policy.Request{
							CommitmentTypes: &[]lnrpc.CommitmentType{
								lnrpc.CommitmentType_ANCHORS,
							},
						},
					},
				},
			},
			fail: false,
		},
		{
			desc: "Invalid RPC address",
			config: Config{
				RPCAddress: "localhost",
			},
			fail: true,
		},
		{
			desc: "Invalid certificate path",
			config: Config{
				RPCAddress:      "127.0.0.1:10001",
				CertificatePath: "tls.cert",
			},
			fail: true,
		},
		{
			desc: "Invalid macaroon path",
			config: Config{
				RPCAddress:      "127.0.0.1:10001",
				CertificatePath: "./testdata/tls.mock",
				MacaroonPath:    "admin.macaroon",
			},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validate(tc.config)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
