package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CA представляет центр сертификации
type CA struct {
	CertPath string
	KeyPath  string
	cert     *x509.Certificate
	key      *rsa.PrivateKey
}

// NewCA создает новый центр сертификации
func NewCA(certPath, keyPath string) *CA {
	return &CA{
		CertPath: certPath,
		KeyPath:  keyPath,
	}
}

// Initialize инициализирует CA (создает или загружает сертификат и ключ)
func (ca *CA) Initialize() error {
	// Проверяем, существуют ли файлы CA
	if _, err := os.Stat(ca.CertPath); os.IsNotExist(err) {
		// Создаем новый CA
		return ca.createCA()
	}
	
	// Загружаем существующий CA
	return ca.loadCA()
}

// createCA создает новый центр сертификации
func (ca *CA) createCA() error {
	// Генерируем приватный ключ для CA
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}
	ca.key = key

	// Создаем шаблон сертификата для CA
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "QUIC Test CA",
			Organization: []string{"QUIC Test Suite"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Создаем сертификат CA (самоподписанный)
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Парсим созданный сертификат
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	ca.cert = cert

	// Сохраняем сертификат CA
	if err := ca.saveCertificate(ca.CertPath, certDER); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}

	// Сохраняем ключ CA
	if err := ca.savePrivateKey(ca.KeyPath, key); err != nil {
		return fmt.Errorf("failed to save CA key: %w", err)
	}

	fmt.Printf("Created new CA certificate: %s\n", ca.CertPath)
	return nil
}

// loadCA загружает существующий центр сертификации
func (ca *CA) loadCA() error {
	// Загружаем сертификат
	certPEM, err := os.ReadFile(ca.CertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	ca.cert = cert

	// Загружаем ключ
	keyPEM, err := os.ReadFile(ca.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode CA key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA key: %w", err)
	}
	ca.key = key

	fmt.Printf("Loaded existing CA certificate: %s\n", ca.CertPath)
	return nil
}

// GenerateServerCertificate генерирует сертификат сервера, подписанный CA
func (ca *CA) GenerateServerCertificate(hosts []string, certPath, keyPath string) error {
	// Генерируем ключ для сервера
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate server key: %w", err)
	}

	// Создаем шаблон сертификата сервера
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   hosts[0], // Первый хост как CN
			Organization: []string{"QUIC Test Suite"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Добавляем все хосты в SAN (Subject Alternative Names)
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	// Создаем сертификат, подписанный CA
	certDER, err := x509.CreateCertificate(rand.Reader, &template, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Сохраняем сертификат сервера
	if err := ca.saveCertificate(certPath, certDER); err != nil {
		return fmt.Errorf("failed to save server certificate: %w", err)
	}

	// Сохраняем ключ сервера
	if err := ca.savePrivateKey(keyPath, key); err != nil {
		return fmt.Errorf("failed to save server key: %w", err)
	}

	fmt.Printf("Generated server certificate for hosts %v: %s\n", hosts, certPath)
	return nil
}

// saveCertificate сохраняет сертификат в PEM формате
func (ca *CA) saveCertificate(path string, certDER []byte) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	certOut, err := os.Create(path)
	if err != nil {
		return err
	}
	defer certOut.Close()

	return pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

// savePrivateKey сохраняет приватный ключ в PEM формате
func (ca *CA) savePrivateKey(path string, key *rsa.PrivateKey) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	keyOut, err := os.Create(path)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	return pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// GetCACertificate возвращает сертификат CA для добавления в доверенные
func (ca *CA) GetCACertificate() *x509.Certificate {
	return ca.cert
}

// SetupDefaultCertificates создает CA и генерирует сертификаты для localhost
func SetupDefaultCertificates() (string, string, error) {
	// Создаем директорию для сертификатов
	certsDir := "certs"
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Инициализируем CA
	ca := NewCA(
		filepath.Join(certsDir, "ca.crt"),
		filepath.Join(certsDir, "ca.key"),
	)

	if err := ca.Initialize(); err != nil {
		return "", "", fmt.Errorf("failed to initialize CA: %w", err)
	}

	// Генерируем сертификат для localhost
	serverCert := filepath.Join(certsDir, "server.crt")
	serverKey := filepath.Join(certsDir, "server.key")

	// Проверяем, существует ли уже сертификат сервера
	if _, err := os.Stat(serverCert); os.IsNotExist(err) {
		hosts := []string{"localhost", "127.0.0.1", "::1"}
		if err := ca.GenerateServerCertificate(hosts, serverCert, serverKey); err != nil {
			return "", "", fmt.Errorf("failed to generate server certificate: %w", err)
		}
	} else {
		fmt.Printf("Using existing server certificate: %s\n", serverCert)
	}

	return serverCert, serverKey, nil
}