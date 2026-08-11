package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"suporter-backend/internal/domain"
)

type SSEBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan domain.Alert]bool
}

func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		subscribers: make(map[string]map[chan domain.Alert]bool),
	}
}

func (b *SSEBroker) Subscribe(projectID string) chan domain.Alert {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[projectID]; !exists {
		b.subscribers[projectID] = make(map[chan domain.Alert]bool)
	}

	ch := make(chan domain.Alert, 20)
	b.subscribers[projectID][ch] = true
	log.Printf("[SSE Hub] Client connected to project '%s'. Active listeners: %d", projectID, len(b.subscribers[projectID]))

	return ch
}

func (b *SSEBroker) Unsubscribe(projectID string, ch chan domain.Alert) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if clients, exists := b.subscribers[projectID]; exists {
		if _, ok := clients[ch]; ok {
			delete(clients, ch)
			close(ch)
			log.Printf("[SSE Hub] Client disconnected from project '%s'. Remaining listeners: %d", projectID, len(clients))
		}
		if len(clients) == 0 {
			delete(b.subscribers, projectID)
		}
	}
}

func (b *SSEBroker) Broadcast(projectID string, alert domain.Alert) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	clients, exists := b.subscribers[projectID]
	if !exists || len(clients) == 0 {
		log.Printf("[SSE Hub] No active subscribers for project '%s'. Alert buffered.", projectID)
		return
	}

	log.Printf("[SSE Hub] Broadcasting alert for project '%s' to %d listener(s)", projectID, len(clients))

	for ch := range clients {
		select {
		case ch <- alert:
		default:
			log.Printf("[SSE Hub] Client channel full for project '%s', skipping payload", projectID)
		}
	}
}

func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request, projectID string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	messageChan := b.Subscribe(projectID)
	defer b.Unsubscribe(projectID, messageChan)

	// Send initial ping event
	fmt.Fprintf(w, "event: ping\ndata: {\"project_id\":\"%s\",\"status\":\"connected\"}\n\n", projectID)
	flusher.Flush()

	notify := r.Context().Done()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: {\"timestamp\":%d}\n\n", time.Now().Unix())
			flusher.Flush()
		case alert, ok := <-messageChan:
			if !ok {
				return
			}
			jsonData, err := json.Marshal(alert)
			if err != nil {
				log.Printf("[SSE Hub] Error marshaling alert: %v", err)
				continue
			}
			fmt.Fprintf(w, "event: alert\ndata: %s\n\n", jsonData)
			flusher.Flush()
		}
	}
}
