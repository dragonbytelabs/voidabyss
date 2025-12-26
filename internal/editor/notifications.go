package editor

import (
	"time"

	"github.com/dragonbytelabs/voidabyss/internal/config"
)

const notificationDisplayDuration = 3 * time.Second

// processNotifications pulls notifications from Lua and updates statusMsg
func (e *Editor) processNotifications() {
	if e.loader == nil || e.loader.Notifications == nil {
		return
	}

	// Clear old notification if timeout expired
	if !e.notifTimeout.IsZero() && time.Now().After(e.notifTimeout) {
		// Don't clear if we're in command mode or showing other messages
		if e.mode != ModeCommand && e.statusMsg != "" {
			// Only clear if it was a notification (check if it starts with level indicator)
			if len(e.statusMsg) > 0 && (e.statusMsg[0] == '[' || e.notifLevel != config.NotifyInfo) {
				e.statusMsg = ""
			}
		}
		e.notifTimeout = time.Time{}
	}

	// Pull latest notification if available
	notif := e.loader.Notifications.Pop()
	if notif != nil {
		e.statusMsg = formatNotification(notif)
		e.notifLevel = notif.Level
		e.notifTimeout = time.Now().Add(notificationDisplayDuration)
	}
}

// formatNotification formats a notification for display
func formatNotification(notif *config.Notification) string {
	switch notif.Level {
	case config.NotifyWarn:
		return "[WARN] " + notif.Message
	case config.NotifyError:
		return "[ERROR] " + notif.Message
	default:
		return notif.Message
	}
}

// GetNotificationLevel returns the current notification level (for future styling)
func (e *Editor) GetNotificationLevel() config.NotificationLevel {
	return e.notifLevel
}
