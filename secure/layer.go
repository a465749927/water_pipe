package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"

	"water_pipe/config"
)

type ConnectionWrapper interface {
	Wrap(conn net.Conn) (net.Conn, error)
}

type ListenerWrapper interface {
	WrapListener(listener net.Listener) (net.Listener, error)
}

type Layer struct {
	config         config.SecureConfig
	connectionWrapper ConnectionWrapper
	listenerWrapper   ListenerWrapper
}

type Client struct {
	nodeID            string
	address           string
	connectionWrapper ConnectionWrapper
}

func NewLayer(cfg config.SecureConfig) (*Layer, error) {
	var connectionWrapper ConnectionWrapper
	var listenerWrapper ListenerWrapper
	var err error

	switch cfg.Method {
	case "tls":
		connectionWrapper, listenerWrapper, err = newTLSWrapper(cfg.TLS)
	case "aes":
		connectionWrapper, listenerWrapper, err = newAESWrapper(cfg.AES)
	case "none":
		connectionWrapper, listenerWrapper = newNoEncryptionWrapper(), newNoEncryptionListenerWrapper()
	default:
		return nil, fmt.Errorf("unsupported encryption method: %s", cfg.Method)
	}

	if err != nil {
		return nil, err
	}

	return &Layer{
		config:         cfg,
		connectionWrapper: connectionWrapper,
		listenerWrapper:   listenerWrapper,
	}, nil
}

func (l *Layer) NewClient(nodeID, address, publicKeyPath string) (*Client, error) {
	return &Client{
		nodeID:            nodeID,
		address:           address,
		connectionWrapper: l.connectionWrapper,
	}, nil
}

func (l *Layer) Listen(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	secureListener, err := l.listenerWrapper.WrapListener(listener)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to create secure listener: %w", err)
	}

	return secureListener, nil
}

func (c *Client) Connect(target string) (net.Conn, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node %s: %w", c.nodeID, err)
	}

	secureConn, err := c.connectionWrapper.Wrap(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to secure connection to node %s: %w", c.nodeID, err)
	}


	return secureConn, nil
}

type tlsWrapper struct {
	tlsConfig *tls.Config
}

func newTLSWrapper(cfg config.TLSConfig) (ConnectionWrapper, ListenerWrapper, error) {
	// Load certificate
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Load CA certificate
	caData, err := ioutil.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caData) {
		return nil, nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}

	wrapper := &tlsWrapper{
		tlsConfig: tlsConfig,
	}

	return wrapper, wrapper, nil
}

func (w *tlsWrapper) Wrap(conn net.Conn) (net.Conn, error) {
	return tls.Client(conn, w.tlsConfig), nil
}

func (w *tlsWrapper) WrapListener(listener net.Listener) (net.Listener, error) {
	return tls.NewListener(listener, w.tlsConfig), nil
}

type aesWrapper struct {
	key []byte
}

func newAESWrapper(cfg config.AESConfig) (ConnectionWrapper, ListenerWrapper, error) {
	key, err := ioutil.ReadFile(cfg.KeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load AES key: %w", err)
	}

	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, nil, fmt.Errorf("invalid AES key length: %d bytes (must be 16, 24, or 32)", keyLen)
	}

	wrapper := &aesWrapper{
		key: key,
	}

	return wrapper, wrapper, nil
}

func (w *aesWrapper) Wrap(conn net.Conn) (net.Conn, error) {
	return &aesConn{
		Conn: conn,
		key:  w.key,
	}, nil
}

func (w *aesWrapper) WrapListener(listener net.Listener) (net.Listener, error) {
	return &aesListener{
		Listener: listener,
		key:      w.key,
	}, nil
}

type aesConn struct {
	net.Conn
	key       []byte
	readBuf   []byte
	writeBuf  []byte
	readMu    sync.Mutex
	writeMu   sync.Mutex
	blockSize int
}

func (c *aesConn) Read(b []byte) (n int, err error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.Conn, lenBuf); err != nil {
		return 0, err
	}
	encryptedLen := int(binary.BigEndian.Uint32(lenBuf))

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(c.Conn, nonce); err != nil {
		return 0, err
	}

	encrypted := make([]byte, encryptedLen)
	if _, err := io.ReadFull(c.Conn, encrypted); err != nil {
		return 0, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return 0, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	decrypted, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return 0, err
	}

	n = copy(b, decrypted)
	return n, nil
}

func (c *aesConn) Write(b []byte) (n int, err error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return 0, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return 0, err
	}

	encrypted := aead.Seal(nil, nonce, b, nil)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)))
	if _, err := c.Conn.Write(lenBuf); err != nil {
		return 0, err
	}

	if _, err := c.Conn.Write(nonce); err != nil {
		return 0, err
	}

	if _, err := c.Conn.Write(encrypted); err != nil {
		return 0, err
	}

	return len(b), nil
}

type aesListener struct {
	net.Listener
	key []byte
}

func (l *aesListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &aesConn{
		Conn: conn,
		key:  l.key,
	}, nil
}

type noEncryptionWrapper struct{}

func newNoEncryptionWrapper() ConnectionWrapper {
	return &noEncryptionWrapper{}
}

func (w *noEncryptionWrapper) Wrap(conn net.Conn) (net.Conn, error) {
	return conn, nil
}

type noEncryptionListenerWrapper struct{}

func newNoEncryptionListenerWrapper() ListenerWrapper {
	return &noEncryptionListenerWrapper{}
}

func (w *noEncryptionListenerWrapper) WrapListener(listener net.Listener) (net.Listener, error) {
	return listener, nil
}
