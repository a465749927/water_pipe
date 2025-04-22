package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"

	"water_pipe/config"
	"water_pipe/health"
	"water_pipe/secure"
	"water_pipe/socks5"
)

type Server struct {
	config      *config.Config
	socks5Server *socks5.Server
	healthChecker *health.Checker
	secureLayer *secure.Layer
	nodes       map[string]*RemoteNode
	listener    net.Listener
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

type RemoteNode struct {
	ID      string
	Address string
	Status  health.Status
	Client  *secure.Client
}

func NewServer(cfg *config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	secureLayer, err := secure.NewLayer(cfg.Secure)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize secure layer: %w", err)
	}

	healthChecker, err := health.NewChecker(cfg.HealthCheck)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize health checker: %w", err)
	}

	var socks5Server *socks5.Server
	if cfg.SOCKS5.Enabled {
		socks5Server, err = socks5.NewServer(cfg.SOCKS5)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize SOCKS5 server: %w", err)
		}
	}

	server := &Server{
		config:        cfg,
		socks5Server:  socks5Server,
		healthChecker: healthChecker,
		secureLayer:   secureLayer,
		nodes:         make(map[string]*RemoteNode),
		ctx:           ctx,
		cancel:        cancel,
	}

	for _, nodeCfg := range cfg.Nodes {
		client, err := secureLayer.NewClient(nodeCfg.ID, nodeCfg.Address, nodeCfg.PublicKey)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize client for node %s: %w", nodeCfg.ID, err)
		}

		server.nodes[nodeCfg.ID] = &RemoteNode{
			ID:      nodeCfg.ID,
			Address: nodeCfg.Address,
			Status:  health.StatusUnknown,
			Client:  client,
		}

		healthChecker.RegisterNode(nodeCfg.ID, nodeCfg.Address)
	}

	return server, nil
}

func (s *Server) Start() error {
	s.healthChecker.Start(s.ctx)

	statusCh := s.healthChecker.Subscribe()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleHealthUpdates(statusCh)
	}()

	if s.socks5Server != nil {
		listener, err := net.Listen("tcp", s.config.Node.ListenAddress)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", s.config.Node.ListenAddress, err)
		}
		s.listener = listener

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.socks5Server.Serve(s.ctx, listener, s.handleConnection)
		}()
	}

	return nil
}

func (s *Server) Stop() error {
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()

	return nil
}

func (s *Server) handleHealthUpdates(statusCh <-chan health.Update) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case update := <-statusCh:
			node, exists := s.nodes[update.NodeID]
			if exists {
				node.Status = update.Status
			}
		}
	}
}

func (s *Server) handleConnection(conn net.Conn, target string) error {
	policy, err := s.findForwardingPolicy(conn.RemoteAddr(), target)
	if err != nil {
		return err
	}

	switch policy.Action.Type {
	case "direct":
		return s.handleDirectForwarding(conn, target)
	case "forward":
		return s.handleNodeForwarding(conn, target, policy.Action.NextHop)
	default:
		return fmt.Errorf("unknown forwarding action type: %s", policy.Action.Type)
	}
}

func (s *Server) findForwardingPolicy(sourceAddr net.Addr, targetAddr string) (*config.ForwardingPolicyConfig, error) {
	if len(s.config.ForwardingPolicies) > 0 {
		return &s.config.ForwardingPolicies[0], nil
	}

	return &config.ForwardingPolicyConfig{
		Name: "default",
		Action: config.ForwardingActionConfig{
			Type: "direct",
		},
	}, nil
}

func (s *Server) handleDirectForwarding(clientConn net.Conn, target string) error {
	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		return fmt.Errorf("failed to connect to target %s: %w", target, err)
	}
	defer targetConn.Close()

	errCh := make(chan error, 2)
	go func() {
		_, err := copyData(targetConn, clientConn)
		errCh <- err
	}()
	go func() {
		_, err := copyData(clientConn, targetConn)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *Server) handleNodeForwarding(clientConn net.Conn, target string, nodeID string) error {
	node, exists := s.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if node.Status != health.StatusHealthy {
		return fmt.Errorf("node %s is not healthy", nodeID)
	}

	nodeConn, err := node.Client.Connect(target)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %w", nodeID, err)
	}
	defer nodeConn.Close()

	errCh := make(chan error, 2)
	go func() {
		_, err := copyData(nodeConn, clientConn)
		errCh <- err
	}()
	go func() {
		_, err := copyData(clientConn, nodeConn)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func copyData(dst, src net.Conn) (int64, error) {
	buffer := make([]byte, 32*1024)
	var total int64

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			written, err := dst.Write(buffer[:n])
			if err != nil {
				return total, err
			}
			total += int64(written)
		}
		if err != nil {
			return total, err
		}
	}
}
