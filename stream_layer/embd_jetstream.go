package streamlayer

import (
	"net"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/config"
)

type internalJetstream struct {
	server *server.Server
	mu     *sync.Mutex
}

var embdJetstream = &internalJetstream{
	server: nil,
	mu:     &sync.Mutex{},
}

func parseHostAndPort(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func startStreamLayer(nodeName string) (*internalJetstream, error) {
	embdJetstream.mu.Lock()
	defer embdJetstream.mu.Unlock()

	if embdJetstream.server != nil {
		return embdJetstream, nil
	}

	host, port, err := parseHostAndPort(config.Config.JetStream.BindAddress)
	if err != nil {
		return nil, err
	}
	opts := &server.Options{
		ServerName:         nodeName,
		Host:               host,
		Port:               port,
		NoSigs:             true,
		JetStream:          true,
		JetStreamMaxMemory: -1,
		JetStreamMaxStore:  -1,
		Cluster: server.ClusterOpts{
			Name: config.ClusterGroup,
		},
		LeafNode: server.LeafNodeOpts{},
	}

	if *config.ClusterPeersFlag != "" {
		opts.Routes = server.RoutesFromStr(*config.ClusterPeersFlag)
	}

	if *config.ClusterAddrFlag != "" {
		host, port, err := parseHostAndPort(*config.ClusterAddrFlag)
		if err != nil {
			return nil, err
		}
		opts.Cluster.ListenStr = *config.ClusterAddrFlag
		opts.Cluster.Host = host
		opts.Cluster.Port = port
	}

	if *config.LeafServerFlag != "" {
		opts.LeafNode.Remotes = parseRemoteLeafOpts()
	}

	if config.Config.JetStream.ServerConfigFile != "" {
		err := opts.ProcessConfigFile(config.Config.JetStream.ServerConfigFile)
		if err != nil {
			return nil, err
		}
	}

	originalRoutes := opts.Routes
	if len(opts.Routes) != 0 {
		opts.Routes = flattenRoutes(originalRoutes, true)
	}

	if opts.StoreDir == "" {
		opts.StoreDir = path.Join(config.DataRootDir, "stream", nodeName)
	}

	s, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	s.SetLogger(
		&streamlayerLogger{log.With().Str("from", "streamLayer").Logger()},
		opts.Debug,
		opts.Trace,
	)
	s.Start()

	embdJetstream.server = s
	return embdJetstream, nil
}

func (xx *internalJetstream) prepareConnection(opts ...nats.Option) (*nats.Conn, error) {
	xx.mu.Lock()
	s := xx.server
	xx.mu.Unlock()

	for !s.ReadyForConnections(1 * time.Second) {
		continue
	}

	opts = append(opts, nats.InProcessServer(s))
	for {
		c, err := nats.Connect("", opts...)
		if err != nil {
			log.Warn().Err(err).Msg("NATS server not accepting connections...")
			continue
		}

		j, err := c.JetStream()
		if err != nil {
			return nil, err
		}

		st, err := j.StreamInfo("nnv-r", nats.MaxWait(1*time.Second))
		if err == nats.ErrStreamNotFound || st != nil {
			log.Info().Msg("Streaming ready...")
			return c, nil
		}

		c.Close()
		log.Debug().Err(err).Msg("Streams not ready, waiting for NATS streams to come up...")
		time.Sleep(1 * time.Second)
	}
}
