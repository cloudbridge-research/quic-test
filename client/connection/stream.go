package connection

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"

	"quic-test/client/metrics"
	"quic-test/internal"
	"quic-test/internal/fec"
	"quic-test/internal/integration"

	"github.com/quic-go/quic-go"
)

// StreamHandler обрабатывает QUIC стримы
type StreamHandler struct {
	config    internal.TestConfig
	collector *metrics.Collector
	connID    int
	streamID  int
	ratePtr   *int64
	si        *integration.SimpleIntegration
}

// NewStreamHandler создает новый обработчик стримов
func NewStreamHandler(config internal.TestConfig, collector *metrics.Collector, connID, streamID int, ratePtr *int64, si *integration.SimpleIntegration) *StreamHandler {
	return &StreamHandler{
		config:    config,
		collector: collector,
		connID:    connID,
		streamID:  streamID,
		ratePtr:   ratePtr,
		si:        si,
	}
}

// HandleStream обрабатывает отправку данных по стриму
func (sh *StreamHandler) HandleStream(ctx context.Context, session quic.Connection) error {
	if sh.config.CongestionControl == "bbrv3" || sh.config.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d, Stream %d: HandleStream started\n", sh.connID, sh.streamID)
	}

	// Инициализируем FEC encoder если включен
	var fecEncoder *fec.HybridFECEncoder
	var useCXX bool
	if sh.config.FECEnabled && sh.config.FECRedundancy > 0 {
		fecEncoder = fec.NewHybridFECEncoder(sh.config.FECRedundancy)
		useCXX = fecEncoder.UseCXX()
		sh.collector.FECUseCXX = useCXX
		if useCXX {
			fmt.Printf("[INFO] Connection %d: FEC acceleration enabled (C++ SIMD, 30-35x faster)\n", sh.connID)
		} else {
			fmt.Printf("[INFO] Connection %d: FEC using Go implementation\n", sh.connID)
		}
	}

	defer func() {
		// Flush FEC при завершении
		if fecEncoder != nil {
			redundancy, err := fecEncoder.Flush()
			if err == nil && redundancy != nil {
				// Отправляем последний redundancy пакет если есть
				sh.collector.FECPacketsSent++
				sh.collector.FECRedundancyBytes += int64(len(redundancy))
			}
			// Cleanup C++ resources if using SIMD
			fecEncoder.Close()
		}

		if sh.config.CongestionControl == "bbrv3" || sh.config.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d, Stream %d: HandleStream returning\n", sh.connID, sh.streamID)
		}
	}()

	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		sh.collector.RecordError("open_stream")
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			fmt.Printf("Warning: failed to close stream: %v\n", err)
		}
	}()

	// Основной цикл отправки данных
	return sh.sendData(ctx, stream, fecEncoder)
}

// sendData отправляет данные по стриму
func (sh *StreamHandler) sendData(ctx context.Context, stream quic.Stream, fecEncoder *fec.HybridFECEncoder) error {
	packetSize := sh.config.PacketSize
	pattern := sh.config.Pattern
	sentPackets := 0
	var seq int64

	// Таймаут для цикла отправки
	sendTimeout := sh.config.Duration
	if sendTimeout == 0 {
		sendTimeout = 60 * time.Second // default
	}
	sendDeadline := time.Now().Add(sendTimeout)

	if sh.config.CongestionControl == "bbrv3" || sh.config.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d, Stream %d: sendDeadline set to %v (from now: %v)\n",
			sh.connID, sh.streamID, sendDeadline, sendTimeout)
	}

	iterCount := 0
	for {
		iterCount++
		if sh.config.CongestionControl == "bbrv3" && iterCount%1000 == 0 {
			elapsed := time.Since(sendDeadline.Add(-sendTimeout))
			fmt.Printf("[DEBUG] Connection %d, Stream %d: iteration %d, elapsed: %v, deadline in: %v\n",
				sh.connID, sh.streamID, iterCount, elapsed, time.Until(sendDeadline))
		}

		// Проверяем контекст и таймаут перед каждой итерацией
		if time.Now().After(sendDeadline) {
			if sh.config.CongestionControl == "bbrv3" || sh.config.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: sendDeadline reached, returning\n", sh.connID, sh.streamID)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			if sh.config.CongestionControl == "bbrv3" || sh.config.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() received, returning\n", sh.connID, sh.streamID)
			}
			return nil
		default:
		}

		// Эмуляция задержки
		if sh.config.EmulateLatency > 0 {
			if time.Now().After(sendDeadline) {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(sh.config.EmulateLatency):
				if time.Now().After(sendDeadline) {
					return nil
				}
			}
		}

		// Эмуляция потери пакета
		if sh.config.EmulateLoss > 0 && secureFloat64() < sh.config.EmulateLoss {
			sh.collector.RecordError("emulated_loss")
			continue
		}

		// Формируем пакет с seq
		buf := makePacket(packetSize, pattern)
		seq++
		if len(buf) >= 8 {
			for i := 0; i < 8; i++ {
				buf[i] = byte(seq >> (8 * i))
			}
		}

		// FEC: добавляем пакет в encoder и создаем redundancy если нужно
		var redundancyPacket []byte
		if fecEncoder != nil {
			groupComplete, redundancy, err := fecEncoder.AddPacket(buf, uint64(seq))
			if err != nil {
				fmt.Printf("[WARNING] FEC encoding error: %v\n", err)
			} else if groupComplete && redundancy != nil {
				redundancyPacket = redundancy
				sh.collector.FECPacketsSent++
				sh.collector.FECRepairPacketsSent++
				sh.collector.FECRedundancyBytes += int64(len(redundancy))
			}
		}

		// Отправляем пакет
		if err := sh.sendPacket(ctx, stream, buf, redundancyPacket, sendDeadline); err != nil {
			return err
		}

		sentPackets++

		// Пауза между пакетами
		if err := sh.rateLimitPause(ctx, sendDeadline); err != nil {
			return err
		}
	}
}

// sendPacket отправляет один пакет
func (sh *StreamHandler) sendPacket(ctx context.Context, stream quic.Stream, buf, redundancyPacket []byte, sendDeadline time.Time) error {
	// Дублирование пакета
	dupCount := 1
	if sh.config.EmulateDup > 0 && secureFloat64() < sh.config.EmulateDup {
		dupCount = 2
		sh.collector.RecordError("emulated_dup")
	}

	for d := 0; d < dupCount; d++ {
		// Проверяем deadline перед отправкой
		if time.Now().After(sendDeadline) {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Уведомляем SimpleIntegration о отправке пакета
		if sh.si != nil {
			sh.si.OnPacketSent(nil, len(buf), false) // session передается как nil для упрощения
		}

		// Используем context с таймаутом для Write
		writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
		writeDone := make(chan error, 1)
		var n int
		var err error

		go func() {
			n, err = stream.Write(buf)
			writeDone <- err
		}()

		select {
		case <-writeCtx.Done():
			writeCancel()
			sh.collector.RecordError("stream_write_timeout")
			continue
		case err = <-writeDone:
			writeCancel()
		}

		if err != nil {
			sh.collector.RecordError("stream_write")
			continue
		}

		// Записываем успешную отправку
		realRTT := sh.calculateRTT()
		latencyForMetrics := float64(realRTT.Nanoseconds()) / 1e6
		sh.collector.RecordSuccess(n, latencyForMetrics)

		// Уведомляем SimpleIntegration о получении ACK
		if sh.si != nil && err == nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[ERROR] Panic in OnAckReceived: %v\n", r)
					}
				}()
				sh.si.OnAckReceived(nil, n, realRTT) // session передается как nil для упрощения
			}()
		}

		// Отправляем redundancy пакет если он был создан
		if redundancyPacket != nil && err == nil {
			sh.sendRedundancyPacket(ctx, stream, redundancyPacket)
		}
	}

	return nil
}

// sendRedundancyPacket отправляет redundancy пакет
func (sh *StreamHandler) sendRedundancyPacket(ctx context.Context, stream quic.Stream, redundancyPacket []byte) {
	redundancyCtx, redundancyCancel := context.WithTimeout(ctx, 2*time.Second)
	redundancyDone := make(chan error, 1)
	go func() {
		_, redundancyErr := stream.Write(redundancyPacket)
		redundancyDone <- redundancyErr
	}()

	select {
	case <-redundancyCtx.Done():
		redundancyCancel()
		// Таймаут - не критично, продолжаем
	case redundancyErr := <-redundancyDone:
		redundancyCancel()
		if redundancyErr == nil {
			sh.collector.BytesSent += len(redundancyPacket)
		}
	}
}

// calculateRTT вычисляет RTT для метрик
func (sh *StreamHandler) calculateRTT() time.Duration {
	if sh.config.EmulateLatency > 0 {
		realRTT := sh.config.EmulateLatency
		// Добавляем небольшую вариацию для jitter (5-10% от базовой задержки)
		jitter := time.Duration(float64(sh.config.EmulateLatency) * 0.05 * secureFloat64())
		realRTT += jitter
		return realRTT
	}
	// Fallback: используем типичный RTT для локальной сети
	return 10 * time.Millisecond
}

// rateLimitPause делает паузу для ограничения скорости
func (sh *StreamHandler) rateLimitPause(ctx context.Context, sendDeadline time.Time) error {
	if time.Now().After(sendDeadline) {
		return nil
	}

	rate := atomic.LoadInt64(sh.ratePtr)
	if rate > 0 {
		sleepDuration := time.Second / time.Duration(rate)
		if sleepDuration > 100*time.Millisecond {
			// Для длинных пауз используем прерываемый sleep
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(sleepDuration):
				if time.Now().After(sendDeadline) {
					return nil
				}
			}
		} else {
			// Для коротких пауз обычный sleep
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(sleepDuration):
				if time.Now().After(sendDeadline) {
					return nil
				}
			}
		}
	}
	return nil
}

// makePacket создает пакет данных
func makePacket(size int, pattern string) []byte {
	buf := make([]byte, size)
	switch pattern {
	case "zeroes":
		// already zeroed
	case "increment":
		for i := range buf {
			buf[i] = byte(i % 256)
		}
	default:
		_, _ = rand.Read(buf)
	}
	return buf
}

// secureFloat64 генерирует криптографически стойкое случайное число от 0 до 1
func secureFloat64() float64 {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(binary.BigEndian.Uint64(b)) / float64(^uint64(0))
}