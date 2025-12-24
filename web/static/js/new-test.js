// QUIC Test Suite - New Test Form JavaScript

class NewTestForm {
    constructor() {
        this.form = document.getElementById('test-form');
        this.presetModal = document.getElementById('preset-modal');
        this.presets = null;
        this.profiles = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadPresets();
        this.setupFormValidation();
        this.setupDependentFields();
    }

    setupEventListeners() {
        // Form submission
        if (this.form) {
            this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        }

        // Load preset button
        const loadPresetBtn = document.getElementById('load-preset');
        if (loadPresetBtn) {
            loadPresetBtn.addEventListener('click', () => this.showPresetModal());
        }

        // Modal close
        const modalClose = document.querySelector('.modal-close');
        if (modalClose) {
            modalClose.addEventListener('click', () => this.hidePresetModal());
        }

        // Close modal on background click
        if (this.presetModal) {
            this.presetModal.addEventListener('click', (e) => {
                if (e.target === this.presetModal) {
                    this.hidePresetModal();
                }
            });
        }

        // Mode change handler
        const modeSelect = document.getElementById('mode');
        if (modeSelect) {
            modeSelect.addEventListener('change', () => this.handleModeChange());
        }

        // FEC enable/disable
        const fecEnabledCheckbox = document.getElementById('fec-enabled');
        if (fecEnabledCheckbox) {
            fecEnabledCheckbox.addEventListener('change', () => this.handleFECToggle());
        }

        // PQC enable/disable
        const pqcEnabledCheckbox = document.getElementById('pqc-enabled');
        if (pqcEnabledCheckbox) {
            pqcEnabledCheckbox.addEventListener('change', () => this.handlePQCToggle());
        }

        // Real-time validation
        const inputs = this.form.querySelectorAll('input, select');
        inputs.forEach(input => {
            input.addEventListener('blur', () => this.validateField(input));
            input.addEventListener('input', () => this.clearFieldError(input));
        });
    }

    async loadPresets() {
        try {
            const response = await fetch('/api/gui/presets');
            const result = await response.json();
            
            if (result.success || result.network_presets) {
                this.presets = result.network_presets || result.data?.network_presets;
                this.profiles = result.test_profiles || result.data?.test_profiles;
                this.populatePresetModal();
            }
        } catch (error) {
            console.error('Failed to load presets:', error);
        }
    }

    populatePresetModal() {
        if (!this.presets || !this.profiles) return;

        // Populate network presets
        const networkPresetsContainer = document.getElementById('network-presets');
        if (networkPresetsContainer && this.presets) {
            const html = this.presets.map(preset => `
                <div class="preset-item" data-type="network" data-name="${preset.name}">
                    <h5>${preset.name}</h5>
                    <p>${preset.description}</p>
                    <div class="preset-details">
                        <span>Latency: ${preset.latency}</span>
                        <span>Bandwidth: ${preset.bandwidth}</span>
                        <span>Loss: ${preset.loss}</span>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" onclick="newTestForm.applyNetworkPreset('${preset.name}')">
                        Apply
                    </button>
                </div>
            `).join('');
            networkPresetsContainer.innerHTML = html;
        }

        // Populate test profiles
        const testProfilesContainer = document.getElementById('test-profiles');
        if (testProfilesContainer && this.profiles) {
            const html = this.profiles.map(profile => `
                <div class="preset-item" data-type="profile" data-name="${profile.name}">
                    <h5>${profile.name}</h5>
                    <p>${profile.description}</p>
                    <div class="preset-details">
                        <span>Duration: ${profile.duration}</span>
                        <span>Connections: ${profile.connections}</span>
                        <span>Rate: ${profile.rate} pps</span>
                    </div>
                    <button type="button" class="btn btn-secondary btn-sm" onclick="newTestForm.applyTestProfile('${profile.name}')">
                        Apply
                    </button>
                </div>
            `).join('');
            testProfilesContainer.innerHTML = html;
        }
    }

    applyNetworkPreset(presetName) {
        const preset = this.presets.find(p => p.name === presetName);
        if (!preset) return;

        // Apply network preset values
        const latencyInput = document.getElementById('emulate-latency');
        const lossInput = document.getElementById('emulate-loss');

        if (latencyInput) {
            latencyInput.value = preset.latency;
        }

        if (lossInput) {
            // Convert percentage to decimal (e.g., "1%" -> 0.01)
            const lossValue = parseFloat(preset.loss.replace('%', '')) / 100;
            lossInput.value = lossValue;
        }

        this.hidePresetModal();
        this.showSuccess(`Applied network preset: ${preset.name}`);
    }

    applyTestProfile(profileName) {
        const profile = this.profiles.find(p => p.name === profileName);
        if (!profile) return;

        // Apply test profile values
        const durationInput = document.getElementById('duration');
        const connectionsInput = document.getElementById('connections');
        const streamsInput = document.getElementById('streams');
        const rateInput = document.getElementById('rate');

        if (durationInput) durationInput.value = profile.duration;
        if (connectionsInput) connectionsInput.value = profile.connections;
        if (streamsInput) streamsInput.value = profile.streams;
        if (rateInput) rateInput.value = profile.rate;

        this.hidePresetModal();
        this.showSuccess(`Applied test profile: ${profile.name}`);
    }

    showPresetModal() {
        if (this.presetModal) {
            this.presetModal.style.display = 'flex';
        }
    }

    hidePresetModal() {
        if (this.presetModal) {
            this.presetModal.style.display = 'none';
        }
    }

    handleModeChange() {
        const mode = document.getElementById('mode').value;
        const serverAddrGroup = document.getElementById('server-addr').closest('.form-group');
        
        if (mode === 'server') {
            // Hide server address for server mode
            if (serverAddrGroup) {
                serverAddrGroup.style.display = 'none';
            }
        } else {
            // Show server address for client and test modes
            if (serverAddrGroup) {
                serverAddrGroup.style.display = 'flex';
            }
        }
    }

    handleFECToggle() {
        const fecEnabled = document.getElementById('fec-enabled').checked;
        const fecRedundancyGroup = document.getElementById('fec-redundancy').closest('.form-group');
        
        if (fecRedundancyGroup) {
            fecRedundancyGroup.style.display = fecEnabled ? 'flex' : 'none';
        }
    }

    handlePQCToggle() {
        const pqcEnabled = document.getElementById('pqc-enabled').checked;
        // PQC algorithm selection could be added here if needed
    }

    setupFormValidation() {
        // Add validation rules
        this.validationRules = {
            duration: {
                required: true,
                pattern: /^\d+[smh]$/,
                message: 'Duration must be in format like 60s, 5m, or 1h'
            },
            connections: {
                required: true,
                min: 1,
                max: 100,
                message: 'Connections must be between 1 and 100'
            },
            streams: {
                required: true,
                min: 1,
                max: 100,
                message: 'Streams must be between 1 and 100'
            },
            'packet-size': {
                required: true,
                min: 64,
                max: 65535,
                message: 'Packet size must be between 64 and 65535 bytes'
            },
            rate: {
                required: true,
                min: 1,
                max: 10000,
                message: 'Rate must be between 1 and 10000 packets per second'
            },
            'emulate-loss': {
                min: 0,
                max: 1,
                message: 'Loss rate must be between 0 and 1'
            },
            'emulate-dup': {
                min: 0,
                max: 1,
                message: 'Duplication rate must be between 0 and 1'
            },
            'fec-redundancy': {
                min: 0.05,
                max: 0.20,
                message: 'FEC redundancy must be between 0.05 and 0.20'
            }
        };
    }

    setupDependentFields() {
        // Initialize dependent field states
        this.handleModeChange();
        this.handleFECToggle();
        this.handlePQCToggle();
    }

    validateField(field) {
        const rule = this.validationRules[field.id];
        if (!rule) return true;

        const value = field.value.trim();
        let isValid = true;
        let message = '';

        // Required validation
        if (rule.required && !value) {
            isValid = false;
            message = 'This field is required';
        }

        // Pattern validation
        if (isValid && rule.pattern && value && !rule.pattern.test(value)) {
            isValid = false;
            message = rule.message;
        }

        // Numeric range validation
        if (isValid && (rule.min !== undefined || rule.max !== undefined)) {
            const numValue = parseFloat(value);
            if (isNaN(numValue)) {
                isValid = false;
                message = 'Must be a valid number';
            } else {
                if (rule.min !== undefined && numValue < rule.min) {
                    isValid = false;
                    message = rule.message;
                }
                if (rule.max !== undefined && numValue > rule.max) {
                    isValid = false;
                    message = rule.message;
                }
            }
        }

        // Show/hide error
        if (isValid) {
            this.clearFieldError(field);
        } else {
            this.showFieldError(field, message);
        }

        return isValid;
    }

    showFieldError(field, message) {
        this.clearFieldError(field);
        
        field.classList.add('error');
        
        const errorDiv = document.createElement('div');
        errorDiv.className = 'field-error';
        errorDiv.textContent = message;
        
        field.parentNode.appendChild(errorDiv);
    }

    clearFieldError(field) {
        field.classList.remove('error');
        
        const existingError = field.parentNode.querySelector('.field-error');
        if (existingError) {
            existingError.remove();
        }
    }

    async handleSubmit(e) {
        e.preventDefault();

        // Validate all fields
        const inputs = this.form.querySelectorAll('input[required], select[required]');
        let isValid = true;

        inputs.forEach(input => {
            if (!this.validateField(input)) {
                isValid = false;
            }
        });

        if (!isValid) {
            this.showError('Please fix the validation errors before submitting');
            return;
        }

        // Collect form data
        const formData = new FormData(this.form);
        const config = this.buildTestConfig(formData);

        // Show loading state
        const submitBtn = this.form.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Starting Test...';
        submitBtn.disabled = true;

        try {
            const response = await fetch('/api/tests', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(config)
            });

            if (!response.ok) {
                const errorText = await response.text();
                let errorMessage;
                try {
                    const errorJson = JSON.parse(errorText);
                    errorMessage = errorJson.error || 'Unknown error';
                } catch {
                    errorMessage = errorText || `HTTP ${response.status}`;
                }
                throw new Error(errorMessage);
            }

            const result = await response.json();

            if (result.success && result.data) {
                const testId = result.data.id;
                this.showSuccess('Test started successfully!');
                
                // Redirect to test details page after a short delay
                setTimeout(() => {
                    window.location.href = `/test/${testId}`;
                }, 1500);
            } else {
                throw new Error(result.error || 'Failed to start test');
            }
        } catch (error) {
            console.error('Failed to start test:', error);
            this.showError(`Failed to start test: ${error.message}`);
        } finally {
            // Restore button state
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        }
    }

    buildTestConfig(formData) {
        const config = {};

        // Basic configuration - use JSON field names (lowercase)
        const rawConnections = formData.get('connections');
        const rawStreams = formData.get('streams');
        
        config.mode = formData.get('mode') || 'test';
        config.duration = formData.get('duration') || '60s'; // Send as string, not nanoseconds
        config.connections = parseInt(rawConnections) || 2;
        config.streams = parseInt(rawStreams) || 4;
        config.addr = formData.get('addr') || 'localhost:9000';
        config.packet_size = parseInt(formData.get('packet_size')) || 1200;
        config.rate = parseInt(formData.get('rate')) || 100;

        // Validate required fields
        if (!config.connections || config.connections <= 0 || isNaN(config.connections)) {
            config.connections = 2; // fallback
        }
        if (!config.streams || config.streams <= 0 || isNaN(config.streams)) {
            config.streams = 4; // fallback
        }

        // Optional configuration
        const congestionControl = formData.get('congestion_control');
        if (congestionControl) {
            config.congestion_control = congestionControl;
        }

        // Network emulation
        const emulateLatency = formData.get('emulate_latency');
        if (emulateLatency) {
            config.emulate_latency = emulateLatency; // Send as string
        }

        const emulateLoss = formData.get('emulate_loss');
        if (emulateLoss) {
            config.emulate_loss = parseFloat(emulateLoss);
        }

        const emulateDup = formData.get('emulate_dup');
        if (emulateDup) {
            config.emulate_dup = parseFloat(emulateDup);
        }

        // Advanced options
        config.prometheus = formData.has('prometheus');
        config.fec_enabled = formData.has('fec_enabled');
        
        if (config.fec_enabled) {
            config.fec_redundancy = parseFloat(formData.get('fec_redundancy')) || 0.10;
        }

        config.pqc_enabled = formData.has('pqc_enabled');

        return config;
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showNotification(message, type) {
        // Remove existing notifications
        const existing = document.querySelectorAll('.notification');
        existing.forEach(n => n.remove());

        // Create notification
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;

        // Add to page
        document.body.appendChild(notification);

        // Auto-remove after delay
        setTimeout(() => {
            notification.remove();
        }, type === 'error' ? 5000 : 3000);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.newTestForm = new NewTestForm();
});

// Add CSS for notifications and field errors
const style = document.createElement('style');
style.textContent = `
    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 1rem 1.5rem;
        border-radius: 8px;
        color: white;
        font-weight: 500;
        z-index: 1000;
        animation: slideIn 0.3s ease;
    }

    .notification-success {
        background-color: var(--success-color);
    }

    .notification-error {
        background-color: var(--error-color);
    }

    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    .field-error {
        color: var(--error-color);
        font-size: 0.875rem;
        margin-top: 0.25rem;
    }

    .form-group input.error,
    .form-group select.error {
        border-color: var(--error-color);
        box-shadow: 0 0 0 3px rgb(220 38 38 / 0.1);
    }

    .preset-item {
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius);
        padding: 1rem;
        margin-bottom: 1rem;
    }

    .preset-item h5 {
        margin: 0 0 0.5rem 0;
        color: var(--text-primary);
    }

    .preset-item p {
        margin: 0 0 0.75rem 0;
        color: var(--text-secondary);
        font-size: 0.875rem;
    }

    .preset-details {
        display: flex;
        gap: 1rem;
        margin-bottom: 0.75rem;
        font-size: 0.75rem;
        color: var(--text-secondary);
    }

    .btn-sm {
        padding: 0.5rem 1rem;
        font-size: 0.75rem;
    }
`;
document.head.appendChild(style);