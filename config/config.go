package config

import (
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/aftermath2/acceptlnd/policy"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config is acceptLND's configuration schema.
type Config struct {
	RPCAddress      string           `yaml:"rpc_address,omitempty"`
	RPCTimeout      *time.Duration   `yaml:"rpc_timeout,omitempty"`
	CertificatePath string           `yaml:"certificate_path,omitempty"`
	MacaroonPath    string           `yaml:"macaroon_path,omitempty"`
	Policies        []*policy.Policy `yaml:"policies,omitempty"`
}

// Load reads the configuration file and returns a new object.
func Load(path string) (Config, error) {
	if path == "" {
		path = "acceptlnd.yml"
	}

	slog.Info("Configuration file: " + path)

	f, err := os.OpenFile(path, os.O_RDONLY, 0o600)
	if err != nil {
		return Config{}, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	var config Config
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return Config{}, errors.Wrap(err, "decoding configuration")
	}

	if err := validate(config); err != nil {
		return Config{}, errors.Wrap(err, "validating configuration")
	}

	return config, nil
}

func validate(config Config) error {
	_, _, err := net.SplitHostPort(config.RPCAddress)
	if err != nil {
		return errors.Wrap(err, "invalid RPC address")
	}

	if _, err := os.Stat(config.CertificatePath); os.IsNotExist(err) {
		return errors.New("the certificate file specified does not exist")
	}

	if _, err := os.Stat(config.MacaroonPath); os.IsNotExist(err) {
		return errors.New("the macaroon file specified does not exist")
	}

	return nil
}
