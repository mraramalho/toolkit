# Toolkit

A versatile Go helper library for handling common web development tasks, including file uploads, directory management, string manipulation, and forced file downloads.

## Features

* **Multi-File Uploads**: Easily handle single or multiple file uploads with built-in validation.
* **Security**: Validate file types (MIME) and enforce maximum file size limits.
* **File Renaming**: Automatically generate safe, random filenames to prevent overwriting and path injection.
* **Slug Generation**: Create URL-friendly slugs from any string.
* **File Downloads**: Force browsers to download files as attachments with custom display names.

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

### 1. Handling File Uploads

You can restrict file types and sizes before processing uploads.

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Optional configuration
    t.MaxFileSize = 1024 * 1024 * 5 // 5MB
    t.AllowedFileTypes = []string{"image/jpeg", "image/png", "application/pdf"}

    // Upload files to the "uploads" directory and rename them randomly
    files, err := t.UploadFiles(r, "./uploads", true)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Access uploaded file info
    for _, f := range files {
        fmt.Println("Saved as:", f.NewFileName)
    }
}

```

### 2. Generating Slugs

Convert strings into URL-safe formats (e.g., "Hello World!" -> "hello-world").

```go
slug, err := t.Slugfy("My Awesome Post Title @2024")
if err == nil {
    fmt.Println(slug) // Output: my-awesome-post-title-2024
}

```

### 3. Forcing File Downloads

Use `DownloadStaticFile` to ensure a file is downloaded rather than opened in the browser.

```go
func Download(w http.ResponseWriter, r *http.Request) {
    path := "./files"
    serverFile := "internal_id_123.pdf"
    displayName := "User_Invoice.pdf"

    t.DownloadStaticFile(w, r, path, serverFile, displayName)
}

```

### 4. Utility Methods

* **`RandomString(len)`**: Generates a random string using a safe character set.
* **`CreateDirIfNotExists(path, mode)`**: Recursively creates folders if they are missing.

---

## API Reference

### `Tools` Configuration

| Field | Type | Description |
| --- | --- | --- |
| `MaxFileSize` | `int` | Maximum allowed size in bytes (defaults to 1GB). |
| `AllowedFileTypes` | `[]string` | Slice of allowed MIME types (e.g., `image/jpeg`). |

### Methods Summary

* **`UploadFiles`**: Returns a slice of `UploadedFile` metadata.
* **`UploadOneFile`**: Convenience method for handling a single file.
* **`Slugfy`**: Returns a cleaned, lowercase, hyphenated string.
* **`RandomString`**: Returns a string of specified length.

---

## Roadmap

* [ ] Read JSON to struct
* [ ] Write JSON to response
* [ ] Push JSON to remote API

## License

[MIT](https://www.google.com/search?q=LICENSE)
