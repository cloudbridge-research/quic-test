// QUIC Dashboard Charts - Chart.js integration and visualization
QUICDashboard.prototype.initializeCharts = function() {
    // Initialize throughput chart
    const throughputCtx = document.getElementById('throughputChart');
    if (throughputCtx) {
        this.charts.throughput = new Chart(throughputCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Throughput (Mbps)',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.1)',
                    tension: 0.1,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Mbps'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    // Initialize RTT and packet loss chart
    const rttLossCtx = document.getElementById('rttLossChart');
    if (rttLossCtx) {
        this.charts.rttLoss = new Chart(rttLossCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'RTT (ms)',
                    data: [],
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.1)',
                    yAxisID: 'y',
                    tension: 0.1
                }, {
                    label: 'Packet Loss (%)',
                    data: [],
                    borderColor: 'rgb(255, 205, 86)',
                    backgroundColor: 'rgba(255, 205, 86, 0.1)',
                    yAxisID: 'y1',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false,
                },
                scales: {
                    x: {
                        display: true,
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    },
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                            display: true,
                            text: 'RTT (ms)'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'Packet Loss (%)'
                        },
                        grid: {
                            drawOnChartArea: false,
                        },
                    }
                }
            }
        });
    }

    // Initialize protocol comparison chart
    const protocolCtx = document.getElementById('protocolComparisonChart');
    if (protocolCtx) {
        this.charts.protocolComparison = new Chart(protocolCtx, {
            type: 'radar',
            data: {
                labels: ['Connection Setup', 'Throughput', 'Latency', 'Security', 'Reliability', 'Efficiency'],
                datasets: [{
                    label: 'QUIC',
                    data: [95, 90, 85, 95, 90, 88],
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    pointBackgroundColor: 'rgb(75, 192, 192)',
                    pointBorderColor: '#fff',
                    pointHoverBackgroundColor: '#fff',
                    pointHoverBorderColor: 'rgb(75, 192, 192)'
                }, {
                    label: 'TCP+TLS',
                    data: [70, 85, 75, 90, 95, 80],
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    pointBackgroundColor: 'rgb(255, 99, 132)',
                    pointBorderColor: '#fff',
                    pointHoverBackgroundColor: '#fff',
                    pointHoverBorderColor: 'rgb(255, 99, 132)'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                elements: {
                    line: {
                        borderWidth: 3
                    }
                },
                scales: {
                    r: {
                        angleLines: {
                            display: false
                        },
                        suggestedMin: 0,
                        suggestedMax: 100
                    }
                }
            }
        });
    }
};

QUICDashboard.prototype.updateCharts = function(metrics) {
    const now = new Date().toLocaleTimeString();
    
    // Update chart data arrays
    this.chartData.timestamps.push(now);
    if (this.chartData.timestamps.length > 20) {
        this.chartData.timestamps.shift();
    }

    // Extract metrics
    const throughput = parseFloat(metrics.aggregated?.throughput_mbps || 0);
    const rtt = parseFloat(metrics.aggregated?.rtt_current || 0);
    const packetLoss = parseFloat(metrics.aggregated?.packet_loss_rate?.replace('%', '') || 0);

    this.chartData.throughput.push(throughput);
    this.chartData.rtt.push(rtt);
    this.chartData.packetLoss.push(packetLoss);

    if (this.chartData.throughput.length > 20) {
        this.chartData.throughput.shift();
        this.chartData.rtt.shift();
        this.chartData.packetLoss.shift();
    }

    // Update throughput chart
    if (this.charts.throughput) {
        this.charts.throughput.data.labels = [...this.chartData.timestamps];
        this.charts.throughput.data.datasets[0].data = [...this.chartData.throughput];
        this.charts.throughput.update('none');
    }

    // Update RTT and packet loss chart
    if (this.charts.rttLoss) {
        this.charts.rttLoss.data.labels = [...this.chartData.timestamps];
        this.charts.rttLoss.data.datasets[0].data = [...this.chartData.rtt];
        this.charts.rttLoss.data.datasets[1].data = [...this.chartData.packetLoss];
        this.charts.rttLoss.update('none');
    }
};

QUICDashboard.prototype.updateProtocolComparison = function(quicMetrics, tcpMetrics) {
    if (!this.charts.protocolComparison) return;

    // Calculate comparison scores based on metrics
    const quicScores = [
        Math.min(100, 100 - (quicMetrics.handshakeTime || 50)), // Connection Setup
        Math.min(100, (quicMetrics.throughput || 50) * 2), // Throughput
        Math.min(100, 100 - (quicMetrics.rtt || 20)), // Latency (lower is better)
        95, // Security (QUIC has built-in TLS)
        Math.min(100, 100 - (quicMetrics.packetLoss || 1) * 10), // Reliability
        Math.min(100, (quicMetrics.efficiency || 80)) // Efficiency
    ];

    const tcpScores = [
        Math.min(100, 100 - (tcpMetrics?.handshakeTime || 80)), // Connection Setup
        Math.min(100, (tcpMetrics?.throughput || 40) * 2), // Throughput
        Math.min(100, 100 - (tcpMetrics?.rtt || 30)), // Latency
        90, // Security (separate TLS)
        Math.min(100, 100 - (tcpMetrics?.packetLoss || 2) * 10), // Reliability
        Math.min(100, (tcpMetrics?.efficiency || 75)) // Efficiency
    ];

    this.charts.protocolComparison.data.datasets[0].data = quicScores;
    this.charts.protocolComparison.data.datasets[1].data = tcpScores;
    this.charts.protocolComparison.update('none');
};