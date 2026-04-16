package application

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/errs"
)

const (
	ClipboardSetText    = 0
	ClipboardText       = 1
	ClipboardSetImage   = 2
	ClipboardImage      = 3
	ClipboardGetContent = 4
)

var clipboardMethods = map[int]string{
	ClipboardSetText:    "SetText",
	ClipboardText:       "Text",
	ClipboardSetImage:   "SetImage",
	ClipboardImage:      "Image",
	ClipboardGetContent: "Content",
}

func (m *MessageProcessor) processClipboardMethod(req *RuntimeRequest) (any, error) {
	args := req.Args.AsMap()

	var text string

	switch req.Method {
	case ClipboardSetText:
		textp := args.String("text")
		if textp == nil {
			return nil, errs.NewInvalidClipboardCallErrorf("missing argument 'text'")
		}
		text = *textp
		globalApplication.Clipboard.SetText(text)
		return unit, nil
	case ClipboardText:
		text, _ = globalApplication.Clipboard.Text()
		return text, nil
	case ClipboardSetImage:
		data := args.String("data")
		if data == nil {
			return nil, errs.NewInvalidClipboardCallErrorf("missing argument 'data'")
		}
		if mimeType := args.String("mimeType"); mimeType != nil && *mimeType != "" && !strings.HasPrefix(*mimeType, "image/") {
			return nil, errs.NewInvalidClipboardCallErrorf("unsupported mime type '%s'", *mimeType)
		}
		imageData, err := clipboardImageFromBase64(*data)
		if err != nil {
			return nil, errs.NewInvalidClipboardCallErrorf("invalid base64 image data: %v", err)
		}
		if !globalApplication.Clipboard.SetImage(imageData) {
			return nil, errs.NewInvalidClipboardCallErrorf("failed to set clipboard image")
		}
		return unit, nil
	case ClipboardImage:
		imageData, ok := globalApplication.Clipboard.Image()
		if !ok {
			return "", nil
		}
		return clipboardImageToBase64(imageData), nil
	case ClipboardGetContent:
		return globalApplication.Clipboard.Content(), nil
	default:
		return nil, errs.NewInvalidClipboardCallErrorf("unknown method: %d", req.Method)
	}
}
