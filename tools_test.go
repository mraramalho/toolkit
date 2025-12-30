package toolkit

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestToolsRandomString(t *testing.T) {
	tools := new(Tools)
	str := tools.RandomString(10)

	if len(str) != 10 {
		t.Errorf("%s has %v letters", str, len(str))
	}

}

var uploadFileTest = []struct {
	testName         string
	expectsError     bool
	fileNames        []string // MUDOU: Agora é um slice de strings
	dirName          string
	renameFile       bool
	allowedFileTypes []string
	maxFileSize      int
	numberOfFiles    int
}{
	{
		testName:         "file too big",
		expectsError:     true,
		fileNames:        []string{"image.png"},
		dirName:          "./test-data",
		renameFile:       true,
		allowedFileTypes: []string{"image/png"},
		maxFileSize:      100,
		numberOfFiles:    1,
	},
	{
		testName:         "invalid file type",
		expectsError:     true,
		fileNames:        []string{"image.png"},
		dirName:          "./test-data",
		renameFile:       true,
		allowedFileTypes: []string{"image/jpeg"},
		maxFileSize:      1024 * 1024,
		numberOfFiles:    1,
	},
	{
		testName:         "no renaming",
		expectsError:     false,
		fileNames:        []string{"image.png"},
		dirName:          "./test-data",
		renameFile:       false,
		allowedFileTypes: []string{"image/png"},
		maxFileSize:      1024 * 1024 * 5,
		numberOfFiles:    1,
	},
	{
		testName:         "upload multiple files",
		expectsError:     false,
		fileNames:        []string{"image.png", "image.png", "image.png"},
		dirName:          "./test-data",
		renameFile:       true,
		allowedFileTypes: []string{"image/png"},
		maxFileSize:      1024 * 1024 * 10,
		numberOfFiles:    3,
	},
}

func TestToolsUploadFiles(t *testing.T) {
	for _, e := range uploadFileTest {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}

		err := os.MkdirAll(e.dirName, 0755)
		if err != nil {
			t.Fatalf("não foi possível garantir o diretório de teste: %v", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer pw.Close()
			defer writer.Close()

			for _, fileName := range e.fileNames {
				filePath := filepath.Join(e.dirName, fileName)

				part, err := writer.CreateFormFile("file", filePath)
				if err != nil {
					return
				}

				f, err := os.Open(filePath)
				if err != nil {
					return
				}

				_, err = io.Copy(part, f)
				f.Close()

				if err != nil {
					return
				}
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		testTools := Tools{
			AllowedFileTypes: e.allowedFileTypes,
			MaxFileSize:      int(e.maxFileSize),
		}

		uploadDir := filepath.Join(e.dirName, "uploads")

		uploadedFiles, err := testTools.UploadFiles(request, uploadDir, e.renameFile)

		if err != nil && !e.expectsError {
			t.Error("test error:", err, "for test:", e.testName)
		}

		if err == nil && e.expectsError {
			t.Errorf("%s: expected error but none found", e.testName)
		}

		if !e.expectsError && len(uploadedFiles) != e.numberOfFiles {
			t.Errorf("Expected %d files, got %d for test: %s", e.numberOfFiles, len(uploadedFiles), e.testName)
		}

		if len(uploadedFiles) > 0 && !e.renameFile {
			for _, upFile := range uploadedFiles {
				if upFile.NewFileName != upFile.OriginalFileName {
					t.Error("file renamed when should not")
				}
			}
		}

		for _, file := range uploadedFiles {
			path := filepath.Join(uploadDir, file.NewFileName)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", file.OriginalFileName, err.Error())
			}

			_ = os.Remove(path)
		}

		pr.Close()

		wg.Wait()
	}
}

func TestToolsUploadOneFile(t *testing.T) {
	for _, e := range uploadFileTest {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer pw.Close()
			defer writer.Close()

			filePath := filepath.Join(e.dirName, e.fileNames[0])
			part, err := writer.CreateFormFile("file", filePath)
			if err != nil {
				return
			}

			f, err := os.Open(filePath)
			if err != nil {
				return
			}

			_, err = io.Copy(part, f)
			f.Close()

			if err != nil {
				return
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		testTools := Tools{
			AllowedFileTypes: e.allowedFileTypes,
			MaxFileSize:      int(e.maxFileSize),
		}

		uploadDir := filepath.Join(e.dirName, "uploads")
		os.MkdirAll(uploadDir, 0755)

		uploadedFile, err := testTools.UploadOneFile(request, uploadDir, e.renameFile)

		if err != nil {
			if !e.expectsError {
				t.Errorf("%s: unexpected error: %s", e.testName, err.Error())
			}
		} else {
			if e.expectsError {
				t.Errorf("%s: expected error but none found", e.testName)
			}

			if uploadedFile != nil {
				if !e.renameFile {
					if uploadedFile.NewFileName != uploadedFile.OriginalFileName {
						t.Errorf("%s: file renamed when should not", e.testName)
					}
				}
				path := filepath.Join(uploadDir, uploadedFile.NewFileName)
				if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
					t.Errorf("%s: expected file to exist: %s", e.testName, statErr.Error())
				} else {
					_ = os.Remove(path)
				}
			}
		}
		pr.Close()
		wg.Wait()
	}
}

var testDirs = []struct {
	testName     string
	expectsError bool
	dirName      string
	mode         os.FileMode
	errorMsg     string
}{
	{"creates dir", false, "./tempDir", 0755, ""},
	{"directory already exists", false, "./tempDir", 0755, ""},
	{"creates subdirectory", false, "./tempDir/anotherTempDir", 0755, ""},
	{"dir cannot be created", true, "C:/Users/x0lc/tempDir", 0755, "mkdir C:/Users/x0lc: Access is denied."},
}

func TestToolsCreateDirIfNotExists(t *testing.T) {
	var testTools Tools
	for _, e := range testDirs {
		err := testTools.CreateDirIfNotExists(e.dirName, e.mode)
		if err != nil && !e.expectsError {
			t.Error("expected no error for test", e.testName, "but found one:", err)
		}

		if err == nil && e.expectsError {
			t.Error("expected one error for test", e.testName, "but none found:", err)
		}

		if e.expectsError && err.Error() != e.errorMsg {
			t.Error("wrong error received for test", e.testName, "expected:", e.errorMsg, "received:", err)
		}
	}

	if err := os.RemoveAll(testDirs[0].dirName); err != nil {
		t.Error("error removing temdirs:", err)
	}
}

var slugTestTable = []struct {
	testName       string
	expectsError   bool
	errorMsg       string
	stringToSlugfy string
	expectedSlug   string
}{
	{"simple slug transformation", false, "", "hello World 123", "hello-world-123"},
	{"all caps string", false, "", "HELLO WORLD ", "hello-world"},
	{"exclamation sign", false, "", "HELLO WORLD!", "hello-world"},
	{"empty slug after slugfy string", true, "empty string, after slug process", "!*%.", ""},
	{"empty string not allowed", true, "empty string not allowed", "", ""},
}

func TestToolsSlugfy(t *testing.T) {
	var testTools Tools
	for _, e := range slugTestTable {
		slug, err := testTools.Slugfy(e.stringToSlugfy)
		if err != nil && !e.expectsError {
			t.Error("unexpected error for test", e.testName, "error:", err)
		}

		if err == nil && e.expectsError {
			t.Error("expected error for test", e.testName, "error:", e.errorMsg, "but none found")
		}

		if err != nil && e.errorMsg != err.Error() {
			t.Error("unexpected error message for test", e.testName, "expected error message:", e.errorMsg, "found:", err.Error())
		}

		if slug != e.expectedSlug {
			t.Error("unexpected slug for test", e.testName, "expected slug:", e.expectedSlug, "slug received:", slug)
		}
	}

}

func TestToolsDownloadStaticFile(t *testing.T) {
	tmpDir := t.TempDir()
	fileName := "testfile.txt"
	content := []byte("conteúdo do arquivo de teste")
	filePath := filepath.Join(tmpDir, fileName)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	displayName := "meu-download.txt"
	tools := New()

	req := httptest.NewRequest("GET", "/download", nil)
	rr := httptest.NewRecorder()

	tools.DownloadStaticFile(rr, req, tmpDir, fileName, displayName)

	if rr.Code != 200 {
		t.Errorf("esperava status 200, recebeu %d", rr.Code)
	}

	expectedHeader := fmt.Sprintf("attachment; filename=\"%s\"", displayName)
	if got := rr.Header().Get("Content-Disposition"); got != expectedHeader {
		t.Errorf("cabeçalho Content-Disposition incorreto: esperado %s, recebeu %s", expectedHeader, got)
	}

	if rr.Body.String() != string(content) {
		t.Errorf("conteúdo do arquivo incorreto: esperado %s, recebeu %s", string(content), rr.Body.String())
	}
}

func TestToolsRunServer(t *testing.T) {
	tools := &Tools{}

	t.Run("Graceful Shutdown via Context", func(t *testing.T) {
		// Use an available port for testing
		srv := &http.Server{
			Addr: "localhost:8081",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Channel to capture the error from RunServer
		errChan := make(chan error)

		go func() {
			errChan <- tools.RunServer(ctx, srv, 2*time.Second)
		}()

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)

		// Trigger shutdown via context
		cancel()

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("expected nil error on graceful shutdown, got %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("server did not shut down within timeout")
		}
	})

	t.Run("Server Port Conflict", func(t *testing.T) {
		// Start a dummy listener on a port
		srv1 := &http.Server{Addr: "localhost:8082"}
		go srv1.ListenAndServe()
		defer srv1.Close()

		time.Sleep(100 * time.Millisecond)

		// Try to start our server on the same port
		srv2 := &http.Server{Addr: "localhost:8082"}
		ctx := context.Background()

		err := tools.RunServer(ctx, srv2, 1*time.Second)
		if err == nil {
			t.Error("expected an error due to port conflict, got nil")
		}
	})

	t.Run("Graceful Shutdown via Signal Agnostic", func(t *testing.T) {
		tools := &Tools{}
		
		testChan := make(chan os.Signal, 1)
		tools.signalChan = testChan

		srv := &http.Server{Addr: "localhost:0"}
		errChan := make(chan error, 1)

		go func() {
			errChan <- tools.RunServer(context.Background(), srv, 2*time.Second)
		}()

		time.Sleep(100 * time.Millisecond)

		testChan <- os.Interrupt

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not respond to injected signal")
		}
	})
}
