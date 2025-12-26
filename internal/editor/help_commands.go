package editor

import "strings"

// showHelp opens a help buffer with the requested topic
func (e *Editor) showHelp(topic string) {
	content, ok := GetHelp(topic)
	if !ok {
		// Show available topics
		topics := ListHelpTopics()
		e.statusMsg = "No help for '" + topic + "'. Try: " + strings.Join(topics, ", ")
		return
	}

	// Split content into lines for popup
	lines := strings.Split(content, "\n")

	// Create title
	title := "HELP"
	if topic != "" && topic != "main" {
		title = "HELP: " + strings.ToUpper(topic)
	}

	// Open popup with help content
	e.popupFixedH = 0 // Use dynamic height
	e.openPopup(title, lines)
}
