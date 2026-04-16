//go:build linux && !android && !server

package application

import (
	"sync"
)

var clipboardLock sync.RWMutex

type linuxClipboard struct{}

func (m linuxClipboard) setText(text string) bool {
	clipboardLock.Lock()
	defer clipboardLock.Unlock()
	return clipboardSetText(text)
}

func (m linuxClipboard) text() (string, bool) {
	clipboardLock.RLock()
	defer clipboardLock.RUnlock()
	return clipboardGetText()
}

func (m linuxClipboard) setImage(data []byte) bool {
	clipboardLock.Lock()
	defer clipboardLock.Unlock()
	return clipboardSetImage(data)
}

func (m linuxClipboard) image() ([]byte, bool) {
	clipboardLock.RLock()
	defer clipboardLock.RUnlock()
	return clipboardGetImage()
}

func newClipboardImpl() *linuxClipboard {
	return &linuxClipboard{}
}
