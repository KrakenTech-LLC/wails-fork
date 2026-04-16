//go:build darwin && !ios && !server

package application

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa -mmacosx-version-min=10.13

#import <Cocoa/Cocoa.h>
#import <stdlib.h>
#import <string.h>

bool setClipboardText(const char* text) {
	NSPasteboard *pasteBoard = [NSPasteboard generalPasteboard];
	NSString *string = [NSString stringWithUTF8String:text];
	[pasteBoard clearContents];
	return [pasteBoard setString:string forType:NSPasteboardTypeString];
}

const char* getClipboardText() {
	NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
	NSString *text = [pasteboard stringForType:NSPasteboardTypeString];
	if (text == nil) {
		return NULL;
	}
	return [text UTF8String];
}

bool setClipboardImage(const void *data, NSUInteger length) {
	NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
	NSData *imageData = [NSData dataWithBytes:data length:length];
	NSImage *image = [[NSImage alloc] initWithData:imageData];
	if (image == nil) {
		return false;
	}
	[pasteboard clearContents];
	BOOL success = [pasteboard writeObjects:@[image]];
	return success;
}

void* getClipboardImage(NSUInteger *length) {
	NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
	NSData *imageData = [pasteboard dataForType:NSPasteboardTypePNG];
	if (imageData == nil) {
		NSArray *classes = @[[NSImage class]];
		if (![pasteboard canReadObjectForClasses:classes options:@{}]) {
			return NULL;
		}
		NSArray *objects = [pasteboard readObjectsForClasses:classes options:@{}];
		if ([objects count] == 0) {
			return NULL;
		}
		NSImage *image = objects[0];
		NSBitmapImageRep *bitmap = nil;
		for (NSImageRep *representation in [image representations]) {
			if ([representation isKindOfClass:[NSBitmapImageRep class]]) {
				bitmap = (NSBitmapImageRep *)representation;
				break;
			}
		}
		if (bitmap == nil) {
			NSData *tiffData = [image TIFFRepresentation];
			if (tiffData == nil) {
				return NULL;
			}
			bitmap = [NSBitmapImageRep imageRepWithData:tiffData];
		}
		if (bitmap == nil) {
			return NULL;
		}
		imageData = [bitmap representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
	}
	if (imageData == nil) {
		return NULL;
	}
	*length = [imageData length];
	void *buffer = malloc(*length);
	memcpy(buffer, [imageData bytes], *length);
	return buffer;
}

void freeClipboardData(void *data) {
	if (data != NULL) {
		free(data);
	}
}

*/
import "C"
import (
	"sync"
	"unsafe"
)

var clipboardLock sync.RWMutex

type macosClipboard struct{}

func (m macosClipboard) setText(text string) bool {
	clipboardLock.Lock()
	defer clipboardLock.Unlock()
	cText := C.CString(text)
	success := C.setClipboardText(cText)
	C.free(unsafe.Pointer(cText))
	return bool(success)
}

func (m macosClipboard) text() (string, bool) {
	clipboardLock.RLock()
	defer clipboardLock.RUnlock()
	clipboardText := C.getClipboardText()
	if clipboardText == nil {
		return "", false
	}
	result := C.GoString(clipboardText)
	return result, true
}

func (m macosClipboard) setImage(data []byte) bool {
	clipboardLock.Lock()
	defer clipboardLock.Unlock()
	if len(data) == 0 {
		return false
	}
	return bool(C.setClipboardImage(unsafe.Pointer(&data[0]), C.ulong(len(data))))
}

func (m macosClipboard) image() ([]byte, bool) {
	clipboardLock.RLock()
	defer clipboardLock.RUnlock()
	var length C.ulong
	imageData := C.getClipboardImage(&length)
	if imageData == nil || length == 0 {
		return nil, false
	}
	defer C.freeClipboardData(imageData)
	return C.GoBytes(imageData, C.int(length)), true
}

func newClipboardImpl() *macosClipboard {
	return &macosClipboard{}
}
