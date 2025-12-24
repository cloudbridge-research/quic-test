# Student Guide to QUIC

This document helps students understand QUIC protocol basics, its differences from TCP, and provides links to important RFC documents.

## Table of Contents

1. [Key Terms](#key-terms)
2. [TCP vs QUIC: Key Differences](#tcp-vs-quic-key-differences)
3. [QUIC Architecture](#quic-architecture)
4. [Congestion Control](#congestion-control)
5. [RFC Documents](#rfc-documents)
6. [Practical Assignments](#practical-assignments)
7. [Additional Resources](#additional-resources)
8. [Career Opportunities at CloudBridge Research](#career-opportunities-at-cloudbridge-research)

## Key Terms

### Basic Concepts

**QUIC (Quick UDP Internet Connections)**
- Transport protocol developed by Google, standardized by IETF
- Runs over UDP, provides reliability and security
- Foundation for HTTP/3

**Stream**
- Logical data channel within a QUIC connection
- Multiple streams can exist in parallel
- Each stream has unique ID and can be unidirectional or bidirectional

**Connection**
- Logical link between client and server
- Can contain multiple streams
- Supports migration between network paths

**Handshake**
- Connection establishment process
- Integrated with TLS 1.3 in QUIC
- Supports 0-RTT for repeat connections

### Performance Metrics

**RTT (Round-Trip Time)**
- Time for packet to travel from sender to receiver and back
- Key metric for network latency assessment
- Measured in milliseconds (ms)

**Jitter**
- Variation in packet delay
- Shows network connection stability
- Important for real-time applications

**Throughput**
- Amount of data transmitted per time unit
- Measured in bits/bytes per second (bps, Mbps, Gbps)
- Different from bandwidth

**Packet Loss**
- Percentage of packets not reaching receiver
- Expressed as percentage (e.g., 0.1% = 1 out of 1000 packets)
- Affects performance and reliability

### Advanced Concepts

**0-RTT (Zero Round-Trip Time)**
- Ability to send application data in first packet
- Available only for repeat connections
- Trade-off between performance and security

**Connection Migration**
- QUIC connection's ability to survive IP address changes
- Useful for mobile devices
- Uses Connection ID for identification

**Multiplexing**
- Transmission of multiple data streams over single connection
- Avoids head-of-line blocking characteristic of HTTP/2 over TCP

## TCP vs QUIC: Key Differences

### Architectural Differences

| Aspect | TCP | QUIC |
|--------|-----|------|
| **Transport Layer** | Separate protocol | Over UDP |
| **Security** | Optional (TLS) | Built-in (TLS 1.3) |
| **Connection Setup** | 3-way handshake + TLS | Integrated handshake |
| **Streams** | Single data stream | Multiple streams |
| **Head-of-line blocking** | Present | Absent |

### Performance

**TCP Disadvantages:**
- **Head-of-line blocking**: single packet loss blocks entire stream
- **Slow start**: each new connection starts at low speed
- **Multiple RTTs**: separate handshakes for TCP and TLS
- **No multiplexing**: single data stream per connection

**QUIC Advantages:**
- **Independent streams**: loss in one stream doesn't affect others
- **0-RTT**: instant resumption for known servers
- **Built-in encryption**: security by default
- **Connection migration**: resilience to network changes

### Use Cases

**TCP is better for:**
- Simple applications without low-latency requirements
- Scenarios where compatibility with legacy equipment is important
- Single-stream applications

**QUIC is better for:**
- Web applications (HTTP/3)
- Real-time communications
- Mobile applications
- Applications with multiple resources

## QUIC Architecture

### Protocol Layers

```
┌─────────────────────────────────────┐
│           HTTP/3 / Application      │
├─────────────────────────────────────┤
│              QUIC Transport         │
│  ┌─────────────┬─────────────────┐  │
│  │   Streams   │   Connection    │  │
│  │             │   Management    │  │
│  └─────────────┴─────────────────┘  │
├─────────────────────────────────────┤
│              TLS 1.3                │
├─────────────────────────────────────┤
│                UDP                  │
├─────────────────────────────────────┤
│                IP                   │
└─────────────────────────────────────┘
```

### QUIC Frame Types

**Connection-level frames:**
- `CONNECTION_CLOSE`: connection termination
- `PING`: connection liveness check
- `PADDING`: packet padding

**Stream-level frames:**
- `STREAM`: stream data
- `RESET_STREAM`: stream reset
- `STOP_SENDING`: stop sending

**Flow control frames:**
- `MAX_DATA`: connection data maximum
- `MAX_STREAM_DATA`: stream data maximum
- `DATA_BLOCKED`: blocking notification

## Congestion Control

### Congestion Control Algorithms

**CUBIC (Linux TCP default)**
- Based on cubic window growth function
- Aggressive in high-speed networks
- Can be inefficient in high-latency networks

**BBR (Bottleneck Bandwidth and RTT)**
- Developed by Google
- Estimates bandwidth and RTT
- Better performance in lossy networks

**BBRv2 (improved version)**
- Better fairness between flows
- Better adaptation to network changes
- Reduced buffering

**BBRv3 (experimental)**
- Dual-scale bandwidth model
- 2% loss threshold
- Adaptive pacing gains

### Testing Algorithms

```bash
# Test with CUBIC
./quic-test --mode=client --cc=cubic --duration=60s

# Test with BBRv2
./quic-test --mode=client --cc=bbrv2 --duration=60s

# Compare algorithms
./quic-test --mode=client --compare-cc=cubic,bbrv2 --duration=120s
```

## RFC Documents

### Core QUIC RFCs

**[RFC 9000](https://tools.ietf.org/html/rfc9000) - QUIC: A UDP-Based Multiplexed and Secure Transport**
- Main QUIC protocol specification
- Describes architecture, frames, connection management
- **Must read**

**[RFC 9001](https://tools.ietf.org/html/rfc9001) - Using TLS to Secure QUIC**
- TLS 1.3 integration with QUIC
- Cryptographic aspects
- Handshake procedures

**[RFC 9002](https://tools.ietf.org/html/rfc9002) - QUIC Loss Detection and Congestion Control**
- Loss detection algorithms
- Congestion control
- Performance metrics

### HTTP/3 and Related RFCs

**[RFC 9114](https://tools.ietf.org/html/rfc9114) - HTTP/3**
- HTTP over QUIC
- Mapping HTTP semantics to QUIC streams
- Server push and other features

**[RFC 9204](https://tools.ietf.org/html/rfc9204) - QPACK: Field Compression for HTTP/3**
- Header compression for HTTP/3
- HPACK replacement from HTTP/2
- Adaptation for multiple streams

### Additional RFCs

**[RFC 9221](https://tools.ietf.org/html/rfc9221) - An Unreliable Datagram Extension to QUIC**
- Unreliable datagram support
- For real-time applications
- WebRTC over QUIC

**[RFC 9368](https://tools.ietf.org/html/rfc9368) - Compatible Version Negotiation for QUIC**
- Protocol version negotiation
- Backward compatibility
- Protocol evolution

### Internet Drafts (in development)

**WebTransport over QUIC**
- [draft-ietf-webtrans-http3](https://datatracker.ietf.org/doc/draft-ietf-webtrans-http3/)
- Web API for QUIC connections
- WebSocket alternative

**MASQUE (Multiplexed Application Substrate over QUIC Encryption)**
- [draft-ietf-masque-connect-udp](https://datatracker.ietf.org/doc/draft-ietf-masque-connect-udp/)
- VPN and proxy over QUIC
- Traffic tunneling

## Practical Assignments

### Lab #1: QUIC Basics

**Goal:** Understand basic QUIC concepts and compare with TCP

**Tasks:**
1. Start QUIC server and client
2. Measure RTT for QUIC and TCP
3. Compare connection establishment time
4. Analyze traffic in Wireshark

```bash
# Start server
./quic-test --mode=server --port=4433

# Test QUIC
./quic-test --mode=client --server=localhost:4433 --duration=30s

# Test TCP for comparison
./quic-test --mode=client --server=localhost:4433 --tcp --duration=30s
```

### Lab #2: Congestion Control

**Goal:** Study different congestion control algorithms

**Tasks:**
1. Test CUBIC, BBR, BBRv2
2. Emulate various network conditions
3. Plot performance graphs
4. Analyze behavior under loss conditions

```bash
# Test with mobile network emulation
./quic-test --profile=mobile --cc=bbrv2 --duration=60s

# Test with high loss
./quic-test --emulate-loss=0.05 --cc=cubic --duration=60s
```

### Lab #3: Multiplexing

**Goal:** Understand multiple streams advantages

**Tasks:**
1. Create test with multiple streams
2. Compare with HTTP/2 over TCP
3. Emulate packet losses
4. Measure head-of-line blocking impact

```bash
# Test with multiple streams
./quic-test --mode=client --streams=8 --connections=2 --duration=60s

# GUI interface for visualization
./quic-gui --addr=:8080
```

### Lab #4: 0-RTT and Connection Migration

**Goal:** Study advanced QUIC features

**Tasks:**
1. Test 0-RTT connections
2. Emulate network switching (connection migration)
3. Measure connection recovery time
4. Analyze 0-RTT security

## Additional Resources

### Books and Articles

**"HTTP/3 Explained" by Daniel Stenberg**
- Free book about HTTP/3 and QUIC
- [https://http3-explained.haxx.se/](https://http3-explained.haxx.se/)

**"The Road to QUIC" by Google**
- QUIC development history
- Implementation technical details

### Analysis Tools

**Wireshark**
- QUIC traffic decoding support
- Handshake and data stream analysis
- [https://www.wireshark.org/](https://www.wireshark.org/)

**qlog**
- Standardized QUIC logging format
- Protocol event visualization
- [https://qlog.edm.uhasselt.be/](https://qlog.edm.uhasselt.be/)

**curl with HTTP/3**
- HTTP/3 connection testing
- Performance measurement
```bash
curl --http3 https://cloudflare-quic.com/
```

### Online Resources

**QUIC Working Group (IETF)**
- [https://datatracker.ietf.org/wg/quic/](https://datatracker.ietf.org/wg/quic/)
- Current specifications and discussions

**Cloudflare QUIC Blog**
- [https://blog.cloudflare.com/tag/quic/](https://blog.cloudflare.com/tag/quic/)
- Practical examples and use cases

**Google QUIC Documentation**
- [https://www.chromium.org/quic/](https://www.chromium.org/quic/)
- Technical documentation and research

### Test Servers

**Cloudflare QUIC Test**
- `https://cloudflare-quic.com/`
- Public server for testing

**Google QUIC Server**
- `quic.rocks:4433`
- Google experimental server

**Facebook QUIC**
- Production environment testing
- Real performance metrics

## Conclusion

QUIC represents a significant step forward in transport protocol development. Understanding its architecture, advantages, and limitations is critically important for modern network engineers and developers.

Use this guide as a starting point for deep protocol study. Practical assignments will help reinforce theoretical knowledge and gain real QUIC experience.

## Career Opportunities at CloudBridge Research

### Internship Program for MEI Students

CloudBridge Research offers a unique internship program for students studying network technologies and QUIC protocol.

**Available Positions:**

**Junior Network Engineer**
- Requirements: TCP/IP basics knowledge, quic-test experience
- Salary: 80,000 - 120,000 RUB/month
- Career growth: Senior Network Engineer in 1-2 years

**Protocol Research Developer**
- Requirements: Go experience, QUIC/HTTP3 understanding
- Salary: 120,000 - 180,000 RUB/month
- Career growth: Lead Research Engineer in 2-3 years

**DevOps/Infrastructure Engineer**
- Requirements: Docker, Kubernetes, CI/CD, monitoring
- Salary: 100,000 - 160,000 RUB/month
- Career growth: Senior DevOps Engineer in 1-2 years

**AI/ML Engineer (network optimization)**
- Requirements: Python, TensorFlow/PyTorch, network protocols knowledge
- Salary: 140,000 - 200,000 RUB/month
- Career growth: Senior AI Engineer in 2-3 years

### Preparation Stages

**Stage 1: Basic Training**
- Study quic-test platform
- Complete all lab assignments
- Participate in open-source development

**Stage 2: Specialization**
- Choose career track
- Deep dive into chosen area
- Work on real CloudBridge tasks

**Stage 3: Practice**
- Paid internship at CloudBridge
- Work on production projects
- Prepare thesis work

### Internship Conditions

**Summer Internship (3 months)**
- Stipend: 40,000 RUB/month
- Format: Hybrid (office + remote)
- Projects: Real quic-test development tasks

**Thesis Practice (6 months)**
- Stipend: 50,000 RUB/month
- Format: Full-time
- Result: Job offer upon successful defense

### Internship Contacts

- Email: careers@cloudbridge-research.ru
- GitHub: https://github.com/cloudbridge-research/quic-test
- Website: https://cloudbridge-research.ru

**Remember:** The best way to learn a protocol is to experiment with it hands-on and apply knowledge in real projects!