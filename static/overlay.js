(function () {
    'use strict';

    const overlayContainer = document.getElementById('overlay-container');
    const alertCard = document.getElementById('alert-card');
    const alertName = document.getElementById('alert-name');
    const alertMessage = document.getElementById('alert-message');
    const badgeText = document.getElementById('badge-text');
    const badgeIcon = document.getElementById('badge-icon');
    const progressBar = document.getElementById('progress-bar');

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

            if (soundType === 'donation' || soundType === 'gold') {
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
            } else {
                const osc1 = audioCtx.createOscillator();
                const osc2 = audioCtx.createOscillator();
                const gain1 = audioCtx.createGain();
                const gain2 = audioCtx.createGain();

                osc1.type = 'sine';
                osc2.type = 'triangle';

                osc1.frequency.setValueAtTime(587.33, now);
                osc1.frequency.exponentialRampToValueAtTime(880, now + 0.1);

                osc2.frequency.setValueAtTime(880, now + 0.1);
                osc2.frequency.exponentialRampToValueAtTime(1174.66, now + 0.2);

                gain1.gain.setValueAtTime(0.25, now);
                gain1.gain.exponentialRampToValueAtTime(0.001, now + 0.35);

                gain2.gain.setValueAtTime(0.2, now + 0.1);
                gain2.gain.exponentialRampToValueAtTime(0.001, now + 0.5);

                osc1.connect(gain1);
                gain1.connect(audioCtx.destination);
                osc2.connect(gain2);
                gain2.connect(audioCtx.destination);

                osc1.start(now);
                osc1.stop(now + 0.35);
                osc2.start(now + 0.1);
                osc2.stop(now + 0.5);
            }
        } catch (e) {
            console.warn('[Audio] Playback error:', e);
        }
    }

    const icons = {
        default: `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>`,
        donation: `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="1" x2="12" y2="23"></line><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path></svg>`,
        sub: `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>`,
        follow: `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l8.72-8.72 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>`
    };

    function processQueue() {
        if (alertQueue.length === 0) {
            isProcessing = false;
            return;
        }

        isProcessing = true;
        const currentAlert = alertQueue.shift();

        alertName.textContent = currentAlert.name || 'Anonymous';
        alertMessage.textContent = currentAlert.message || '';

        const type = (currentAlert.type || 'default').toLowerCase();
        badgeText.textContent = type === 'donation' ? 'DONATION ALERT' :
                                type === 'sub' ? 'NEW SUBSCRIBER' :
                                type === 'follow' ? 'NEW FOLLOWER' : 'NEW MESSAGE';

        badgeIcon.innerHTML = icons[type] || icons.default;

        alertCard.className = 'alert-card hidden';
        if (type !== 'default') {
            alertCard.classList.add(`theme-${type}`);
        }

        progressBar.classList.remove('animating');
        progressBar.style.animationDuration = '0ms';
        void progressBar.offsetWidth;

        playNotificationSound(currentAlert.sound || type);

        requestAnimationFrame(() => {
            alertCard.classList.remove('hidden');
            alertCard.classList.add('active');

            const duration = currentAlert.duration || 5000;
            progressBar.style.animationDuration = `${duration}ms`;
            progressBar.classList.add('animating');

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
