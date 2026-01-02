# Toolkit

A versatile Go helper library for handling common web development tasks, including file uploads, graceful server management, JSON processing, and utility functions.

## Features

* **Graceful Server Shutdown**: Run HTTP or HTTPS servers that handle termination signals (`os.Interrupt`, `SIGTERM`) and context cancellation without dropping active requests.
* **JSON Processing**: Securely decode requests with size limits, encode responses with custom headers, and handle errors with injectable templates.
* **Remote JSON Posting**: Push JSON data to remote services and receive responses with ease.
* **Multi-File Uploads**: Easily handle single or multiple file uploads with built-in MIME type validation and size limits.
* **Security**: Enforce maximum file size limits, validate file types, and prevent path injection.
* **File Renaming**: Automatically generate safe, random filenames to prevent overwriting and path injection.
* **Slug Generation**: Create URL-friendly slugs from any string for clean SEO-friendly paths.
* **Forced Downloads**: Force browsers to download files as attachments with custom display names.

---

## Installation

```bash
go get github.com/mraramalho/toolkit

```

## Usage

First, import the package and initialize the `Tools` struct:

```go
import "github.com/yourusername/toolkit"

var t toolkit.Tools

```

### 1. Graceful Server Management

The `RunServer` method manages the server lifecycle, listening for system interrupts to ensure a clean exit.

```go
func main() {
    t := toolkit.New()
    srv := &http.Server{Addr: ":8080", Handler: myRouter}
    
    // Shut down gracefully with a 30-second timeout.
    // To use HTTPS, pass the cert and key paths as the final arguments.
    err := t.RunServer(context.Background(), srv, 30*time.Second)
    if err != nil {
        log.Fatal(err)
    }
}

```

### 2. Advanced JSON & Error Handling

You can read and write JSON using the toolkit's secure defaults. You can also inject a custom error template to maintain API consistency between microservices.

```go
// Optional: Define a custom error structure for your project
type MyError struct {
    Code    int    `json:"status_code"`
    Message string `json:"msg"`
}

func (e MyError) Prepare(err error, status int) any {
    return MyError{Code: status, Message: err.Error()}
}

func JSONHandler(w http.ResponseWriter, r *http.Request) {
    t.ErrorResponseTemplate = MyError{} // Inversion of Control
    
    var payload struct{ Name string `json:"name"` }

    if err := t.ReadJSON(w, r, &payload); err != nil {
        t.ErrorJSON(w, err, http.StatusBadRequest)
        return
    }

    t.WriteJSON(w, http.StatusAccepted, toolkit.JSONResponse{
        Error: false, 
        Message: "Processed",
    })
}

```

### 3. Remote JSON Pushing

Send data to a remote service and get the response.

```go
resp, statusCode, err := t.PushJSONToRemote("https://api.example.com/data", myData)
if err == nil {
    defer resp.Body.Close()
}

```

### 4. Handling File Uploads

Restrict file types and sizes before processing uploads. Files can be automatically renamed to prevent collisions.

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    t.MaxFileSize = 1024 * 1024 * 5 // 5MB
    t.AllowedFileTypes = []string{"image/jpeg", "image/png"}

    files, err := t.UploadFiles(r, "./uploads", true)
    if err != nil {
        t.ErrorJSON(w, err)
        return
    }
}

```

### 5. Generating Slugs

Convert strings into URL-safe formats (e.g., `"Hello World!"` -> `"hello-world"`).

```go
slug, _ := t.Slugfy("My Awesome Post Title @2025")
// Output: my-awesome-post-title-2025

```

### 6. Forced File Downloads

Use `DownloadStaticFile` to ensure a file is downloaded rather than opened in the browser window.

```go
func Download(w http.ResponseWriter, r *http.Request) {
    t.DownloadStaticFile(w, r, "./files", "internal_id.pdf", "Invoice_Dec_2025.pdf")
}

```
### 7. Utility Methods

* **`RandomString(len)`**: Generates a random string using a safe character set.
* **`CreateDirIfNotExists(path, mode)`**: Recursively creates folders if they are missing.
---

## API Reference

### `Tools` Configuration

| Field | Type | Description |
| --- | --- | --- |
| `MaxFileSize` | `int` | Maximum allowed size in bytes for file uploads. |
| `MaxJSONSize` | `int` | Maximum allowed size in bytes for JSON bodies (defaults to 1MB). |
| `AllowedFileTypes` | `[]string` | Slice of allowed MIME types for uploads. |
| `AllowUnknownFields` | `bool` | If false, `ReadJSON` returns an error if the body contains extra keys. |
| `ErrorResponseTemplate` | `ErrorTemplate` | Interface to inject custom error structures. |

### Methods Summary

* **`RunServer`**: Cross-platform HTTP/HTTPS server with graceful shutdown.
* **`ReadJSON / WriteJSON`**: Secure JSON decoding/encoding.
* **`ErrorJSON`**: Standardized error responses using templates.
* **`PushJSONToRemote`**: Simplified HTTP POST for JSON data.
* **`UploadFiles`**: Processes multipart form uploads and returns metadata.
* **`UploadOneFile`**: Convenience method for handling a single file upload.
* **`Slugfy`**: Returns a cleaned, lowercase, hyphenated string.
* **`RandomString`**: Generates a secure random string of specified length.
* **`CreateDirIfNotExists`**: Helper to ensure a directory structure exists on disk.
* **`DownloadStaticFile`**: Forces a file download via `Content-Disposition`.
---

## License

MIT

@DECOFMA WAS HERE.