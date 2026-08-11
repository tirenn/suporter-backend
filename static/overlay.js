(function () {
    'use strict';

    const overlayContainer = document.getElementById('overlay-container');
    const alertCard = document.getElementById('alert-card');

    let customStyleTag = document.getElementById('custom-template-style');
    if (!customStyleTag) {
        customStyleTag = document.createElement('style');
        customStyleTag.id = 'custom-template-style';
        document.head.appendChild(customStyleTag);
    }

    const defaultTemplateHTML = `<div class="donation-alert-container">
  <div class="donation-header">
    <div class="coin-icon">💰</div>
    <div class="alert-badge">SUPERCHAT DONATION</div>
  </div>
  <div class="donation-amount">{{amount}}</div>
  <div class="donation-donor">{{name}}</div>
  <div class="donation-message-box">
    <p class="donation-message">"{{message}}"</p>
  </div>
</div>`;

    const defaultTemplateCSS = `.donation-alert-container {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.95), rgba(5, 150, 105, 0.95));
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  padding: 28px;
  max-width: 450px;
  color: #ffffff;
  font-family: 'Outfit', sans-serif;
  text-align: center;
  box-shadow: 0 20px 50px rgba(16, 185, 129, 0.5), inset 0 0 15px rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(16px);
  animation: popIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.donation-header { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 12px; }
.coin-icon { font-size: 1.5rem; animation: bounce 1s infinite alternate; }
.alert-badge { background: rgba(0, 0, 0, 0.35); padding: 4px 14px; border-radius: 12px; font-size: 0.75rem; font-weight: 800; letter-spacing: 0.1em; color: #a7f3d0; text-transform: uppercase; }
.donation-amount { font-size: 2.8rem; font-weight: 800; color: #fef08a; text-shadow: 0 4px 12px rgba(0, 0, 0, 0.4); letter-spacing: -0.02em; margin-bottom: 4px; }
.donation-donor { font-size: 1.4rem; font-weight: 700; color: #ffffff; margin-bottom: 14px; }
.donation-message-box { background: rgba(0, 0, 0, 0.25); border: 1px dashed rgba(255, 255, 255, 0.25); border-radius: 12px; padding: 12px 16px; }
.donation-message { font-size: 0.98rem; font-style: italic; line-height: 1.5; color: #ecfdf5; }
@keyframes popIn { 0% { transform: scale(0.6) translateY(30px); opacity: 0; } 100% { transform: scale(1) translateY(0); opacity: 1; } }
@keyframes bounce { from { transform: translateY(0); } to { transform: translateY(-6px); } }`;

    const alertQueue = [];
    let isProcessing = false;
    let audioCtx = null;

    // Apply URL Query Parameter Alignment (e.g. ?align=top-right or ?align=center)
    function applyAlignmentFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        const alignParam = urlParams.get('align');

        if (alignParam) {
            const validAlignments = [
                'top-left', 'top-center', 'top-right',
                'center-left', 'center', 'center-right',
                'bottom-left', 'bottom-center', 'bottom-right'
            ];
            const cleanAlign = alignParam.toLowerCase().trim();
            if (validAlignments.includes(cleanAlign)) {
                overlayContainer.className = `align-${cleanAlign}`;
                console.log('[Overlay] Applied alignment:', cleanAlign);
            }
        }
    }

    // Synthesize alert chime via Web Audio API
    function playNotificationSound(soundType) {
        if (soundType === 'silent') return;

        try {
            if (!audioCtx) {
                const AudioContext = window.AudioContext || window.webkitAudioContext;
                if (!AudioContext) return;
                audioCtx = new AudioContext();
            }

            if (audioCtx.state === 'suspended') {
                audioCtx.resume();
            }

            const now = audioCtx.currentTime;

            const notes = [523.25, 659.25, 783.99, 1046.50];
            notes.forEach((freq, idx) => {
                const osc = audioCtx.createOscillator();
                const gain = audioCtx.createGain();
                osc.type = 'sine';
                osc.frequency.setValueAtTime(freq, now + idx * 0.08);

                gain.gain.setValueAtTime(0.2, now + idx * 0.08);
                gain.gain.exponentialRampToValueAtTime(0.001, now + idx * 0.08 + 0.4);

                osc.connect(gain);
                gain.connect(audioCtx.destination);

                osc.start(now + idx * 0.08);
                osc.stop(now + idx * 0.08 + 0.45);
            });
        } catch (e) {
            console.warn('[Audio] Playback error:', e);
        }
    }

    function renderTemplateHTML(templateStr, dataMap) {
        if (!templateStr) return '';
        let rendered = templateStr;

        const amountVal = dataMap.amount || dataMap.Amount || dataMap.value || '';
        const nameVal = dataMap.name || dataMap.Name || dataMap.donor || 'Anonymous';
        const msgVal = dataMap.message || dataMap.Message || dataMap.description || dataMap.Description || '';

        const fullMap = {
            ...dataMap,
            name: nameVal,
            Name: nameVal,
            amount: amountVal,
            Amount: amountVal,
            message: msgVal,
            Message: msgVal,
            description: msgVal,
            Description: msgVal
        };

        for (const [key, val] of Object.entries(fullMap)) {
            if (!key) continue;
            const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'gi');
            rendered = rendered.replace(regex, val !== undefined && val !== null ? val : '');
        }
        return rendered;
    }

    function processQueue() {
        if (alertQueue.length === 0) {
            isProcessing = false;
            return;
        }

        isProcessing = true;
        const currentAlert = alertQueue.shift();

        let extraPayload = {};
        if (currentAlert.payload) {
            if (typeof currentAlert.payload === 'string') {
                try {
                    extraPayload = JSON.parse(currentAlert.payload);
                } catch (e) {
                    extraPayload = {};
                }
            } else if (typeof currentAlert.payload === 'object') {
                extraPayload = currentAlert.payload;
            }
        }

        const dataMap = {
            name: currentAlert.name || extraPayload.name || 'Anonymous',
            amount: currentAlert.amount || extraPayload.amount || '',
            message: currentAlert.message || extraPayload.message || extraPayload.description || '',
            description: currentAlert.message || extraPayload.description || extraPayload.message || '',
            ...extraPayload
        };

        console.log('[Overlay] Processing alert with dataMap:', dataMap);

        // Fallback handler: If template HTML is empty or missing, use default donation overlay template!
        let targetHTML = (currentAlert.html_template && currentAlert.html_template.trim() !== '') ? currentAlert.html_template : defaultTemplateHTML;
        let targetCSS = (currentAlert.css_style && currentAlert.css_style.trim() !== '') ? currentAlert.css_style : defaultTemplateCSS;

        const compiledHTML = renderTemplateHTML(targetHTML, dataMap);
        customStyleTag.textContent = targetCSS;

        alertCard.innerHTML = compiledHTML + `<div class="progress-bar" id="progress-bar"></div>`;
        alertCard.className = 'alert-card hidden custom-template-active';

        const currentProgressBar = document.getElementById('progress-bar');
        if (currentProgressBar) {
            currentProgressBar.classList.remove('animating');
            currentProgressBar.style.animationDuration = '0ms';
            void currentProgressBar.offsetWidth;
        }

        playNotificationSound('donation');

        requestAnimationFrame(() => {
            alertCard.classList.remove('hidden');
            alertCard.classList.add('active');

            const duration = currentAlert.duration || 7000;
            if (currentProgressBar) {
                currentProgressBar.style.animationDuration = `${duration}ms`;
                currentProgressBar.classList.add('animating');
            }

            setTimeout(() => {
                alertCard.classList.remove('active');
                alertCard.classList.add('fade-out');

                setTimeout(() => {
                    alertCard.classList.remove('fade-out');
                    alertCard.classList.add('hidden');
                    setTimeout(processQueue, 300);
                }, 400);
            }, duration);
        });
    }

    function enqueueAlert(alertData) {
        alertQueue.push(alertData);
        if (!isProcessing) {
            processQueue();
        }
    }

    function connectProjectSSE() {
        applyAlignmentFromURL();

        const pathParts = window.location.pathname.split('/overlay/');
        if (pathParts.length < 2) {
            console.error('[Overlay] Invalid overlay path. Project ID missing.');
            return;
        }

        const projectID = pathParts[1].split('/')[0];
        const streamUrl = window.location.origin + '/overlay/' + projectID + '/stream';
        console.log('[Overlay] Connecting stream for project ID:', projectID, streamUrl);

        const eventSource = new EventSource(streamUrl);

        eventSource.onopen = function () {
            console.log('[Overlay] EventSource connected for project:', projectID);
        };

        eventSource.addEventListener('alert', function (e) {
            try {
                const alertData = JSON.parse(e.data);
                console.log('[Overlay] Alert received:', alertData);
                enqueueAlert(alertData);
            } catch (err) {
                console.error('[Overlay] Error parsing alert:', err);
            }
        });

        eventSource.onerror = function (err) {
            console.warn('[Overlay] Disconnected. Reconnecting in 3s...', err);
            eventSource.close();
            setTimeout(connectProjectSSE, 3000);
        };
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', connectProjectSSE);
    } else {
        connectProjectSSE();
    }
})();
