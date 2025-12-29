package toolkit

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const randStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will
// have access to all the methods with the reciever *Tools
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// New returns an instance of Tools
func New() *Tools {
	return &Tools{}
}

// RandomString generates a safe random string of length l, using randStringSource as source
// for the string.
func (t *Tools) RandomString(l int) string {
	res := make([]byte, l)
	for i := range res {
		n := rand.IntN(len(randStringSource))
		res[i] = randStringSource[n]
	}
	return string(res)
}

// UploadedFile saves information about an uploaded file
type UploadedFile struct {
	OriginalFileName string
	NewFileName      string
	FileSize         int64
}

// UploadFiles uploads an slice of files to a server
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	if err := t.CreateDirIfNotExists(uploadDir, 0755); err != nil {
		return nil, err
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	r.Body = http.MaxBytesReader(nil, r.Body, int64(t.MaxFileSize))

	if err := r.ParseMultipartForm(int64(t.MaxFileSize)); err != nil {
		return nil, errors.New("the uploaded file is too big.")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {

			uploadedFile, err := func() (*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buffer := make([]byte, 512)
				if _, err = infile.Read(buffer); err != nil {
					return nil, err
				}

				allowed := false
				contenType := http.DetectContentType(buffer)
				if len(t.AllowedFileTypes) > 0 {
					for _, ft := range t.AllowedFileTypes {
						if strings.EqualFold(contenType, ft) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("invalid file type")
				}

				if _, err := infile.Seek(0, 0); err != nil {
					return nil, err
				}

				uploadedFile.OriginalFileName = hdr.Filename

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				outfile, err := os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName))
				if err != nil {
					return nil, err
				}

				defer outfile.Close()
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize

				return &uploadedFile, nil

			}()

			if err != nil {
				return uploadedFiles, err
			}
			uploadedFiles = append(uploadedFiles, uploadedFile)
		}
	}
	return uploadedFiles, nil
}

func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}
	return files[0], nil
}

// CreateDirIfNotExists creates a dir if it does not exist
func (t *Tools) CreateDirIfNotExists(path string, mode os.FileMode) error {
	if err := os.MkdirAll(path, mode); err != nil {
		return err
	}
	return nil
}

// Slugfy creates a simple slug from a string
func (t *Tools) Slugfy(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string not allowed")
	}

	re := regexp.MustCompile(`[^a-z\d]+`)
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(slug) == 0 {
		return "", errors.New("empty string, after slug process")
	}

	return slug, nil
}

// DownloadStaticFile downloads a file, and tries to force the browser to avoid displaying it
// in the browser window by setting content disposition. It also allows specification of
// the display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	filePath := filepath.Join(p, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))

	http.ServeFile(w, r, filePath)
}


// WORKING WITH JSON

// TODO: Reading Json
// TODO: Writing Json
// TODO: Push Json to a remote server
