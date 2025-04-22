package socks5

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/armon/go-socks5"
	"water_pipe/config"
)

type Server struct {
	config     config.SOCKS5Config
	socks5     *socks5.Server
	handler    ConnectionHandler
}

type ConnectionHandler func(conn net.Conn, target string) error

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewServer(cfg config.SOCKS5Config) (*Server, error) {
	socks5Config := &socks5.Config{}

	if cfg.Auth.Method == "username_password" {
		creds, err := loadCredentials(cfg.Auth.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}

		credStore := socks5.StaticCredentials{}
		for _, cred := range creds {
			credStore[cred.Username] = cred.Password
		}

		auth := socks5.UserPassAuthenticator{
			Credentials: credStore,
		}
		socks5Config.AuthMethods = []socks5.Authenticator{auth}
	}

	server := &Server{
		config: cfg,
	}

	socks5Server, err := socks5.New(socks5Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 server: %w", err)
	}

	server.socks5 = socks5Server
	return server, nil
}

func (s *Server) Serve(ctx context.Context, listener net.Listener, handler ConnectionHandler) error {
	s.handler = handler

	dialer := &customDialer{
		handler: handler,
	}

	socks5Config := &socks5.Config{
		Dial: dialer.Dial,
	}

	if s.config.Auth.Method == "username_password" {
		creds, err := loadCredentials(s.config.Auth.CredentialsFile)
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		credStore := socks5.StaticCredentials{}
		for _, cred := range creds {
			credStore[cred.Username] = cred.Password
		}

		auth := socks5.UserPassAuthenticator{
			Credentials: credStore,
		}
		socks5Config.AuthMethods = []socks5.Authenticator{auth}
	}

	newServer, err := socks5.New(socks5Config)
	if err != nil {
		return fmt.Errorf("failed to create SOCKS5 server with custom dialer: %w", err)
	}

	s.socks5 = newServer

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return fmt.Errorf("failed to accept connection: %w", err)
			}
		}

		go func() {
			if err := s.socks5.ServeConn(conn); err != nil {
				fmt.Printf("Error serving SOCKS5 connection: %v\n", err)
			}
		}()
	}
}

func loadCredentials(path string) ([]Credentials, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds []Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return creds, nil
}

type customDialer struct {
	handler ConnectionHandler
}

type tcpAddrConn struct {
	net.Conn
	localAddr  *net.TCPAddr
	remoteAddr *net.TCPAddr
}

func (c *tcpAddrConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *tcpAddrConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (d *customDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	client, server := net.Pipe()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	portNum, err := net.LookupPort(network, port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	var ip net.IP
	if host == "" {
		ip = net.IPv4(127, 0, 0, 1)
	} else {
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			ip = net.IPv4(127, 0, 0, 1)
		} else {
			ip = ips[0]
		}
	}

	remoteAddr := &net.TCPAddr{
		IP:   ip,
		Port: portNum,
	}
	localAddr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0, // Ephemeral port
	}

	clientConn := &tcpAddrConn{
		Conn:       client,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	go func() {
		if err := d.handler(server, addr); err != nil {
			fmt.Printf("Error handling connection: %v\n", err)
			server.Close()
		}
	}()

	return clientConn, nil
}
