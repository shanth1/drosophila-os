package telemetry

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

const subscriberCapacity = 64

// Hub stores serialized state only; HTTP readers never touch engine arrays.
type Hub struct {
	mu          sync.Mutex
	runID       string
	sequence    uint64
	timestamp   time.Time
	state       json.RawMessage
	subscribers map[*Subscription]struct{}
	closed      bool
}

type Subscription struct {
	Snapshot []byte
	Updates  <-chan []byte
	Done     <-chan struct{}
	updates  chan []byte
	done     chan struct{}
}

func New(initial State) (*Hub, error) {
	data, err := encodeState(initial)
	if err != nil {
		return nil, err
	}
	return &Hub{runID: rand.Text(), timestamp: time.Now().UTC(), state: data, subscribers: make(map[*Subscription]struct{})}, nil
}

func encodeState(state State) ([]byte, error) {
	if state.Modules == nil {
		state.Modules = []Module{}
	}
	if state.Entities == nil {
		state.Entities = []Entity{}
	}
	if state.Observations == nil {
		state.Observations = []Observation{}
	}
	return json.Marshal(state)
}

// Publish is called by the runtime owner after a completed tick. The entire
// state/event pair and subscription handoff share one sequence boundary.
func (h *Hub) Publish(state State, event *Event) error {
	data, err := encodeState(state)
	if err != nil {
		return err
	}
	var eventData []byte
	if event != nil {
		eventData, err = json.Marshal(event)
		if err != nil {
			return err
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return errors.New("telemetry hub is closed")
	}
	h.state, h.timestamp = data, time.Now().UTC()
	h.sequence++
	message, err := h.messageLocked("state", data)
	if err != nil {
		return err
	}
	h.broadcastLocked(message)
	if event != nil {
		h.sequence++
		message, err := h.messageLocked("event", eventData)
		if err != nil {
			return err
		}
		h.broadcastLocked(message)
	}
	return nil
}

func (h *Hub) messageLocked(kind string, payload json.RawMessage) ([]byte, error) {
	return json.Marshal(Message{APIVersion: APIVersion, RunID: h.runID, Sequence: h.sequence, Timestamp: h.timestamp, Type: kind, Payload: payload})
}

func (h *Hub) Snapshot() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.messageLocked("snapshot", h.state)
}

// Subscribe captures the snapshot and registers updates atomically. Returned
// message bytes are immutable. Slow subscribers are disconnected, not skipped.
func (h *Hub) Subscribe() (*Subscription, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil, errors.New("telemetry hub is closed")
	}
	snapshot, err := h.messageLocked("snapshot", h.state)
	if err != nil {
		return nil, err
	}
	updates, done := make(chan []byte, subscriberCapacity), make(chan struct{})
	subscriber := &Subscription{Snapshot: snapshot, Updates: updates, Done: done, updates: updates, done: done}
	h.subscribers[subscriber] = struct{}{}
	return subscriber, nil
}

func (h *Hub) Unsubscribe(subscriber *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeLocked(subscriber)
}

func (h *Hub) removeLocked(subscriber *Subscription) {
	if _, exists := h.subscribers[subscriber]; exists {
		delete(h.subscribers, subscriber)
		close(subscriber.done)
	}
}

func (h *Hub) broadcastLocked(message []byte) {
	for subscriber := range h.subscribers {
		select {
		case subscriber.updates <- message:
		default:
			h.removeLocked(subscriber)
		}
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for subscriber := range h.subscribers {
		h.removeLocked(subscriber)
	}
}
