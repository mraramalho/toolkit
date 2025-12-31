# Toolkit

A versatile Go helper library for handling common web development tasks, including file uploads, graceful server management, directory handling, and string manipulation.

## Features

* **Graceful Server Shutdown**: Run HTTP or HTTPS servers that handle termination signals (`os.Interrupt`) and context cancellation without dropping active requests.
* **JSON Processing**: Securely decode requests with size limits and encode responses with custom headers and status codes.
* **Multi-File Uploads**: Easily handle single or multiple file uploads with built-in MIME type validation and size limits.
* **Security**: Validate file types (MIME) and enforce maximum file size limits.
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

### 2. Working with JSON

The toolkit provides a symmetrical way to read and write JSON, handling error states and headers automatically.

```go
func JSONHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Name string `json:"name"`
    }

    // Read JSON securely
    if err := t.ReadJSON(w, r, &payload); err != nil {
        t.WriteJSON(w, http.StatusBadRequest, toolkit.JSONResponse{
            Error:   true,
            Message: err.Error(),
        })
        return
    }

    // Write a standardized JSON response
    response := toolkit.JSONResponse{
        Error:   false,
        Message: "Success",
        Data:    payload,
    }

    t.WriteJSON(w, http.StatusAccepted, response)
}

```

### 3. Handling File Uploads

Restrict file types and sizes before processing uploads. Files can be automatically renamed to prevent collisions.

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    t.MaxFileSize = 1024 * 1024 * 5 // 5MB limit
    t.AllowedFileTypes = []string{"image/jpeg", "image/png", "application/pdf"}

    // Upload files to the "uploads" folder and rename them randomly
    files, err := t.UploadFiles(r, "./uploads", true)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("Uploaded %d files successfully.\n", len(files))
}

```

### 4. Generating Slugs

Convert strings into URL-safe formats (e.g., `"Hello World!"` -> `"hello-world"`).

```go
slug, _ := t.Slugfy("My Awesome Post Title @2025")
// Output: my-awesome-post-title-2025

```

### 5. Forced File Downloads

Use `DownloadStaticFile` to ensure a file is downloaded rather than opened in the browser window.

```go
func Download(w http.ResponseWriter, r *http.Request) {
    t.DownloadStaticFile(w, r, "./files", "internal_id.pdf", "Invoice_Dec_2025.pdf")
}

```
### 6. Utility Methods

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

### Methods Summary

* **`RunServer`**: Starts an HTTP/HTTPS server with cross-platform graceful shutdown logic.
* **`ReadJSON`**: Decodes a JSON request body into a pointer with error wrapping.
* **`WriteJSON`**: Encodes a response into JSON, sets the `Content-Type`, and writes headers.
* **`UploadFiles`**: Processes multipart form uploads and returns metadata.
* **`UploadOneFile`**: Convenience method for handling a single file upload.
* **`Slugfy`**: Returns a cleaned, lowercase, hyphenated string.
* **`RandomString`**: Generates a secure random string of specified length.
* **`CreateDirIfNotExists`**: Helper to ensure a directory structure exists on disk.
* **`DownloadStaticFile`**: Forces a file download via `Content-Disposition`.

---

## Roadmap

* [ ] Write JSON to response
* [ ] Push JSON to a remote server

## License

MIT
