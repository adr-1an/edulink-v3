export function uploadToPresignedURL(
    file: File,
    url: string,
    onProgress: (progress: number) => void,
) {
    return new Promise<void>((resolve, reject) => {
        const request = new XMLHttpRequest()
        request.open("PUT", url)
        request.setRequestHeader("Content-Type", file.type)
        request.upload.addEventListener("progress", (event) => {
            if (event.lengthComputable) {
                onProgress(Math.min(99, Math.round(event.loaded / event.total * 100)))
            }
        })
        request.addEventListener("load", () => {
            if (request.status >= 200 && request.status < 300) {
                onProgress(100)
                resolve()
            } else {
                reject(new Error(`Storage returned HTTP ${request.status}.`))
            }
        })
        request.addEventListener("error", () => reject(new Error("The browser could not reach file storage.")))
        request.addEventListener("abort", () => reject(new Error("The upload was canceled.")))
        request.send(file)
    })
}
