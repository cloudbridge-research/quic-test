// QUIC Dashboard API - Server communication and data management
QUICDashboard.prototype.loadPreset = function(presetType) {
    this.currentPreset = presetType;
    this.log(`Loading preset: ${presetType}`);

    const presets = {
        'beginner': {
            connections: 1,
            streams: 1,
            'packet-size': 1200,
            rate: 50,
            duration: 30,
            'congestion-control': 'cubic',
            'packet-loss': 0,
            latency: 0,
            'enable-0rtt': false,
            'no-tls': true
        },
        'congestion-comparison': {
            connections: 2,
            streams: 4,
            'packet-size': 1200,
            rate: 200,
            duration: 120,
            'congestion-control': 'bbr',
            'packet-loss': 1,
            latency: 50,
            'enable-0rtt': true,
            'no-tls': false
        },
        'handshake-analysis': {
            connections: 1,
            streams: 1,
            'packet-size': 64,
            rate: 10,
            duration: 60,
            'congestion-control': 'cubic',
            'enable-0rtt': true,
            'enable-key-update': true,
            'no-tls': false
        },
        'network-conditions': {
            connections: 3,
            streams: 6,
            'packet-size': 1200,
            rate: 300,
            duration: 180,
            'congestion-control': 'bbrv3',
            'packet-loss': 5,
            latency: 100,
            'packet-dup': 1,
            'no-tls': false
        }
    };

    const preset = presets[presetType];
    if (preset) {
        Object.keys(preset).forEach(key => {
            const element = document.getElementById(key);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = preset[key];
                } else {
                    element.value = preset[key];
                }
            }
        });
        this.log(`Preset "${presetType}" loaded successfully`);
    }
};

QUICDashboard.prototype.applyNetworkProfile = function(profile) {
    const profiles = {
        'perfect': { loss: 0, latency: 0, bandwidth: 0 },
        'ethernet': { loss: 0, latency: 1, bandwidth: 1000 },
        'wifi': { loss: 0.1, latency: 5, bandwidth: 100 },
        '4g': { loss: 1, latency: 50, bandwidth: 50 },
        '3g': { loss: 3, latency: 150, bandwidth: 10 },
        'satellite': { loss: 0.5, latency: 600, bandwidth: 25 }
    };

    const config = profiles[profile];
    if (config) {
        document.getElementById('packet-loss').value = config.loss;
        document.getElementById('latency').value = config.latency;
        document.getElementById('bandwidth-limit').value = config.bandwidth;
        this.log(`Applied network profile: ${profile}`);
    }
};

QUICDashboard.prototype.getConfig = function() {
    const config = {
        addr: document.getElementById('addr').value,
        connections: parseInt(document.getElementById('connections').value),
        streams: parseInt(document.getElementById('streams').value),
        packet_size: parseInt(document.getElementById('packet-size').value),
        rate: parseInt(document.getElementById('rate').value),
        duration: parseInt(document.getElementById('duration').value) + 's',
        pattern: document.getElementById('pattern').value,
        prometheus: document.getElementById('enable-prometheus')?.checked || true,
        no_tls: document.getElementById('no-tls')?.checked || true
    };

    // Add QUIC-specific settings
    const congestionControl = document.getElementById('congestion-control');
    if (congestionControl) config.congestion_control = congestionControl.value;

    const maxIdleTimeout = document.getElementById('max-idle-timeout');
    if (maxIdleTimeout) config.max_idle_timeout = parseInt(maxIdleTimeout.value);

    const handshakeTimeout = document.getElementById('handshake-timeout');
    if (handshakeTimeout) config.handshake_timeout = parseInt(handshakeTimeout.value);

    const maxStreamData = document.getElementById('max-stream-data');
    if (maxStreamData) config.max_stream_data = parseInt(maxStreamData.value) * 1024;

    const keepAlive = document.getElementById('keep-alive');
    if (keepAlive) config.keep_alive = parseInt(keepAlive.value);

    const enable0RTT = document.getElementById('enable-0rtt');
    if (enable0RTT) config.enable_0rtt = enable0RTT.checked;

    const enableKeyUpdate = document.getElementById('enable-key-update');
    if (enableKeyUpdate) config.enable_key_update = enableKeyUpdate.checked;

    const enableDatagrams = document.getElementById('enable-datagrams');
    if (enableDatagrams) config.enable_datagrams = enableDatagrams.checked;

    // Add network emulation settings
    const packetLoss = document.getElementById('packet-loss');
    if (packetLoss) config.emulate_loss = parseFloat(packetLoss.value) / 100;

    const latency = document.getElementById('latency');
    if (latency) config.emulate_latency = parseInt(latency.value) + 'ms';

    const packetDup = document.getElementById('packet-dup');
    if (packetDup) config.emulate_dup = parseFloat(packetDup.value) / 100;

    // Add advanced settings
    const fecRedundancy = document.getElementById('fec-redundancy');
    const enableFEC = document.getElementById('enable-fec');
    if (enableFEC?.checked && fecRedundancy) {
        config.fec_enabled = true;
        config.fec_redundancy = parseFloat(fecRedundancy.value) / 100;
    }

    const enablePQC = document.getElementById('enable-pqc');
    const pqcAlgorithm = document.getElementById('pqc-algorithm');
    if (enablePQC?.checked && pqcAlgorithm) {
        config.pqc_enabled = true;
        config.pqc_algorithm = pqcAlgorithm.value;
    }

    const enableAIRouting = document.getElementById('enable-ai-routing');
    if (enableAIRouting?.checked) {
        config.ai_enabled = true;
        config.ai_service_url = 'http://localhost:5000';
    }

    return config;
};

QUICDashboard.prototype.updateStatus = function() {
    try {
        this.makeRequest('/api/quic/status')
            .then(data => {
                if (data.success) {
                    this.updateProcessStatus('server', data.data.server);
                    this.updateProcessStatus('client', data.data.client);
                    this.updateMetrics(data.data.metrics);
                    this.updateEducationalInsights(data.data.metrics);
                }
            })
            .catch(error => {
                this.log(`Error updating status: ${error.message}`);
            });
    } catch (error) {
        this.log(`Error updating status: ${error.message}`);
    }
};

QUICDashboard.prototype.updateProcessStatus = function(type, status) {
    const statusElement = document.getElementById(`${type}-status`);
    
    if (status.running) {
        statusElement.textContent = `Running (PID: ${status.pid})`;
        statusElement.className = 'status-indicator status-running';
    } else {
        statusElement.textContent = 'Stopped';
        statusElement.className = 'status-indicator status-stopped';
    }
};

QUICDashboard.prototype.updateMetrics = function(metrics) {
    if (metrics && metrics.aggregated) {
        const agg = metrics.aggregated;
        
        // Overview metrics
        document.getElementById('packets-sent').textContent = agg.total_packets_sent || '0';
        document.getElementById('packets-received').textContent = agg.total_packets_received || '0';
        document.getElementById('bytes-sent').textContent = this.formatBytes(agg.total_bytes_sent || 0);
        document.getElementById('throughput').textContent = `${agg.throughput_mbps || '0'} Mbps`;

        // Detailed metrics
        const rttCurrent = document.getElementById('rtt-current');
        if (rttCurrent) rttCurrent.textContent = `${agg.rtt_current || '0'} ms`;
        
        const rttMin = document.getElementById('rtt-min');
        if (rttMin) rttMin.textContent = `${agg.rtt_min || '0'} ms`;
        
        const packetLossRate = document.getElementById('packet-loss-rate');
        if (packetLossRate) packetLossRate.textContent = agg.packet_loss_rate || '0%';
        
        const retransmissions = document.getElementById('retransmissions');
        if (retransmissions) retransmissions.textContent = agg.retransmissions || '0';
        
        const congestionWindow = document.getElementById('congestion-window');
        if (congestionWindow) congestionWindow.textContent = `${agg.congestion_window || '0'} KB`;
        
        const flowControlWindow = document.getElementById('flow-control-window');
        if (flowControlWindow) flowControlWindow.textContent = `${agg.flow_control_window || '0'} KB`;

        // Protocol metrics
        const handshakeTime = document.getElementById('handshake-time');
        if (handshakeTime) handshakeTime.textContent = `${agg.handshake_time || '0'} ms`;
        
        const zeroRttSuccess = document.getElementById('zero-rtt-success');
        if (zeroRttSuccess) zeroRttSuccess.textContent = `${agg.zero_rtt_success || '0'}%`;
        
        const keyUpdates = document.getElementById('key-updates');
        if (keyUpdates) keyUpdates.textContent = agg.key_updates || '0';
        
        const streamCreationTime = document.getElementById('stream-creation-time');
        if (streamCreationTime) streamCreationTime.textContent = `${agg.stream_creation_time || '0'} ms`;
        
        const connectionMigrations = document.getElementById('connection-migrations');
        if (connectionMigrations) connectionMigrations.textContent = agg.connection_migrations || '0';
        
        const datagramSuccess = document.getElementById('datagram-success');
        if (datagramSuccess) datagramSuccess.textContent = `${agg.datagram_success || '0'}%`;

        // Update charts with new data
        this.updateCharts(metrics);
    }
};

QUICDashboard.prototype.updateEducationalInsights = function(metrics) {
    if (!metrics || !metrics.aggregated) return;

    const agg = metrics.aggregated;
    
    // Connection quality calculation
    let quality = 100;
    const rtt = parseFloat(agg.rtt_current) || 0;
    const loss = parseFloat(agg.packet_loss_rate) || 0;
    const throughput = parseFloat(agg.throughput_mbps) || 0;

    if (rtt > 100) quality -= 20;
    else if (rtt > 50) quality -= 10;
    
    if (loss > 5) quality -= 30;
    else if (loss > 1) quality -= 15;
    
    if (throughput < 1) quality -= 20;
    else if (throughput < 10) quality -= 10;

    quality = Math.max(0, Math.min(100, quality));
    
    const qualityFill = document.querySelector('.quality-fill');
    const qualityText = document.querySelector('.quality-text');
    if (qualityFill && qualityText) {
        qualityFill.style.width = `${quality}%`;
        let qualityLabel = 'Poor';
        if (quality > 80) qualityLabel = 'Excellent';
        else if (quality > 60) qualityLabel = 'Good';
        else if (quality > 40) qualityLabel = 'Fair';
        
        qualityText.textContent = `${qualityLabel} (${quality}%)`;
    }

    // Congestion control efficiency
    const ccEfficiency = Math.min(100, Math.max(0, 100 - (rtt / 10) - (loss * 10)));
    const meterValue = document.querySelector('.meter-value');
    if (meterValue) {
        meterValue.textContent = `${Math.round(ccEfficiency)}%`;
    }
};

// Server and client control methods
QUICDashboard.prototype.startServer = function() {
    try {
        const config = this.getConfig();
        config.mode = 'server';
        config.addr = ':9000';
        
        this.makeRequest('/api/quic/server/start', {
            method: 'POST',
            body: JSON.stringify(config)
        }).then(data => {
            if (data.success) {
                this.log('Server started successfully');
                if (this.currentPreset) {
                    this.log(`Using preset: ${this.currentPreset}`);
                }
            } else {
                this.log(`Failed to start server: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Server start error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Server start error: ${error.message}`);
    }
};

QUICDashboard.prototype.stopServer = function() {
    try {
        this.makeRequest('/api/quic/server/stop', {
            method: 'POST'
        }).then(data => {
            if (data.success) {
                this.log('Server stopped successfully');
            } else {
                this.log(`Failed to stop server: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Server stop error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Server stop error: ${error.message}`);
    }
};

QUICDashboard.prototype.startClient = function() {
    try {
        const config = this.getConfig();
        config.mode = 'client';
        
        this.makeRequest('/api/quic/client/start', {
            method: 'POST',
            body: JSON.stringify(config)
        }).then(data => {
            if (data.success) {
                this.log('Client started successfully');
                if (this.activeScenario) {
                    this.log(`Running scenario: ${this.activeScenario}`);
                }
            } else {
                this.log(`Failed to start client: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Client start error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Client start error: ${error.message}`);
    }
};

QUICDashboard.prototype.stopClient = function() {
    try {
        this.makeRequest('/api/quic/client/stop', {
            method: 'POST'
        }).then(data => {
            if (data.success) {
                this.log('Client stopped successfully');
            } else {
                this.log(`Failed to stop client: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Client stop error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Client stop error: ${error.message}`);
    }
};

QUICDashboard.prototype.startTest = function() {
    try {
        const config = this.getConfig();
        config.mode = 'test';
        
        this.makeRequest('/api/quic/test/start', {
            method: 'POST',
            body: JSON.stringify(config)
        }).then(data => {
            if (data.success) {
                this.log('Integrated test started successfully');
                if (this.activeScenario) {
                    this.log(`Educational scenario "${this.activeScenario}" is now active`);
                }
            } else {
                this.log(`Failed to start test: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Test start error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Test start error: ${error.message}`);
    }
};

QUICDashboard.prototype.stopAll = function() {
    try {
        this.makeRequest('/api/quic/test/stop', {
            method: 'POST'
        }).then(data => {
            if (data.success) {
                this.log('All processes stopped successfully');
                if (this.activeScenario) {
                    this.log(`Scenario "${this.activeScenario}" completed`);
                    this.activeScenario = null;
                }
            } else {
                this.log(`Failed to stop processes: ${data.error}`);
            }
        }).catch(error => {
            this.log(`Stop all error: ${error.message}`);
        });
    } catch (error) {
        this.log(`Stop all error: ${error.message}`);
    }
};