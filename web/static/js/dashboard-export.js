// QUIC Dashboard Export - Capture, reporting, and data export functionality

// Capture and export methods
QUICDashboard.prototype.startCapture = function() {
    this.captureActive = true;
    
    const startBtn = document.getElementById('start-capture');
    const stopBtn = document.getElementById('stop-capture');
    const statusText = document.getElementById('capture-status-text');
    
    if (startBtn) startBtn.disabled = true;
    if (stopBtn) stopBtn.disabled = false;
    if (statusText) statusText.textContent = 'Capturing packets...';
    
    this.log('Packet capture started');
    
    // Simulate packet counting
    this.captureInterval = setInterval(() => {
        const sizeElement = document.getElementById('capture-size');
        if (sizeElement && this.captureActive) {
            const currentCount = parseInt(sizeElement.textContent) || 0;
            sizeElement.textContent = `${currentCount + Math.floor(Math.random() * 10) + 1} packets`;
        }
    }, 1000);
};

QUICDashboard.prototype.stopCapture = function() {
    this.captureActive = false;
    
    if (this.captureInterval) {
        clearInterval(this.captureInterval);
    }
    
    const startBtn = document.getElementById('start-capture');
    const stopBtn = document.getElementById('stop-capture');
    const downloadBtn = document.getElementById('download-pcap');
    const statusText = document.getElementById('capture-status-text');
    
    if (startBtn) startBtn.disabled = false;
    if (stopBtn) stopBtn.disabled = true;
    if (downloadBtn) downloadBtn.disabled = false;
    if (statusText) statusText.textContent = 'Capture complete';
    
    this.log('Packet capture stopped');
};

QUICDashboard.prototype.downloadPcap = function() {
    // Simulate PCAP download
    const blob = new Blob(['# Simulated PCAP data\n# This would contain actual packet capture data'], 
                         { type: 'application/octet-stream' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `quic-capture-${new Date().toISOString().slice(0, 19)}.pcap`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    this.log('PCAP file downloaded');
};

QUICDashboard.prototype.generateReport = function() {
    const includeCharts = document.getElementById('include-charts')?.checked || false;
    const includeAnalysis = document.getElementById('include-analysis')?.checked || false;
    const includeRecommendations = document.getElementById('include-recommendations')?.checked || false;
    
    this.log('Generating PDF report...');
    
    // Simulate report generation
    setTimeout(() => {
        const downloadBtn = document.getElementById('download-report');
        if (downloadBtn) downloadBtn.disabled = false;
        this.log('PDF report generated successfully');
    }, 3000);
};

QUICDashboard.prototype.downloadReport = function() {
    // Simulate PDF download
    const reportContent = this.generateReportContent();
    const blob = new Blob([reportContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `quic-educational-report-${new Date().toISOString().slice(0, 19)}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    this.log('Educational report downloaded');
};

QUICDashboard.prototype.generateReportContent = function() {
    return `QUIC Educational Dashboard Report
Generated: ${new Date().toLocaleString()}

=== Test Configuration ===
${JSON.stringify(this.getConfig(), null, 2)}

=== Performance Metrics ===
Throughput Data: ${JSON.stringify(this.chartData.throughput)}
RTT Data: ${JSON.stringify(this.chartData.rtt)}
Packet Loss Data: ${JSON.stringify(this.chartData.packetLoss)}

=== Educational Analysis ===
This report contains analysis of QUIC protocol performance
under various network conditions and configurations.

Key findings:
- QUIC shows improved connection establishment times
- Stream multiplexing eliminates head-of-line blocking
- Advanced congestion control adapts to network conditions
- Built-in security features provide robust protection

=== Recommendations ===
1. Use 0-RTT for returning clients to reduce latency
2. Enable connection migration for mobile scenarios
3. Tune congestion control based on network characteristics
4. Monitor packet loss and adjust FEC accordingly

This is a simulated report. In a real implementation,
this would contain detailed charts, graphs, and analysis.`;
};

QUICDashboard.prototype.exportData = function(format) {
    const data = {
        timestamp: new Date().toISOString(),
        metrics: this.chartData,
        config: this.getConfig(),
        scenario: this.activeScenario
    };

    let content, filename, mimeType;

    switch (format) {
        case 'csv':
            content = this.convertToCSV(data);
            filename = `quic-data-${new Date().toISOString().slice(0, 19)}.csv`;
            mimeType = 'text/csv';
            break;
        case 'json':
            content = JSON.stringify(data, null, 2);
            filename = `quic-data-${new Date().toISOString().slice(0, 19)}.json`;
            mimeType = 'application/json';
            break;
        case 'config':
            content = JSON.stringify(this.getConfig(), null, 2);
            filename = `quic-config-${new Date().toISOString().slice(0, 19)}.json`;
            mimeType = 'application/json';
            break;
        default:
            return;
    }

    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    this.log(`Data exported as ${format.toUpperCase()}`);
};

QUICDashboard.prototype.convertToCSV = function(data) {
    const headers = ['Timestamp', 'Throughput (Mbps)', 'RTT (ms)', 'Packet Loss (%)'];
    const rows = [headers.join(',')];
    
    const maxLength = Math.max(
        data.metrics.timestamps?.length || 0,
        data.metrics.throughput?.length || 0,
        data.metrics.rtt?.length || 0,
        data.metrics.packetLoss?.length || 0
    );
    
    for (let i = 0; i < maxLength; i++) {
        const row = [
            data.metrics.timestamps?.[i] || '',
            data.metrics.throughput?.[i] || '',
            data.metrics.rtt?.[i] || '',
            data.metrics.packetLoss?.[i] || ''
        ];
        rows.push(row.join(','));
    }
    
    return rows.join('\n');
};