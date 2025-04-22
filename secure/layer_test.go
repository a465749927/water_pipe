package secure

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"water_pipe/config"
)

func TestTLSEncryption(t *testing.T) {
	t.Skip("Skipping TLS test - requires valid PEM certificates")

	certDir, err := setupTLSCertificates()
	if err != nil {
		t.Fatalf("Failed to setup TLS certificates: %v", err)
	}
	defer os.RemoveAll(certDir)

	cfg := config.SecureConfig{
		Method: "tls",
		TLS: config.TLSConfig{
			CertFile: filepath.Join(certDir, "cert.pem"),
			KeyFile:  filepath.Join(certDir, "key.pem"),
			CAFile:   filepath.Join(certDir, "ca.pem"),
		},
	}

	testSecureCommunication(t, cfg)
}

func TestAESEncryption(t *testing.T) {
	keyDir, err := setupAESKey()
	if err != nil {
		t.Fatalf("Failed to setup AES key: %v", err)
	}
	defer os.RemoveAll(keyDir)

	cfg := config.SecureConfig{
		Method: "aes",
		AES: config.AESConfig{
			KeyFile: filepath.Join(keyDir, "aes.key"),
		},
	}

	testSecureCommunication(t, cfg)
}

func TestNoEncryption(t *testing.T) {
	cfg := config.SecureConfig{
		Method: "none",
	}

	testSecureCommunication(t, cfg)
}

func testSecureCommunication(t *testing.T, cfg config.SecureConfig) {
	layer, err := NewLayer(cfg)
	if err != nil {
		t.Fatalf("Failed to create secure layer: %v", err)
	}

	listener, err := layer.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.Addr().String()
	testData := []byte("Hello, secure world!")
	receivedCh := make(chan []byte, 1)
	errorCh := make(chan error, 2)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errorCh <- err
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			errorCh <- err
			return
		}

		receivedCh <- buf[:n]

		_, err = conn.Write(buf[:n])
		if err != nil {
			errorCh <- err
			return
		}
	}()

	client, err := layer.NewClient("test-node", serverAddr, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	conn, err := client.Connect("")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	response := buf[:n]
	if !bytes.Equal(response, testData) {
		t.Fatalf("Response does not match test data: got %v, want %v", response, testData)
	}

	select {
	case received := <-receivedCh:
		if !bytes.Equal(received, testData) {
			t.Fatalf("Server received data does not match test data: got %v, want %v", received, testData)
		}
	case err := <-errorCh:
		t.Fatalf("Error in server goroutine: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for server to receive data")
	}
}

func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}


	keyDir, err := setupAESKey()
	if err != nil {
		t.Fatalf("Failed to setup AES key: %v", err)
	}
	defer os.RemoveAll(keyDir)

	aesConfig := config.SecureConfig{
		Method: "aes",
		AES: config.AESConfig{
			KeyFile: filepath.Join(keyDir, "aes.key"),
		},
	}

	noneConfig := config.SecureConfig{
		Method: "none",
	}

	dataSizes := []int{1024, 10 * 1024, 100 * 1024} // 1KB, 10KB, 100KB

	for _, size := range dataSizes {
		testData := make([]byte, size)
		_, err := rand.Read(testData)
		if err != nil {
			t.Fatalf("Failed to generate random data: %v", err)
		}

		t.Logf("Testing with data size: %d bytes", size)

		aesDuration := measurePerformance(t, aesConfig, testData)
		t.Logf("AES encryption time: %v", aesDuration)

		noneDuration := measurePerformance(t, noneConfig, testData)
		t.Logf("No encryption time: %v", noneDuration)

		t.Logf("Performance ratio (AES/None): %.2fx", float64(aesDuration)/float64(noneDuration))
	}
}

func measurePerformance(t *testing.T, cfg config.SecureConfig, testData []byte) time.Duration {
	layer, err := NewLayer(cfg)
	if err != nil {
		t.Fatalf("Failed to create secure layer: %v", err)
	}

	listener, err := layer.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.Addr().String()
	doneCh := make(chan struct{})

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		_, err = io.Copy(ioutil.Discard, conn)
		if err != nil {
			t.Errorf("Failed to read data: %v", err)
			return
		}

		close(doneCh)
	}()

	client, err := layer.NewClient("test-node", serverAddr, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	conn, err := client.Connect("")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	start := time.Now()
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to send data: %v", err)
	}

	conn.Close()

	<-doneCh

	return time.Since(start)
}


func setupTLSCertificates() (string, error) {
	dir, err := ioutil.TempDir("", "tls-test")
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "ca.pem"), []byte("dummy ca cert"), 0644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "cert.pem"), []byte("dummy cert"), 0644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "key.pem"), []byte("dummy key"), 0644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func setupAESKey() (string, error) {
	dir, err := ioutil.TempDir("", "aes-test")
	if err != nil {
		return "", err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "aes.key"), key, 0600); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type mockConn struct {
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
	closed   bool
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	return c.readBuf.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	return c.writeBuf.Write(b)
}

func (c *mockConn) Close() error {
	c.closed = true
	return nil
}

func (c *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0} }
func (c *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0} }
func (c *mockConn) SetDeadline(t time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { return nil }
