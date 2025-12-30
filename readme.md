# Toolkit

A versatile Go helper library for handling common web development tasks, including file uploads, graceful server management, directory handling, and string manipulation.

## Features

* **Graceful Server Shutdown**: Run HTTP or HTTPS servers that handle termination signals (`os.Interrupt`) and context cancellation without dropping active requests.
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

### 2. Handling File Uploads

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

### 3. Generating Slugs

Convert strings into URL-safe formats (e.g., `"Hello World!"` -> `"hello-world"`).

```go
slug, _ := t.Slugfy("My Awesome Post Title @2025")
// Output: my-awesome-post-title-2025

```

### 4. Forced File Downloads

Use `DownloadStaticFile` to ensure a file is downloaded rather than opened in the browser window.

```go
func Download(w http.ResponseWriter, r *http.Request) {
    t.DownloadStaticFile(w, r, "./files", "internal_id.pdf", "Invoice_Dec_2025.pdf")
}

```
### 5. Utility Methods

* **`RandomString(len)`**: Generates a random string using a safe character set.
* **`CreateDirIfNotExists(path, mode)`**: Recursively creates folders if they are missing.
---

## API Reference

### `Tools` Configuration

| Field | Type | Description |
| --- | --- | --- |
| `MaxFileSize` | `int` | Maximum allowed size in bytes (defaults to 1GB). |
| `AllowedFileTypes` | `[]string` | Slice of allowed MIME types (e.g., `image/jpeg`, `application/pdf`). |

### Methods Summary

* **`RunServer`**: Starts an HTTP/HTTPS server with cross-platform graceful shutdown logic.
* **`UploadFiles`**: Processes multipart form uploads and returns metadata.
* **`UploadOneFile`**: Convenience method for handling a single file upload.
* **`Slugfy`**: Returns a cleaned, lowercase, hyphenated string.
* **`RandomString`**: Generates a secure random string of specified length.
* **`CreateDirIfNotExists`**: Helper to ensure a directory structure exists on disk.
* **`DownloadStaticFile`**: Forces a file download via `Content-Disposition`.

---

## Roadmap

* [ ] Read JSON to struct (with validation)
* [ ] Write JSON to response
* [ ] Push JSON to a remote server

## License

MIT