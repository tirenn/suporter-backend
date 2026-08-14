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

    const defaultTemplateHTML = `<div class="cartoon-alert-container">
  <div class="cartoon-header">
    <div class="cartoon-sparkle">💥</div>
    <div class="cartoon-badge">Suporter datang!!!</div>
    <div class="cartoon-sparkle">⚡</div>
  </div>
  <div class="cartoon-hero">
    <span class="cartoon-name">{{name}}</span>
    <span class="cartoon-action">mengirimkan</span>
    <span class="cartoon-amount">Rp {{amount}}</span>
  </div>
  <div class="cartoon-message-bubble">
    <p class="cartoon-message">{{message}}</p>
  </div>
</div>`;

    const defaultTemplateCSS = `@import url('https://fonts.googleapis.com/css2?family=Fredoka:wght@600;700;800&family=Nunito:wght@700;800;900&display=swap');

.cartoon-alert-container {
  background: linear-gradient(135deg, #FFF066 0%, #FFB800 50%, #FF8A00 100%);
  border: 4px solid #1E293B;
  border-radius: 24px;
  padding: 24px 28px;
  max-width: 480px;
  box-shadow: 6px 8px 0px #0F172A, 0 20px 40px rgba(0, 0, 0, 0.25);
  font-family: 'Fredoka', 'Nunito', sans-serif;
  text-align: center;
  position: relative;
  overflow: hidden;
  animation: cartoonPopIn 0.55s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.cartoon-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.cartoon-sparkle {
  font-size: 1.4rem;
  animation: cartoonBounce 0.8s infinite alternate ease-in-out;
}

.cartoon-badge {
  background: #FF4757;
  color: #FFFFFF;
  border: 3px solid #1E293B;
  border-radius: 50px;
  padding: 4px 18px;
  font-size: 1rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  box-shadow: 3px 3px 0px #1E293B;
  transform: rotate(-1deg);
}

.cartoon-hero {
  font-size: 1.35rem;
  font-weight: 800;
  color: #1E293B;
  line-height: 1.35;
  margin-bottom: 14px;
  text-shadow: 1px 1px 0px rgba(255, 255, 255, 0.6);
}

.cartoon-name {
  color: #2E5BFF;
  font-weight: 900;
  text-decoration: underline wavy #FF4757;
  padding: 0 4px;
}

.cartoon-action {
  color: #1E293B;
  font-weight: 700;
  margin: 0 4px;
}

.cartoon-amount {
  color: #059669;
  background: #FFFFFF;
  border: 2.5px solid #1E293B;
  border-radius: 12px;
  padding: 2px 10px;
  font-weight: 900;
  display: inline-block;
  box-shadow: 2px 3px 0px #1E293B;
  margin-left: 4px;
}

.cartoon-message-bubble {
  background: #FFFFFF;
  border: 3.5px solid #1E293B;
  border-radius: 18px;
  padding: 12px 18px;
  box-shadow: 4px 4px 0px #1E293B;
  position: relative;
  margin-top: 6px;
}

.cartoon-message {
  font-family: 'Nunito', sans-serif;
  font-size: 1.05rem;
  font-weight: 800;
  color: #1E293B;
  line-height: 1.4;
  margin: 0;
  word-break: break-word;
}

@keyframes cartoonPopIn {
  0% { transform: scale(0.4) rotate(-8deg); opacity: 0; }
  70% { transform: scale(1.06) rotate(2deg); opacity: 1; }
  100% { transform: scale(1) rotate(0deg); opacity: 1; }
}

@keyframes cartoonBounce {
  from { transform: translateY(0) scale(1); }
  to { transform: translateY(-5px) scale(1.15); }
}`;

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

        let amountVal = dataMap.amount || dataMap.Amount || dataMap.value || '';
        // If amount starts with "Rp" or "Rp.", strip it so template's "Rp {{amount}}" or "Rp {amount}" renders cleanly
        if (typeof amountVal === 'string') {
            amountVal = amountVal.replace(/^Rp\.?\s*/i, '').trim();
        }

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
            const cleanVal = val !== undefined && val !== null ? val : '';
            // Match {{ key }} or { key }
            const doubleBracketRegex = new RegExp(`{{\\s*${key}\\s*}}`, 'gi');
            const singleBracketRegex = new RegExp(`(?<!{){\\s*${key}\\s*}(?!})`, 'gi');
            rendered = rendered.replace(doubleBracketRegex, cleanVal).replace(singleBracketRegex, cleanVal);
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
