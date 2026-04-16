package application

type clipboardImpl interface {
	setText(text string) bool
	text() (string, bool)
	setImage(data []byte) bool
	image() ([]byte, bool)
}

type ClipboardContentKind string

const (
	ClipboardContentKindEmpty ClipboardContentKind = "empty"
	ClipboardContentKindText  ClipboardContentKind = "text"
	ClipboardContentKindImage ClipboardContentKind = "image"
)

type ClipboardContent struct {
	Kind     ClipboardContentKind `json:"kind"`
	Text     string               `json:"text,omitempty"`
	Data     string               `json:"data,omitempty"`
	MimeType string               `json:"mimeType,omitempty"`
}

type Clipboard struct {
	impl clipboardImpl
}

func newClipboard() *Clipboard {
	return &Clipboard{
		impl: newClipboardImpl(),
	}
}

func (c *Clipboard) SetText(text string) bool {
	return InvokeSyncWithResult(func() bool {
		return c.impl.setText(text)
	})
}

func (c *Clipboard) Text() (string, bool) {
	return InvokeSyncWithResultAndOther(c.impl.text)
}

func (c *Clipboard) SetImage(data []byte) bool {
	return InvokeSyncWithResult(func() bool {
		return c.impl.setImage(data)
	})
}

func (c *Clipboard) Image() ([]byte, bool) {
	return InvokeSyncWithResultAndOther(c.impl.image)
}

func (c *Clipboard) Content() ClipboardContent {
	if imageData, ok := c.Image(); ok {
		return ClipboardContent{
			Kind:     ClipboardContentKindImage,
			Data:     clipboardImageToBase64(imageData),
			MimeType: clipboardImageMimeType,
		}
	}

	if text, ok := c.Text(); ok {
		return ClipboardContent{
			Kind:     ClipboardContentKindText,
			Text:     text,
			MimeType: clipboardTextMimeType,
		}
	}

	return ClipboardContent{
		Kind: ClipboardContentKindEmpty,
	}
}
