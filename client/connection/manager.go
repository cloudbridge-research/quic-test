package connection

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"quic-test/internal"
	"quic-test/internal/ca"
	"quic-test/internal/integration"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"go.uber.org/zap"
)

// Manager управляет QUIC соединениями
type Manager struct {
	config    internal.TestConfig
	transport *quic.Transport
	session   quic.Connection
	mu        sync.RWMutex
}

// NewManager создает новый менеджер соединений
func NewManager(config internal.TestConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// Connect устанавливает QUIC соединение
func (m *Manager) Connect(ctx context.Context, si *integration.SimpleIntegration) (quic.Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Создаем TLS конфигурацию
	tlsConf, err := m.createTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	// Создаем UDP соединение
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}

	// Создаем Transport
	m.transport = &quic.Transport{
		Conn: udpConn,
	}

	// Создаем QUIC конфигурацию с tracer для BBRv3
	var quicConfig *quic.Config
	if si != nil && m.config.CongestionControl == "bbrv3" {
		logger, _ := zap.NewDevelopment()
		
		quicConfig = &quic.Config{
			Tracer: func(ctx context.Context, perspective logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
				connectionIDStr := fmt.Sprintf("conn_%s", connID.String())
				return integration.NewConnectionTracerForConnection(logger, si, connectionIDStr)
			},
		}
	}

	// Устанавливаем соединение
	session, err := m.transport.Dial(ctx, parseAddr(m.config.Addr), tlsConf, quicConfig)
	if err != nil {
		m.transport.Close()
		return nil, fmt.Errorf("failed to dial QUIC: %w", err)
	}

	m.session = session

	// Сохраняем connection для использования в tracer (если используется BBRv3)
	if si != nil && m.config.CongestionControl == "bbrv3" && session != nil {
		connectionID := fmt.Sprintf("conn_%s", session.RemoteAddr().String())
		integration.StoreConnection(connectionID, session)
	}

	return session, nil
}

// Close закрывает соединение
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	if m.session != nil {
		if err := m.session.CloseWithError(0, "client done"); err != nil {
			errs = append(errs, fmt.Errorf("failed to close session: %w", err))
		}
		m.session = nil
	}

	if m.transport != nil {
		if err := m.transport.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close transport: %w", err))
		}
		m.transport = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// GetConnectionState возвращает состояние соединения
func (m *Manager) GetConnectionState() quic.ConnectionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.session != nil {
		return m.session.ConnectionState()
	}
	return quic.ConnectionState{}
}

// createTLSConfig создает TLS конфигурацию
func (m *Manager) createTLSConfig() (*tls.Config, error) {
	if m.config.CertPath != "" && m.config.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(m.config.CertPath, m.config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		return &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-test"},
		}, nil
	}

	// Если сертификаты не указаны, пытаемся использовать CA
	if !m.config.NoTLS {
		// Проверяем, есть ли сертификаты от CA
		serverCert, serverKey, err := ca.SetupDefaultCertificates()
		if err != nil {
			fmt.Printf("Warning: Failed to setup CA certificates: %v\n", err)
			// Fallback на стандартную генерацию
			return internal.GenerateTLSConfig(m.config.NoTLS), nil
		}

		// Загружаем сертификат от CA
		cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
		if err != nil {
			fmt.Printf("Warning: Failed to load CA certificate: %v\n", err)
			return internal.GenerateTLSConfig(m.config.NoTLS), nil
		}

		return &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true, // Для тестирования
			NextProtos:         []string{"quic-test"},
		}, nil
	}

	// Используем единую функцию для генерации TLS конфигурации
	return internal.GenerateTLSConfig(m.config.NoTLS), nil
}

// parseAddr парсит адрес в формате "host:port" и возвращает *net.UDPAddr
func parseAddr(addr string) *net.UDPAddr {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		// Fallback на простой парсинг
		host, port := "127.0.0.1", "9000"
		if len(addr) > 0 {
			parts := splitHostPort(addr)
			if len(parts) == 2 {
				host, port = parts[0], parts[1]
				// Если host пустой (например, ":9000"), используем localhost
				if host == "" {
					host = "127.0.0.1"
				}
			} else if len(parts) == 1 {
				// Только порт (например, ":9000" или "9000")
				if parts[0] != "" {
					port = parts[0]
				}
			}
		}
		udpAddr = &net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: parseInt(port),
		}
	} else {
		// Проверяем, что IP не пустой и не IPv6 :: (который может вызвать проблемы)
		if udpAddr.IP == nil || udpAddr.IP.IsUnspecified() {
			// Если IP пустой или неопределенный, используем 127.0.0.1
			udpAddr.IP = net.ParseIP("127.0.0.1")
		}
	}
	return udpAddr
}

// splitHostPort разделяет "host:port"
func splitHostPort(addr string) []string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return []string{addr[:i], addr[i+1:]}
		}
	}
	return []string{addr}
}

// parseInt парсит строку в int
func parseInt(s string) int {
	val := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			val = val*10 + int(s[i]-'0')
		}
	}
	return val
}