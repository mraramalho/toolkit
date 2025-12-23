package toolkit

import (
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRandomString(t *testing.T) {
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
	// Casos de teste de 1 arquivo (note as chaves {} dentro do slice de nomes)
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
	// NOVO CASO: Múltiplos Arquivos
	// Estou usando o mesmo arquivo físico 3 vezes para simular o upload de 3 arquivos diferentes
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

		info, err := os.Stat(e.dirName)
		if os.IsNotExist(err) || !info.IsDir() {
			if err := os.MkdirAll(e.dirName, 0755); err != nil {
				t.Error(err)
			}
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
					t.Errorf("creating form file error: %s", err.Error())
					return
				}

				f, err := os.Open(filePath)
				if err != nil {
					t.Errorf("opening file error: %s", err.Error())
					return
				}

				_, err = io.Copy(part, f)
				f.Close()

				if err != nil {
					t.Errorf("error copying file bytes: %s", err.Error())
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

		os.MkdirAll(uploadDir, 0755)

		uploadedFiles, err := testTools.UploadFiles(request, uploadDir, e.renameFile)

		if err != nil && !e.expectsError {
			t.Error("test error:", err, "for test:", e.testName)
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

		wg.Wait()
	}
}
