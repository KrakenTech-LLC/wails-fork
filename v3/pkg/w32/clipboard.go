//go:build windows

/*
 * Based on code originally from https://github.com/atotto/clipboard. Copyright (c) 2013 Ato Araki. All rights reserved.
 */

package w32

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	cfUnicodetext = 13
	gmemMoveable  = 0x0002
)

var (
	procRegisterClipboardFormat = moduser32.NewProc("RegisterClipboardFormatW")
	procGetDIBits               = modgdi32.NewProc("GetDIBits")
	procGlobalSize              = kernel32.NewProc("GlobalSize")
)

// waitOpenClipboard opens the clipboard, waiting for up to a second to do so.
func waitOpenClipboard() error {
	started := time.Now()
	limit := started.Add(time.Second)
	var r uintptr
	var err error
	for time.Now().Before(limit) {
		r, _, err = procOpenClipboard.Call(0)
		if r != 0 {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return err
}

func GetClipboardText() (string, error) {
	// LockOSThread ensure that the whole method will keep executing on the same thread from begin to end (it actually locks the goroutine thread attribution).
	// Otherwise if the goroutine switch thread during execution (which is a common practice), the OpenClipboard and CloseClipboard will happen on two different threads, and it will result in a clipboard deadlock.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if formatAvailable, _, err := procIsClipboardFormatAvailable.Call(cfUnicodetext); formatAvailable == 0 {
		return "", err
	}
	err := waitOpenClipboard()
	if err != nil {
		return "", err
	}

	h, _, err := procGetClipboardData.Call(cfUnicodetext)
	if h == 0 {
		_, _, _ = procCloseClipboard.Call()
		return "", err
	}

	l, _, err := kernelGlobalLock.Call(h)
	if l == 0 {
		_, _, _ = procCloseClipboard.Call()
		return "", err
	}

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(l))[:])

	r, _, err := kernelGlobalUnlock.Call(h)
	if r == 0 {
		_, _, _ = procCloseClipboard.Call()
		return "", err
	}

	closed, _, err := procCloseClipboard.Call()
	if closed == 0 {
		return "", err
	}
	return text, nil
}

func SetClipboardText(text string) error {
	// LockOSThread ensure that the whole method will keep executing on the same thread from begin to end (it actually locks the goroutine thread attribution).
	// Otherwise if the goroutine switch thread during execution (which is a common practice), the OpenClipboard and CloseClipboard will happen on two different threads, and it will result in a clipboard deadlock.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := waitOpenClipboard()
	if err != nil {
		return err
	}

	r, _, err := procEmptyClipboard.Call(0)
	if r == 0 {
		_, _, _ = procCloseClipboard.Call()
		return err
	}

	data, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	// "If the hMem parameter identifies a memory object, the object must have
	// been allocated using the function with the GMEM_MOVEABLE flag."
	h, _, err := kernelGlobalAlloc.Call(gmemMoveable, uintptr(len(data)*int(unsafe.Sizeof(data[0]))))
	if h == 0 {
		_, _, _ = procCloseClipboard.Call()
		return err
	}
	defer func() {
		if h != 0 {
			kernelGlobalFree.Call(h)
		}
	}()

	l, _, err := kernelGlobalLock.Call(h)
	if l == 0 {
		_, _, _ = procCloseClipboard.Call()
		return err
	}

	r, _, err = kernelLstrcpy.Call(l, uintptr(unsafe.Pointer(&data[0])))
	if r == 0 {
		_, _, _ = procCloseClipboard.Call()
		return err
	}

	r, _, err = kernelGlobalUnlock.Call(h)
	if r == 0 {
		if err.(syscall.Errno) != 0 {
			_, _, _ = procCloseClipboard.Call()
			return err
		}
	}

	r, _, err = procSetClipboardData.Call(cfUnicodetext, h)
	if r == 0 {
		_, _, _ = procCloseClipboard.Call()
		return err
	}
	h = 0 // suppress deferred cleanup
	closed, _, err := procCloseClipboard.Call()
	if closed == 0 {
		return err
	}
	return nil
}

func GetClipboardImage() ([]byte, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	pngFormat := registerClipboardFormat("PNG")
	if pngFormat != 0 && IsClipboardFormatAvailable(pngFormat) {
		imageData, err := getClipboardPNG(pngFormat)
		if err == nil {
			return imageData, nil
		}
	}

	if !IsClipboardFormatAvailable(CF_BITMAP) {
		return nil, syscall.EINVAL
	}

	if err := waitOpenClipboard(); err != nil {
		return nil, err
	}
	defer procCloseClipboard.Call()

	handle, _, err := procGetClipboardData.Call(uintptr(CF_BITMAP))
	if handle == 0 {
		return nil, err
	}

	var bmp BITMAP
	if GetObject(HGDIOBJ(handle), unsafe.Sizeof(bmp), unsafe.Pointer(&bmp)) == 0 {
		return nil, syscall.GetLastError()
	}

	width := int(bmp.BmWidth)
	height := int(bmp.BmHeight)
	if width <= 0 || height <= 0 {
		return nil, syscall.EINVAL
	}

	hdc := CreateCompatibleDC(0)
	if hdc == 0 {
		return nil, syscall.GetLastError()
	}
	defer DeleteDC(hdc)

	oldBitmap := SelectObject(hdc, HGDIOBJ(HBITMAP(handle)))
	defer SelectObject(hdc, oldBitmap)

	var bi BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	bi.BmiHeader.BiWidth = bmp.BmWidth
	bi.BmiHeader.BiHeight = bmp.BmHeight
	bi.BmiHeader.BiPlanes = 1
	bi.BmiHeader.BiBitCount = 32
	bi.BmiHeader.BiCompression = BI_RGB

	pixels := make([]byte, width*height*4)
	ret, _, err := procGetDIBits.Call(
		uintptr(hdc),
		handle,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, err
	}

	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceIndex := ((height-1-y)*width + x) * 4
			targetIndex := result.PixOffset(x, y)
			result.Pix[targetIndex+0] = pixels[sourceIndex+2]
			result.Pix[targetIndex+1] = pixels[sourceIndex+1]
			result.Pix[targetIndex+2] = pixels[sourceIndex+0]
			result.Pix[targetIndex+3] = pixels[sourceIndex+3]
		}
	}

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, result); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func SetClipboardImage(data []byte) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sourceImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}

	rgba := image.NewRGBA(sourceImage.Bounds())
	draw.Draw(rgba, rgba.Bounds(), sourceImage, sourceImage.Bounds().Min, draw.Src)

	hBitmap, err := CreateHBITMAPFromImage(rgba)
	if err != nil {
		return err
	}
	releaseBitmap := true
	defer func() {
		if releaseBitmap {
			DeleteObject(HGDIOBJ(hBitmap))
		}
	}()

	var pngBuffer bytes.Buffer
	if err := png.Encode(&pngBuffer, sourceImage); err != nil {
		return err
	}

	pngHandle, err := newClipboardDataHandle(pngBuffer.Bytes())
	if err != nil {
		return err
	}
	releasePNG := true
	defer func() {
		if releasePNG {
			kernelGlobalFree.Call(pngHandle)
		}
	}()

	if err := waitOpenClipboard(); err != nil {
		return err
	}

	r, _, err := procEmptyClipboard.Call(0)
	if r == 0 {
		procCloseClipboard.Call()
		return err
	}

	if pngFormat := registerClipboardFormat("PNG"); pngFormat != 0 {
		r, _, err = procSetClipboardData.Call(uintptr(pngFormat), pngHandle)
		if r == 0 {
			procCloseClipboard.Call()
			return err
		}
		releasePNG = false
	}

	r, _, err = procSetClipboardData.Call(uintptr(CF_BITMAP), uintptr(hBitmap))
	if r == 0 {
		procCloseClipboard.Call()
		return err
	}
	releaseBitmap = false

	closed, _, err := procCloseClipboard.Call()
	if closed == 0 {
		return err
	}
	return nil
}

func registerClipboardFormat(name string) uint {
	nameUTF16, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0
	}
	ret, _, _ := procRegisterClipboardFormat.Call(uintptr(unsafe.Pointer(nameUTF16)))
	return uint(ret)
}

func getClipboardPNG(format uint) ([]byte, error) {
	if err := waitOpenClipboard(); err != nil {
		return nil, err
	}
	defer procCloseClipboard.Call()

	handle, _, err := procGetClipboardData.Call(uintptr(format))
	if handle == 0 {
		return nil, err
	}

	locked, _, err := kernelGlobalLock.Call(handle)
	if locked == 0 {
		return nil, err
	}
	defer kernelGlobalUnlock.Call(handle)

	size, _, err := procGlobalSize.Call(handle)
	if size == 0 {
		return nil, err
	}

	data := make([]byte, int(size))
	copy(data, unsafe.Slice((*byte)(unsafe.Pointer(locked)), int(size)))
	return data, nil
}

func newClipboardDataHandle(data []byte) (uintptr, error) {
	handle, _, err := kernelGlobalAlloc.Call(gmemMoveable, uintptr(len(data)))
	if handle == 0 {
		return 0, err
	}

	locked, _, err := kernelGlobalLock.Call(handle)
	if locked == 0 {
		kernelGlobalFree.Call(handle)
		return 0, err
	}

	copy(unsafe.Slice((*byte)(unsafe.Pointer(locked)), len(data)), data)

	r, _, unlockErr := kernelGlobalUnlock.Call(handle)
	if r == 0 && unlockErr.(syscall.Errno) != 0 {
		kernelGlobalFree.Call(handle)
		return 0, unlockErr
	}

	return handle, nil
}
