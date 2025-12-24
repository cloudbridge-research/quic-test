# quic-test

Professional QUIC protocol testing platform for network engineers, researchers, and educators.

[![CI](https://github.com/cloudbridge-research/quic-test/actions/workflows/pipeline.yml/badge.svg)](https://github.com/cloudbridge-research/quic-test/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudbridge-research/quic-test)](https://goreportcard.com/report/github.com/cloudbridge-research/quic-test)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](readme_en.md) | **Русский**

## Что это?

`quic-test` — это профессиональная платформа для тестирования и анализа производительности протокола QUIC. Разработана для образовательных и исследовательских целей, с акцентом на воспроизводимость результатов и детальную аналитику.

**Основные возможности:**
- Web GUI интерфейс для менее технических пользователей
- Измерение latency, jitter, throughput для QUIC и TCP
- Эмуляция различных сетевых условий (потери, задержки, bandwidth)
- Real-time TUI визуализация
- Экспорт метрик в Prometheus
- WebTransport и HTTP/3 load testing
- Forward Error Correction (FEC) с SIMD оптимизацией
- Post-Quantum Cryptography симуляция
- BBRv3 congestion control с dual-scale bandwidth estimation

## Quick Start

### GUI Interface (Рекомендуется для начинающих)

```bash
# Сборка GUI
make build

# Запуск GUI сервера
make gui
# или
./quic-gui --addr=:8080 --api-addr=:8081

# Открыть браузер
open http://localhost:8080
```

**GUI возможности:**
- Создание тестов через веб-форму
- Real-time мониторинг активных тестов
- История тестов с детальными метриками
- Остановка тестов одним кликом
- Готовые пресеты для различных сценариев

### Docker (рекомендуется для production)

```bash
# Запуск с GUI
docker run -p 8080:8080 -p 8081:8081 -p 9000:9000/udp mlanies/quic-test:latest gui

# Запуск клиентского теста
docker run mlanies/quic-test:latest --mode=client --server=demo.quic.tech:4433

# Запуск сервера
docker run -p 4433:4433/udp mlanies/quic-test:latest --mode=server
```

### Command Line Interface

```bash
# Сборка из исходников
git clone https://github.com/cloudbridge-research/quic-test
cd quic-test

# Сборка FEC библиотеки (опционально, для лучшей производительности)
cd internal/fec && make && cd ../..

# Сборка всех компонентов
make build

# Базовый тест
./quic-test --mode=client --server=demo.quic.tech:4433
```

## Основные режимы

```bash
# Простой тест latency/throughput
./quic-test --mode=client --server=localhost:4433 --duration=30s

# Сравнение QUIC vs TCP
./quic-test --mode=client --compare-tcp --duration=60s

# Эмуляция мобильной сети
./quic-test --profile=mobile --duration=30s

# TUI мониторинг
./cmd/tui/tui --server=localhost:4433

# WebTransport тестирование
make test-webtransport

# HTTP/3 load testing
make test-http3
```

## Архитектура

```
quic-test/
├── cmd/
│   ├── gui/                # Web GUI интерфейс
│   ├── tui/                # Terminal UI мониторинг
│   ├── experimental/       # Экспериментальные функции
│   ├── quic-client/        # QUIC клиент
│   ├── quic-server/        # QUIC сервер
│   ├── dashboard/          # Дашборд
│   ├── masque/             # MASQUE VPN тесты
│   ├── ice/                # ICE/STUN/TURN тесты
│   └── security-test/      # Тесты безопасности
├── client/                 # QUIC клиент (legacy)
├── server/                 # QUIC сервер (legacy)
├── internal/
│   ├── gui/                # GUI сервер и API
│   ├── quic/               # QUIC логика
│   ├── fec/                # Forward Error Correction (C++/AVX2)
│   ├── congestion/         # BBRv2/BBRv3 алгоритмы
│   ├── webtransport/       # WebTransport поддержка
│   ├── http3/              # HTTP/3 load testing
│   ├── pqc/                # Post-Quantum Crypto симуляция
│   ├── metrics/            # Prometheus метрики
│   └── ai/                 # AI интеграция
├── web/                    # Web GUI статические файлы
│   ├── static/css/         # CSS стили
│   └── static/js/          # JavaScript
└── docs/                   # Документация
```

**Подробнее:** [docs/architecture.md](docs/architecture.md)

## Возможности

### Стабильные функции

- **Web GUI интерфейс** — удобный веб-интерфейс для создания и мониторинга тестов
- **QUIC client/server** — на базе quic-go с расширениями
- **Измерение RTT, jitter, throughput** — детальная аналитика производительности
- **Эмуляция сетевых профилей** — mobile, satellite, fiber, WiFi
- **TUI визуализация** — real-time мониторинг в терминале
- **Prometheus экспорт** — интеграция с системами мониторинга
- **BBRv2 congestion control** — современный алгоритм управления перегрузкой

### Экспериментальные функции

- **BBRv3 congestion control** — с dual-scale bandwidth estimation и 2% loss threshold
- **Forward Error Correction (FEC)** — с AVX2/SIMD оптимизацией
- **WebTransport support** — тестирование WebTransport соединений
- **HTTP/3 load testing** — нагрузочное тестирование HTTP/3
- **Post-Quantum Cryptography** — симуляция PQC алгоритмов (ML-KEM, Dilithium)
- **MASQUE VPN тестирование** — тесты VPN через QUIC
- **ICE/STUN/TURN тесты** — тестирование NAT traversal

### В планах (Roadmap)

- Автоматическое обнаружение аномалий
- Multi-cloud deployment
- Расширенная AI интеграция
- Поддержка QUIC v2

**Полный roadmap:** [docs/roadmap.md](docs/roadmap.md)

## Документация

- **[Отчет о сотрудничестве с МЭИ](docs/MEI_COLLABORATION_REPORT.md)** — показатели проекта и программа стажировок
- **[Путеводитель для студентов](docs/STUDENT_GUIDE.md)** — терминология, TCP vs QUIC, RFC документы
- **[API Reference](docs/API_REFERENCE.md)** — полная справка по REST API
- **[CLI Reference](docs/cli.md)** — справка по командам
- **[Architecture](docs/architecture.md)** — детальная архитектура
- **[Education](docs/education.md)** — лабораторные работы для университетов
- **[AI Integration](docs/ai-routing-integration.md)** — интеграция с AI Routing Lab
- **[Case Studies](docs/case-studies.md)** — результаты тестов с методикой
- **[TUI User Guide](docs/TUI_USER_GUIDE.md)** — руководство по TUI интерфейсу

## GUI Интерфейс

Web GUI предоставляет удобный интерфейс для пользователей без глубоких технических знаний:

### Основные возможности GUI:
- **Dashboard** — обзор активных тестов и системного статуса
- **New Test** — создание тестов через веб-форму с валидацией
- **Test History** — просмотр всех выполненных тестов
- **Test Details** — детальный просмотр метрик и логов теста
- **Real-time Updates** — автоматическое обновление статуса тестов

### API Endpoints:
- `POST /api/tests` — создание нового теста
- `GET /api/tests` — получение списка тестов
- `GET /api/tests/{id}` — получение деталей теста
- `DELETE /api/tests/{id}` — остановка теста
- `GET /api/metrics/current` — текущие агрегированные метрики
- `GET /api/metrics/prometheus` — метрики в формате Prometheus

**Подробнее:** [docs/API_REFERENCE.md](docs/API_REFERENCE.md)

## Для университетов

Проект разработан с акцентом на образование и подготовку кадров. Включает готовые лабораторные работы, образовательные материалы и программу стажировок.

### Образовательные ресурсы:
- **[Путеводитель для студентов](docs/STUDENT_GUIDE.md)** — терминология, сравнение TCP vs QUIC, RFC документы
- **Практические лабораторные работы** с пошаговыми инструкциями
- **Готовые сценарии тестирования** для различных сетевых условий

### Лабораторные работы:
- **ЛР #1:** Основы QUIC — handshake, 0-RTT, миграция соединений
- **ЛР #2:** Congestion Control — сравнение BBRv2 vs BBRv3
- **ЛР #3:** Производительность — QUIC vs TCP в различных условиях
- **ЛР #4:** Forward Error Correction — влияние FEC на производительность
- **ЛР #5:** Post-Quantum Cryptography — тестирование PQC алгоритмов

### Программа стажировок CloudBridge Research

**Для студентов МЭИ доступны следующие возможности:**

**Карьерные траектории:**
- Junior Network Engineer (80,000 - 120,000 руб/мес)
- Protocol Research Developer (120,000 - 180,000 руб/мес)
- DevOps/Infrastructure Engineer (100,000 - 160,000 руб/мес)
- AI/ML Engineer (140,000 - 200,000 руб/мес)

**Условия стажировки:**
- Летняя стажировка: 40,000 руб/мес (3 месяца)
- Дипломная практика: 50,000 руб/мес (6 месяцев)
- Гибридный формат работы (офис + удаленка)
- Возможность трудоустройства после успешного завершения

**Подробнее:** [docs/education.md](docs/education.md) | [Отчет о сотрудничестве](docs/MEI_COLLABORATION_REPORT.md)

## Интеграция с AI Routing Lab

`quic-test` экспортирует метрики в Prometheus, которые используются в [AI Routing Lab](https://github.com/cloudbridge-research/ai-routing-lab) для обучения моделей предсказания оптимальных маршрутов.

**Пример:**
```bash
# Запуск с Prometheus экспортом
./quic-test --mode=server --prometheus-port=9090

# AI Routing Lab собирает метрики
curl http://localhost:9090/metrics
```

**Подробнее:** [docs/ai-routing-integration.md](docs/ai-routing-integration.md)

## Разработка

```bash
# Запуск тестов
make test

# Полный набор тестов
make all

# Smoke test
make smoke

# Сборка Docker образа
make docker-build

# Запуск в Docker
make docker-run

# Линтинг
golangci-lint run

# Статус сборки
make status
```

### Доступные Make команды:
- `make build` — сборка всех бинарных файлов
- `make gui` — запуск GUI сервера
- `make test` — базовые функциональные тесты
- `make bench-rtt` — бенчмарки RTT
- `make bench-loss` — бенчмарки потерь пакетов
- `make soak-2h` — 2-часовой стресс-тест
- `make regression` — полный набор регрессионных тестов
- `make performance` — тесты производительности

## Лицензия

MIT License. См. [LICENSE](LICENSE).

## Контакты

- **GitHub:** [cloudbridge-research/quic-test](https://github.com/cloudbridge-research/quic-test)
- **Блог:** [cloudbridge-research.ru](https://cloudbridge-research.ru)
- **Email:** info@cloudbridge-research.ru