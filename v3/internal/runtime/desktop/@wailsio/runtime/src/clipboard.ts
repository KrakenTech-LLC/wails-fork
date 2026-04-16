/*
 _	   __	  _ __
| |	 / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

import {newRuntimeCaller, objectNames} from "./runtime.js";

const call = newRuntimeCaller(objectNames.Clipboard);

const ClipboardSetText = 0;
const ClipboardText = 1;
const ClipboardSetImage = 2;
const ClipboardImage = 3;
const ClipboardGetContent = 4;

export type ClipboardContentKind = "empty" | "text" | "image";

export interface ClipboardContentResult {
    kind: ClipboardContentKind;
    text?: string;
    data?: string;
    mimeType?: string;
}

/**
 * Sets the text to the Clipboard.
 *
 * @param text - The text to be set to the Clipboard.
 * @return A Promise that resolves when the operation is successful.
 */
export function SetText(text: string): Promise<void> {
    return call(ClipboardSetText, {text});
}

/**
 * Get the Clipboard text
 *
 * @returns A promise that resolves with the text from the Clipboard.
 */
export function Text(): Promise<string> {
    return call(ClipboardText);
}

/**
 * Sets an image on the Clipboard using base64-encoded image bytes.
 *
 * The runtime accepts any image format the platform decoder understands
 * and normalises reads back to PNG bytes.
 *
 * @param data - Base64-encoded image bytes.
 * @param mimeType - Optional source mime type.
 * @return A Promise that resolves when the operation is successful.
 */
export function SetImage(data: string, mimeType?: string): Promise<void> {
    return call(ClipboardSetImage, {data, mimeType});
}

/**
 * Gets the current clipboard image as base64-encoded PNG bytes.
 *
 * @returns A promise that resolves with PNG image bytes in base64 form.
 */
export function Image(): Promise<string> {
    return call(ClipboardImage);
}

/**
 * Gets the current clipboard contents, describing whether it is text, image, or empty.
 *
 * Images are returned as base64-encoded PNG bytes.
 */
export function Content(): Promise<ClipboardContentResult> {
    return call(ClipboardGetContent);
}
