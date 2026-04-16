package application

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	clipboardImageMimeType = "image/png"
	clipboardTextMimeType  = "text/plain"
)

func decodeClipboardImage(data []byte) (image.Image, error) {
	imageData, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return imageData, nil
}

func encodeClipboardImage(imageData image.Image) ([]byte, error) {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imageData); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func clipboardImageToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func clipboardImageFromBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
