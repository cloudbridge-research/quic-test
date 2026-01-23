// QUIC Dashboard Core - Main dashboard class and initialization
class QUICDashboard {
    constructor() {
        this.updateInterval = null;
        this.currentPreset = null;
        this.activeScenario = null;
        this.scenarioProgress = null;
        this.charts = {};
        this.chartData = {
            throughput: [],
            rtt: [],
            packetLoss: [],
            timestamps: []
        };
        this.captureActive = false;
        this.init();
    }

    init() {
        this.bindEvents();
        this.initializeTabs();
        this.initializeCharts();
        this.startPeriodicUpdates();
        this.log('Dashboard initialized successfully');
    }

    bindEvents() {
        // Server controls
        document.getElementById('server-start').addEventListener('click', () => this.startServer());
        document.getElementById('server-stop').addEventListener('click', () => this.stopServer());
        
        // Client controls
        document.getElementById('client-start').addEventListener('click', () => this.startClient());
        document.getElementById('client-stop').addEventListener('click', () => this.stopClient());
        
        // Test controls
        document.getElementById('test-start').addEventListener('click', () => this.startTest());
        document.getElementById('test-stop').addEventListener('click', () => this.stopAll());

        // Preset controls
        document.querySelectorAll('.preset-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const presetCard = e.target.closest('.preset-card');
                const presetType = presetCard.dataset.preset;
                this.loadPreset(presetType);
            });
        });

        // Scenario controls
        document.querySelectorAll('.scenario-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const scenarioCard = e.target.closest('.scenario-card');
                const scenarioType = scenarioCard.dataset.scenario;
                this.startScenario(scenarioType);
            });
        });

        // Network profile change
        const networkProfile = document.getElementById('network-profile');
        if (networkProfile) {
            networkProfile.addEventListener('change', (e) => this.applyNetworkProfile(e.target.value));
        }

        // Log controls
        const clearLogs = document.getElementById('clear-logs');
        if (clearLogs) {
            clearLogs.addEventListener('click', () => this.clearLogs());
        }

        const exportLogs = document.getElementById('export-logs');
        if (exportLogs) {
            exportLogs.addEventListener('click', () => this.exportLogs());
        }

        // Scenario progress controls
        const pauseScenario = document.getElementById('pause-scenario');
        if (pauseScenario) {
            pauseScenario.addEventListener('click', () => this.pauseScenario());
        }

        const stopScenario = document.getElementById('stop-scenario');
        if (stopScenario) {
            stopScenario.addEventListener('click', () => this.stopScenario());
        }

        const exportScenario = document.getElementById('export-scenario');
        if (exportScenario) {
            exportScenario.addEventListener('click', () => this.exportScenarioResults());
        }

        // Capture controls
        const startCapture = document.getElementById('start-capture');
        if (startCapture) {
            startCapture.addEventListener('click', () => this.startCapture());
        }

        const stopCapture = document.getElementById('stop-capture');
        if (stopCapture) {
            stopCapture.addEventListener('click', () => this.stopCapture());
        }

        const downloadPcap = document.getElementById('download-pcap');
        if (downloadPcap) {
            downloadPcap.addEventListener('click', () => this.downloadPcap());
        }

        // Report generation
        const generateReport = document.getElementById('generate-report');
        if (generateReport) {
            generateReport.addEventListener('click', () => this.generateReport());
        }

        const downloadReport = document.getElementById('download-report');
        if (downloadReport) {
            downloadReport.addEventListener('click', () => this.downloadReport());
        }

        // Data export
        const exportCsv = document.getElementById('export-csv');
        if (exportCsv) {
            exportCsv.addEventListener('click', () => this.exportData('csv'));
        }

        const exportJson = document.getElementById('export-json');
        if (exportJson) {
            exportJson.addEventListener('click', () => this.exportData('json'));
        }

        const exportConfig = document.getElementById('export-config');
        if (exportConfig) {
            exportConfig.addEventListener('click', () => this.exportData('config'));
        }
    }

    initializeTabs() {
        // Configuration tabs
        document.querySelectorAll('.config-tabs .tab-button').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                this.switchConfigTab(tabName);
            });
        });

        // Metrics tabs
        document.querySelectorAll('.metrics-tabs .tab-button').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                this.switchMetricsTab(tabName);
            });
        });
    }

    switchConfigTab(tabName) {
        // Remove active class from all config tabs
        document.querySelectorAll('.config-tabs .tab-button').forEach(btn => btn.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

        // Add active class to selected tab
        document.querySelector(`.config-tabs .tab-button[data-tab="${tabName}"]`).classList.add('active');
        document.getElementById(`${tabName}-tab`).classList.add('active');
    }

    switchMetricsTab(tabName) {
        // Remove active class from all metrics tabs
        document.querySelectorAll('.metrics-tabs .tab-button').forEach(btn => btn.classList.remove('active'));
        document.querySelectorAll('#overview-metrics, #detailed-metrics, #protocol-metrics, #educational-metrics').forEach(content => content.classList.remove('active'));

        // Add active class to selected tab
        document.querySelector(`.metrics-tabs .tab-button[data-tab="${tabName}"]`).classList.add('active');
        document.getElementById(`${tabName}-metrics`).classList.add('active');
    }

    log(message) {
        const logs = document.getElementById('logs');
        const timestamp = new Date().toLocaleTimeString();
        logs.textContent += `\n[${timestamp}] ${message}`;
        logs.scrollTop = logs.scrollHeight;
    }

    clearLogs() {
        document.getElementById('logs').textContent = 'Logs cleared.';
    }

    exportLogs() {
        const logs = document.getElementById('logs').textContent;
        const blob = new Blob([logs], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `quic-dashboard-logs-${new Date().toISOString().slice(0, 19)}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        this.log('Logs exported successfully');
    }

    async makeRequest(url, options = {}) {
        try {
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });
            return await response.json();
        } catch (error) {
            this.log(`Request error: ${error.message}`);
            throw error;
        }
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    startPeriodicUpdates() {
        this.updateStatus();
        this.updateInterval = setInterval(() => this.updateStatus(), 2000);
    }

    stopPeriodicUpdates() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
            this.updateInterval = null;
        }
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new QUICDashboard();
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (window.dashboard) {
        window.dashboard.stopPeriodicUpdates();
    }
});