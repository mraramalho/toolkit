// toolkit is a versatile Go helper library for handling common web development tasks,
// including file uploads, directory management, string manipulation, and forced file downloads.
package toolkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const randStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will
// have access to all the methods with the reciever *Tools
type Tools struct {
	MaxFileSize        int
	AllowedFileTypes   []string
	MaxJSONSize        int
	AllowUnknownFields bool
	signalChan         chan os.Signal
}

// New returns an instance of Tools
func New() *Tools {
	return &Tools{}
}

// RunServer starts a web server with support for HTTP or HTTPS and implements a graceful shutdown.
//
// Parameters:
//   - ctx: A context that, when canceled, will trigger the server shutdown.
//   - srv: A pointer to an http.Server instance.
//   - shutdownTimeout: The maximum time to wait for active requests to finish before forcing closure.
//   - certKeyFiles: An optional variadic slice of strings.
//     If exactly two strings are provided, they are treated as [certFile, keyFile] for TLS.
//     If omitted, the function checks srv.TLSConfig or defaults to standard HTTP.
//
// The method blocks until a termination signal (SIGINT, SIGTERM) is received,
// the context is canceled, or the server encounters a fatal error.
func (t *Tools) RunServer(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration, certKeyFiles ...string) error {
	serverErrChan := make(chan error, 1)

	go func() {
		var err error

		// Determine if we should use TLS
		if len(certKeyFiles) == 2 {
			log.Printf("starting HTTPS server on %s", srv.Addr)
			err = srv.ListenAndServeTLS(certKeyFiles[0], certKeyFiles[1])
		} else if srv.TLSConfig != nil && (len(srv.TLSConfig.Certificates) > 0 || srv.TLSConfig.GetCertificate != nil) {
			log.Printf("starting HTTPS server on %s (using TLSConfig)", srv.Addr)
			err = srv.ListenAndServeTLS("", "") // Use certs from TLSConfig
		} else {
			log.Printf("starting HTTP server on %s", srv.Addr)
			err = srv.ListenAndServe()
		}

		if !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	stop := t.signalChan
	if stop == nil {
		stop = make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	}

	select {
	case err := <-serverErrChan:
		return err
	case <-stop:
		log.Println("shutdown signal received")
	case <-ctx.Done():
		log.Println("context canceled")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		if closeErr := srv.Close(); closeErr != nil {
			return fmt.Errorf("server forced to close: %w", errors.Join(err, closeErr))
		}
		return err
	}

	log.Println("server exited gracefully")
	return nil
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

// JSONResponse is the type fo sending json around
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1024 * 1024
	if t.MaxJSONSize > 0 {
		maxBytes = t.MaxJSONSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)

	if !t.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(&data)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must no be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("error unmarshaling JSON: %s", err.Error())

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only one JSON value")
	}

	return nil
}

// TODO: Writing Json

func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {

	return nil
}

// TODO: Push Json to a remote server
