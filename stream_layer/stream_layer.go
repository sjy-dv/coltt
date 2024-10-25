package streamlayer

import (
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/config"
)

func Connect() (*nats.Conn, error) {
	opts := setupConnOptions()

	creds, err := getNatsAuthFromConfig()
	if err != nil {
		return nil, err
	}

	tls, err := getNatsTLSFromConfig()
	if err != nil {
		return nil, err
	}

	opts = append(opts, creds...)
	opts = append(opts, tls...)
	if len(config.Config.JetStream.URLs) == 0 {
		embedded, err := startStreamLayer(config.Config.NodeName())
		if err != nil {
			return nil, err
		}

		return embedded.prepareConnection(opts...)
	}

	url := strings.Join(config.Config.JetStream.URLs, ", ")

	var conn *nats.Conn
	for i := 0; i < config.Config.JetStream.ConnectRetries; i++ {
		conn, err = nats.Connect(url, opts...)
		if err == nil && conn.Status() == nats.CONNECTED {
			break
		}

		log.Warn().
			Err(err).
			Int("attempt", i+1).
			Int("attempt_limit", config.Config.JetStream.ConnectRetries).
			Str("status", conn.Status().String()).
			Msg("NATS connection failed")
	}

	return conn, err
}

func getNatsAuthFromConfig() ([]nats.Option, error) {
	opts := make([]nats.Option, 0)

	if config.Config.JetStream.CredsUser != "" {
		opt := nats.UserInfo(config.Config.JetStream.CredsUser, config.Config.JetStream.CredsPassword)
		opts = append(opts, opt)
	}

	if config.Config.JetStream.SeedFile != "" {
		opt, err := nats.NkeyOptionFromSeed(config.Config.JetStream.SeedFile)
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	return opts, nil
}

func getNatsTLSFromConfig() ([]nats.Option, error) {
	opts := make([]nats.Option, 0)

	if config.Config.JetStream.CAFile != "" {
		opt := nats.RootCAs(config.Config.JetStream.CAFile)
		opts = append(opts, opt)
	}

	if config.Config.JetStream.CertFile != "" && config.Config.JetStream.KeyFile != "" {
		opt := nats.ClientCert(config.Config.JetStream.CertFile, config.Config.JetStream.KeyFile)
		opts = append(opts, opt)
	}

	return opts, nil
}

func setupConnOptions() []nats.Option {
	return []nats.Option{
		nats.Name(config.Config.NodeName()),
		nats.RetryOnFailedConnect(true),
		nats.ReconnectWait(time.Duration(config.Config.JetStream.ReconnectWaitSeconds) * time.Second),
		nats.MaxReconnects(config.Config.JetStream.ConnectRetries),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Error().
				Err(nc.LastError()).
				Msg("NATS client exiting")
		}),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Error().
				Err(err).
				Msg("NATS client disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info().
				Str("url", nc.ConnectedUrl()).
				Msg("NATS client reconnected")
		}),
	}
}
