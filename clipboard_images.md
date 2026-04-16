# Clipboard Images in Wails v3

This document describes the new Wails v3 clipboard image support added in this branch.

## Goals

The clipboard implementation now supports two clipboard payload categories across the desktop targets used by Wails v3:

- Plain text
- Images

Rich text, HTML, RTF, and other structured formats are intentionally out of scope for this change.

## What Changed

The clipboard API now supports:

- Setting text
- Reading text
- Setting images from raw image bytes
- Reading images back as PNG bytes
- Reading the clipboard as a typed payload so callers can tell whether the current entry is text, image, or empty

## Clipboard Content Model

A new structured clipboard result is returned by the runtime content API:

```ts
export type ClipboardContentKind = "empty" | "text" | "image";

export interface ClipboardContentResult {
  kind: ClipboardContentKind;
  text?: string;
  data?: string;
  mimeType?: string;
}
```

Rules:

- `kind === "empty"`: the clipboard does not currently expose supported text or image data
- `kind === "text"`: `text` contains the clipboard text and `mimeType` is `text/plain`
- `kind === "image"`: `data` contains base64-encoded PNG bytes and `mimeType` is `image/png`

## Runtime API

The desktop runtime module `@wailsio/runtime` clipboard namespace now exposes these functions:

```ts
Clipboard.SetText(text: string): Promise<void>
Clipboard.Text(): Promise<string>
Clipboard.SetImage(data: string, mimeType?: string): Promise<void>
Clipboard.Image(): Promise<string>
Clipboard.Content(): Promise<ClipboardContentResult>
```

### `Clipboard.SetImage(data, mimeType?)`

- `data` must be a base64 string representing image bytes
- The source bytes may be any image format the platform decoder can read
- The optional `mimeType` is accepted for caller context, but the runtime currently only distinguishes image vs text, not specific rich subtypes

### `Clipboard.Image()`

- Returns the current clipboard image as a base64 string
- The returned bytes are always PNG bytes
- If no image is available, the runtime returns an empty string

### `Clipboard.Content()`

This is the recommended way to inspect the clipboard type before consuming it.

Example:

```ts
import * as Clipboard from "@wailsio/runtime/clipboard";

const content = await Clipboard.Content();

if (content.kind === "text") {
  console.log(content.text);
}

if (content.kind === "image") {
  const pngBytesBase64 = content.data!;
  const dataUrl = `data:${content.mimeType};base64,${pngBytesBase64}`;
  console.log(dataUrl);
}
```

## Go API

The application clipboard manager now supports image-aware operations:

```go
Clipboard.SetText(text string) bool
Clipboard.Text() (string, bool)
Clipboard.SetImage(data []byte) bool
Clipboard.Image() ([]byte, bool)
Clipboard.Content() ClipboardContent
```

`Clipboard.Content()` returns:

```go
type ClipboardContent struct {
    Kind     ClipboardContentKind `json:"kind"`
    Text     string               `json:"text,omitempty"`
    Data     string               `json:"data,omitempty"`
    MimeType string               `json:"mimeType,omitempty"`
}
```

Image content is normalized to PNG when read back.

## Encoding Contract

The transport contract for image data is:

- Input from JS/runtime: base64-encoded image bytes
- Storage in native clipboard: native OS image clipboard format
- Output back to JS/runtime: base64-encoded PNG bytes

This gives you:

- Real image clipboard integration at the OS level
- A simple cross-platform payload shape for your own tools
- A stable output format regardless of how the image was originally supplied

## Type Detection

Clipboard type detection is now performed by checking image support before text support.

That means:

- If the clipboard currently contains an image, `Clipboard.Content()` reports `kind: "image"`
- Otherwise, if text is available, it reports `kind: "text"`
- Otherwise, it reports `kind: "empty"`

This intentionally collapses clipboard state into a simple model for now.

## Platform Implementation Notes

### macOS

The macOS backend now:

- Writes images via `NSPasteboard` using `NSImage`
- Reads image data from the pasteboard and converts it to PNG bytes when necessary
- Keeps text support via `NSPasteboardTypeString`

### Linux GTK3

The GTK3 backend now:

- Writes text with `gtk_clipboard_set_text`
- Reads text with `gtk_clipboard_wait_for_text`
- Writes images with `gtk_clipboard_set_image`
- Reads images with `gtk_clipboard_wait_for_image`
- Converts images to PNG bytes using `gdk_pixbuf_save_to_buffer`

### Linux GTK4

The GTK4 backend now:

- Keeps the existing async-to-sync wrapper approach for clipboard reads
- Reads text with `gdk_clipboard_read_text_async`
- Reads images with `gdk_clipboard_read_texture_async`
- Writes images with `gdk_clipboard_set_texture`
- Normalizes image reads to PNG bytes using `gdk_texture_save_to_png_bytes`

### Windows

The Windows backend now:

- Keeps text support with the existing Unicode clipboard path
- Writes images to the native clipboard as `CF_BITMAP`
- Also stores PNG bytes in the registered `PNG` clipboard format when available
- Prefers reading the registered `PNG` clipboard format first
- Falls back to reading `CF_BITMAP` and converting it to PNG bytes

This means Wails-to-Wails clipboard image transfers can preserve PNG bytes directly, while still interoperating with native Windows applications that consume standard bitmap clipboard data.

## Why PNG on Readback

Clipboard systems expose different native image formats on each OS.

To avoid pushing that complexity into every Wails app, the image read path normalizes everything to PNG bytes. This makes downstream use simpler:

- turn it into a `data:` URL
- write it to disk
- decode it in Go
- pass it to another API

## Limitations

This change does not yet support:

- Rich text detection
- HTML clipboard payloads
- Multiple simultaneous clipboard representations in the public API
- Arbitrary binary clipboard formats
- Mobile clipboard image support on Android or iOS

Android and iOS clipboard implementations still only have placeholders in this branch.

## Recommended Usage Pattern

If you want a single code path in your apps, use:

```ts
const content = await Clipboard.Content();

switch (content.kind) {
case "text":
  // use content.text
  break;
case "image":
  // use content.data as base64 PNG
  break;
default:
  // unsupported or empty clipboard
}
```

If you already know you are writing an image:

```ts
await Clipboard.SetImage(base64ImageBytes, "image/png");
```

## Summary

The new clipboard behavior is intentionally simple:

- Text remains text
- Images become first-class clipboard content
- Image reads come back as base64 PNG bytes
- `Clipboard.Content()` tells you whether the clipboard currently holds text, image, or nothing supported

This change gives Wails v3 applications a clean cross-platform path for clipboard images without introducing rich text handling yet.
