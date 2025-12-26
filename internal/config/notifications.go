package config

import (
	"sync"
	"time"
)

// NotificationLevel represents the severity of a notification
type NotificationLevel int

const (
	NotifyInfo NotificationLevel = iota
	NotifyWarn
	NotifyError
)

// Notification represents a single notification message
type Notification struct {
	Message   string
	Level     NotificationLevel
	Timestamp time.Time
}

// NotificationQueue is a thread-safe queue for notifications
type NotificationQueue struct {
	mu            sync.Mutex
	notifications []Notification
	lastMessage   *Notification
}

// NewNotificationQueue creates a new notification queue
func NewNotificationQueue() *NotificationQueue {
	return &NotificationQueue{
		notifications: make([]Notification, 0),
	}
}

// Push adds a notification to the queue
func (nq *NotificationQueue) Push(msg string, level NotificationLevel) {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	notif := Notification{
		Message:   msg,
		Level:     level,
		Timestamp: time.Now(),
	}

	nq.notifications = append(nq.notifications, notif)
	nq.lastMessage = &notif
}

// Pop returns and removes the oldest notification, or nil if empty
func (nq *NotificationQueue) Pop() *Notification {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	if len(nq.notifications) == 0 {
		return nil
	}

	notif := nq.notifications[0]
	nq.notifications = nq.notifications[1:]
	return &notif
}

// Peek returns the oldest notification without removing it
func (nq *NotificationQueue) Peek() *Notification {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	if len(nq.notifications) == 0 {
		return nil
	}

	return &nq.notifications[0]
}

// Last returns the most recent notification
func (nq *NotificationQueue) Last() *Notification {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	return nq.lastMessage
}

// Len returns the number of pending notifications
func (nq *NotificationQueue) Len() int {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	return len(nq.notifications)
}

// Clear removes all notifications
func (nq *NotificationQueue) Clear() {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	nq.notifications = make([]Notification, 0)
	nq.lastMessage = nil
}

// ParseLevel converts a string to NotificationLevel
func ParseLevel(level string) NotificationLevel {
	switch level {
	case "warn", "warning":
		return NotifyWarn
	case "error", "err":
		return NotifyError
	default:
		return NotifyInfo
	}
}

// String returns the string representation of the level
func (l NotificationLevel) String() string {
	switch l {
	case NotifyWarn:
		return "warn"
	case NotifyError:
		return "error"
	default:
		return "info"
	}
}
