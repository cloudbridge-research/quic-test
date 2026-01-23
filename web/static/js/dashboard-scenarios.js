// QUIC Dashboard Scenarios - Educational scenario execution and management
QUICDashboard.prototype.startScenario = function(scenarioType) {
    this.activeScenario = scenarioType;
    this.log(`Starting educational scenario: ${scenarioType}`);

    // Show progress container
    const progressContainer = document.getElementById('scenario-progress');
    if (progressContainer) {
        progressContainer.style.display = 'block';
    }

    // Update scenario title
    const titleElement = document.getElementById('current-scenario-title');
    if (titleElement) {
        const scenarioNames = {
            'handshake-demo': 'QUIC Handshake Demo',
            'congestion-comparison': 'Congestion Control Comparison',
            'stream-multiplexing': 'Stream Multiplexing Demo',
            'loss-recovery': 'Loss Recovery Analysis',
            'migration-demo': 'Connection Migration Demo',
            'security-features': 'Security Features Exploration'
        };
        titleElement.textContent = scenarioNames[scenarioType] || 'Running Scenario';
    }

    // Initialize scenario steps
    this.initializeScenarioSteps(scenarioType);

    // Load appropriate preset for scenario
    const scenarioPresets = {
        'handshake-demo': 'handshake-analysis',
        'congestion-comparison': 'congestion-comparison',
        'stream-multiplexing': 'beginner',
        'loss-recovery': 'network-conditions',
        'migration-demo': 'network-conditions',
        'security-features': 'handshake-analysis'
    };

    const preset = scenarioPresets[scenarioType];
    if (preset) {
        this.loadPreset(preset);
    }

    // Start scenario execution
    this.executeScenario(scenarioType);
};

QUICDashboard.prototype.initializeScenarioSteps = function(scenarioType) {
    const stepsList = document.getElementById('scenario-steps-list');
    if (!stepsList) return;

    const scenarioSteps = {
        'handshake-demo': [
            { title: 'Initial Client Hello', description: 'Client sends connection request', duration: 2000 },
            { title: 'Server Response', description: 'Server responds with version negotiation', duration: 1000 },
            { title: 'Key Exchange', description: 'Cryptographic handshake', duration: 3000 },
            { title: 'Connection Ready', description: 'Secure connection established', duration: 1000 }
        ],
        'congestion-comparison': [
            { title: 'Initialize CUBIC', description: 'Start CUBIC congestion control test', duration: 15000 },
            { title: 'Initialize BBR', description: 'Start BBR congestion control test', duration: 15000 },
            { title: 'Network Stress', description: 'Apply network conditions', duration: 20000 },
            { title: 'Compare Results', description: 'Analyze performance differences', duration: 10000 }
        ],
        'stream-multiplexing': [
            { title: 'Single Stream', description: 'Establish single stream connection', duration: 5000 },
            { title: 'Multiple Streams', description: 'Create multiple concurrent streams', duration: 10000 },
            { title: 'Stream Priority', description: 'Demonstrate stream prioritization', duration: 8000 },
            { title: 'Flow Control', description: 'Show per-stream flow control', duration: 7000 }
        ],
        'loss-recovery': [
            { title: 'Normal Operation', description: 'Baseline packet transmission', duration: 8000 },
            { title: 'Introduce Loss', description: 'Simulate packet loss conditions', duration: 12000 },
            { title: 'Loss Detection', description: 'QUIC detects packet loss', duration: 5000 },
            { title: 'Recovery Process', description: 'Retransmission and recovery', duration: 10000 }
        ],
        'migration-demo': [
            { title: 'Initial Connection', description: 'Establish connection on primary path', duration: 5000 },
            { title: 'Path Validation', description: 'Validate alternative network path', duration: 8000 },
            { title: 'Connection Migration', description: 'Migrate to new network path', duration: 7000 },
            { title: 'Verify Migration', description: 'Confirm successful migration', duration: 5000 }
        ],
        'security-features': [
            { title: 'TLS Integration', description: 'Built-in TLS 1.3 encryption', duration: 6000 },
            { title: 'Connection ID', description: 'Privacy-preserving connection IDs', duration: 8000 },
            { title: 'Key Rotation', description: 'Automatic key updates', duration: 10000 },
            { title: 'Replay Protection', description: 'Anti-replay mechanisms', duration: 6000 }
        ]
    };

    const steps = scenarioSteps[scenarioType] || [];
    stepsList.innerHTML = '';

    steps.forEach((step, index) => {
        const stepElement = document.createElement('div');
        stepElement.className = 'step-item';
        stepElement.innerHTML = `
            <div class="step-number">${index + 1}</div>
            <div class="step-content">
                <div class="step-title">${step.title}</div>
                <div class="step-description">${step.description}</div>
            </div>
        `;
        stepsList.appendChild(stepElement);
    });

    this.scenarioSteps = steps;
};

QUICDashboard.prototype.executeScenario = function(scenarioType) {
    if (!this.scenarioSteps) return;

    let currentStep = 0;
    let totalDuration = this.scenarioSteps.reduce((sum, step) => sum + step.duration, 0);
    let elapsedTime = 0;

    const updateProgress = () => {
        const progress = (elapsedTime / totalDuration) * 100;
        const progressFill = document.getElementById('scenario-progress-fill');
        const progressText = document.getElementById('scenario-progress-text');
        
        if (progressFill) progressFill.style.width = `${progress}%`;
        if (progressText) progressText.textContent = `${Math.round(progress)}% Complete`;

        // Update step states
        const stepElements = document.querySelectorAll('.step-item');
        stepElements.forEach((element, index) => {
            element.classList.remove('active', 'completed');
            if (index < currentStep) {
                element.classList.add('completed');
            } else if (index === currentStep) {
                element.classList.add('active');
            }
        });

        // Update insights
        this.updateScenarioInsights(scenarioType, currentStep, progress);
    };

    const executeStep = (stepIndex) => {
        if (stepIndex >= this.scenarioSteps.length) {
            this.completeScenario();
            return;
        }

        currentStep = stepIndex;
        updateProgress();

        const step = this.scenarioSteps[stepIndex];
        this.log(`Executing step ${stepIndex + 1}: ${step.title}`);

        setTimeout(() => {
            elapsedTime += step.duration;
            executeStep(stepIndex + 1);
        }, step.duration);
    };

    // Start the test with scenario-specific configuration
    setTimeout(() => {
        this.startTest();
        executeStep(0);
    }, 1000);
};

QUICDashboard.prototype.updateScenarioInsights = function(scenarioType, currentStep, progress) {
    const insightsContent = document.getElementById('scenario-insights-content');
    if (!insightsContent) return;

    const insights = {
        'handshake-demo': [
            'QUIC combines transport and cryptographic handshakes for faster connection establishment.',
            'The server can respond immediately with encrypted data, reducing round trips.',
            'Key exchange uses modern cryptographic algorithms for enhanced security.',
            'Connection is now ready for high-performance data transmission.'
        ],
        'congestion-comparison': [
            'CUBIC uses a cubic function to determine congestion window growth.',
            'BBR focuses on bandwidth and RTT measurements for optimal throughput.',
            'Network conditions significantly impact algorithm performance.',
            'BBR typically shows better performance in high-bandwidth, high-latency networks.'
        ],
        'stream-multiplexing': [
            'Single stream establishes the baseline connection performance.',
            'Multiple streams eliminate head-of-line blocking issues.',
            'Stream prioritization ensures critical data is transmitted first.',
            'Per-stream flow control prevents any single stream from overwhelming the connection.'
        ],
        'loss-recovery': [
            'Normal operation shows optimal packet transmission patterns.',
            'Packet loss simulation demonstrates real-world network conditions.',
            'QUIC uses sophisticated algorithms to detect packet loss quickly.',
            'Recovery mechanisms ensure reliable data delivery despite network issues.'
        ],
        'migration-demo': [
            'Initial connection establishes the primary communication path.',
            'Path validation ensures the alternative route is viable.',
            'Connection migration maintains session continuity during network changes.',
            'Migration verification confirms successful transition without data loss.'
        ],
        'security-features': [
            'TLS 1.3 integration provides mandatory encryption for all QUIC connections.',
            'Connection IDs protect user privacy by preventing connection tracking.',
            'Automatic key rotation enhances security without interrupting data flow.',
            'Replay protection prevents malicious packet reuse attacks.'
        ]
    };

    const scenarioInsights = insights[scenarioType] || [];
    if (currentStep < scenarioInsights.length) {
        insightsContent.textContent = scenarioInsights[currentStep];
    }
};

QUICDashboard.prototype.completeScenario = function() {
    this.log(`Scenario "${this.activeScenario}" completed successfully`);
    
    const progressFill = document.getElementById('scenario-progress-fill');
    const progressText = document.getElementById('scenario-progress-text');
    
    if (progressFill) progressFill.style.width = '100%';
    if (progressText) progressText.textContent = '100% Complete';

    // Mark all steps as completed
    const stepElements = document.querySelectorAll('.step-item');
    stepElements.forEach(element => {
        element.classList.remove('active');
        element.classList.add('completed');
    });

    // Show completion message
    const insightsContent = document.getElementById('scenario-insights-content');
    if (insightsContent) {
        insightsContent.innerHTML = `
            <strong>Scenario Complete!</strong><br>
            The ${this.activeScenario.replace('-', ' ')} scenario has finished successfully. 
            Review the metrics and charts above to analyze the results. 
            You can export the results or start another scenario.
        `;
    }

    // Enable export button
    const exportButton = document.getElementById('export-scenario');
    if (exportButton) {
        exportButton.disabled = false;
    }
};

QUICDashboard.prototype.pauseScenario = function() {
    // Implementation for pausing scenario
    this.log('Scenario paused');
};

QUICDashboard.prototype.stopScenario = function() {
    this.activeScenario = null;
    this.scenarioSteps = null;
    
    const progressContainer = document.getElementById('scenario-progress');
    if (progressContainer) {
        progressContainer.style.display = 'none';
    }
    
    this.stopAll();
    this.log('Scenario stopped');
};

QUICDashboard.prototype.exportScenarioResults = function() {
    if (!this.activeScenario) return;

    const results = {
        scenario: this.activeScenario,
        timestamp: new Date().toISOString(),
        metrics: this.chartData,
        config: this.getConfig()
    };

    const blob = new Blob([JSON.stringify(results, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `scenario-${this.activeScenario}-${new Date().toISOString().slice(0, 19)}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    this.log('Scenario results exported');
};