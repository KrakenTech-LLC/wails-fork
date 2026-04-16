//go:build ios

package application

type iosClipboardImpl struct{}

func newClipboardImpl() clipboardImpl {
	return &iosClipboardImpl{}
}

func (c *iosClipboardImpl) setText(text string) bool {
	// iOS clipboard implementation would go here
	return true
}

func (c *iosClipboardImpl) text() (string, bool) {
	// iOS clipboard implementation would go here
	return "", false
}

func (c *iosClipboardImpl) setImage(data []byte) bool {
	return false
}

func (c *iosClipboardImpl) image() ([]byte, bool) {
	return nil, false
}

// SetClipboardText sets the clipboard text on iOS
func (c *ClipboardManager) SetClipboardText(text string) error {
	// iOS clipboard implementation would go here
	// For now, return nil as a placeholder
	return nil
}

// GetClipboardText gets the clipboard text on iOS
func (c *ClipboardManager) GetClipboardText() (string, error) {
	// iOS clipboard implementation would go here
	// For now, return empty string
	return "", nil
}
