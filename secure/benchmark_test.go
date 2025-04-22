package secure

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	"water_pipe/config"
)

func BenchmarkAESEncryption(b *testing.B) {
	keyDir, err := setupAESKey()
	if err != nil {
		b.Fatalf("Failed to setup AES key: %v", err)
	}
	defer os.RemoveAll(keyDir)

	cfg := config.SecureConfig{
		Method: "aes",
		AES: config.AESConfig{
			KeyFile: filepath.Join(keyDir, "aes.key"),
		},
	}

	benchmarkEncryption(b, cfg)
}

func BenchmarkNoEncryption(b *testing.B) {
	cfg := config.SecureConfig{
		Method: "none",
	}

	benchmarkEncryption(b, cfg)
}

func benchmarkEncryption(b *testing.B, cfg config.SecureConfig) {
	layer, err := NewLayer(cfg)
	if err != nil {
		b.Fatalf("Failed to create secure layer: %v", err)
	}

	listener, err := layer.Listen("127.0.0.1:0")
	if err != nil {
		b.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.Addr().String()
	errCh := make(chan error, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}
				errCh <- err
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				_, err := io.Copy(ioutil.Discard, c)
				if err != nil && err != io.EOF {
					errCh <- err
				}
			}(conn)
		}
	}()

	client, err := layer.NewClient("test-node", serverAddr, "")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	testData := make([]byte, 10*1024)
	_, err = rand.Read(testData)
	if err != nil {
		b.Fatalf("Failed to generate random data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := client.Connect("")
		if err != nil {
			b.Fatalf("Failed to connect to server: %v", err)
		}

		_, err = conn.Write(testData)
		if err != nil {
			conn.Close()
			b.Fatalf("Failed to send data: %v", err)
		}

		conn.Close()
	}

	b.StopTimer()

	select {
	case err := <-errCh:
		b.Fatalf("Error in server goroutine: %v", err)
	default:
	}
}

func BenchmarkAESEncryptionVarious(b *testing.B) {
	keyDir, err := setupAESKey()
	if err != nil {
		b.Fatalf("Failed to setup AES key: %v", err)
	}
	defer os.RemoveAll(keyDir)

	cfg := config.SecureConfig{
		Method: "aes",
		AES: config.AESConfig{
			KeyFile: filepath.Join(keyDir, "aes.key"),
		},
	}

	dataSizes := []int{1024, 10 * 1024, 100 * 1024} // 1KB, 10KB, 100KB

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("Size-%dKB", size/1024), func(b *testing.B) {
			benchmarkEncryptionSize(b, cfg, size)
		})
	}
}

func BenchmarkNoEncryptionVarious(b *testing.B) {
	cfg := config.SecureConfig{
		Method: "none",
	}

	dataSizes := []int{1024, 10 * 1024, 100 * 1024} // 1KB, 10KB, 100KB

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("Size-%dKB", size/1024), func(b *testing.B) {
			benchmarkEncryptionSize(b, cfg, size)
		})
	}
}

func benchmarkEncryptionSize(b *testing.B, cfg config.SecureConfig, dataSize int) {
	layer, err := NewLayer(cfg)
	if err != nil {
		b.Fatalf("Failed to create secure layer: %v", err)
	}

	listener, err := layer.Listen("127.0.0.1:0")
	if err != nil {
		b.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.Addr().String()
	errCh := make(chan error, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}
				errCh <- err
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				_, err := io.Copy(ioutil.Discard, c)
				if err != nil && err != io.EOF {
					errCh <- err
				}
			}(conn)
		}
	}()

	client, err := layer.NewClient("test-node", serverAddr, "")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	testData := make([]byte, dataSize)
	_, err = rand.Read(testData)
	if err != nil {
		b.Fatalf("Failed to generate random data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := client.Connect("")
		if err != nil {
			b.Fatalf("Failed to connect to server: %v", err)
		}

		_, err = conn.Write(testData)
		if err != nil {
			conn.Close()
			b.Fatalf("Failed to send data: %v", err)
		}

		conn.Close()
	}

	b.StopTimer()

	select {
	case err := <-errCh:
		b.Fatalf("Error in server goroutine: %v", err)
	default:
	}
}
