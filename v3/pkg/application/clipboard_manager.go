package application

// ClipboardManager manages clipboard operations
type ClipboardManager struct {
	app       *App
	clipboard *Clipboard
}

// newClipboardManager creates a new ClipboardManager instance
func newClipboardManager(app *App) *ClipboardManager {
	return &ClipboardManager{
		app: app,
	}
}

// SetText sets text in the clipboard
func (cm *ClipboardManager) SetText(text string) bool {
	return cm.getClipboard().SetText(text)
}

// Text gets text from the clipboard
func (cm *ClipboardManager) Text() (string, bool) {
	return cm.getClipboard().Text()
}

// SetImage sets image bytes in the clipboard.
func (cm *ClipboardManager) SetImage(data []byte) bool {
	return cm.getClipboard().SetImage(data)
}

// Image gets clipboard image bytes as PNG data.
func (cm *ClipboardManager) Image() ([]byte, bool) {
	return cm.getClipboard().Image()
}

// Content gets the current clipboard contents as either text or image.
func (cm *ClipboardManager) Content() ClipboardContent {
	return cm.getClipboard().Content()
}

// getClipboard returns the clipboard instance, creating it if needed (lazy initialization)
func (cm *ClipboardManager) getClipboard() *Clipboard {
	if cm.clipboard == nil {
		cm.clipboard = newClipboard()
	}
	return cm.clipboard
}
