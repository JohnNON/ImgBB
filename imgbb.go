package imgbb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	endpoint = "https://api.imgbb.com/1/upload"
	host     = "imgbb.com"
	origin   = "https://imgbb.com"
	referer  = "https://imgbb.com/"

	maxSize = 33554432
)

var (
	// ErrFileEmpty is an error for empty image file.
	ErrFileEmpty = errors.New("image file is empty")

	// ErrFileSize is an error for too large image size.
	ErrFileSize = errors.New("image is too large (max image size is 32mb)")
)

// Image is a struct with image data to upload.
type Image struct {
	name       string
	size       int
	expiration string
	file       []byte
}

// NewImage creates a new Image.
func NewImage(name string, expiration string, file []byte) (*Image, error) {
	size := len(file)

	if size <= 0 {
		return nil, ErrFileEmpty
	}

	if size > maxSize {
		return nil, ErrFileSize
	}

	return &Image{
		name:       name,
		size:       size,
		expiration: expiration,
		file:       file,
	}, nil
}

// Error is an upload error response.
type Error struct {
	StatusCode int     `json:"status_code"`
	StatusText string  `json:"status_txt"`
	ErrInfo    ErrInfo `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s: %v", e.StatusCode, e.StatusText, e.ErrInfo)
}

// ErrInfo is an upload error info response.
type ErrInfo struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Context string `json:"context"`
}

// Response is an upload success response.
type Response struct {
	Data       Data `json:"data"`
	StatusCode int  `json:"status"`
	Success    bool `json:"success"`
}

// Data is an information about uploaded file.
type Data struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URLViewer  string `json:"url_viewer"`
	URL        string `json:"url"`
	DisplayURL string `json:"display_url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int    `json:"size"`
	Time       int64  `json:"time"`
	Expiration int64  `json:"expiration"`
	Image      Info   `json:"image"`
	Thumb      Info   `json:"thumb"`
	Medium     Info   `json:"medium"`
	DeleteURL  string `json:"delete_url"`
}

// Info is an additional info about uploaded file.
type Info struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}

type Option func(*ImgBB)

func WithEndpoint(endpoint string) Option {
	return func(imgBB *ImgBB) {
		imgBB.endpoint = endpoint
	}
}

// ImgBB is a ImgBB api client.
type ImgBB struct {
	client http.Client

	key string

	endpoint string
}

// New create a new ImgBB api client.
func New(client http.Client, key string, opts ...Option) *ImgBB {
	imgBB := &ImgBB{
		client:   client,
		key:      key,
		endpoint: endpoint,
	}

	for _, o := range opts {
		o(imgBB)
	}

	return imgBB
}

// Upload is a function to upload image to ImgBB.
func (i *ImgBB) Upload(ctx context.Context, img *Image) (*Response, error) {
	req, err := i.prepareRequest(ctx, img)
	if err != nil {
		return nil, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, Error{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			ErrInfo: ErrInfo{
				Message: fmt.Sprintf("http client request do: %v", err),
			},
		}
	}
	defer resp.Body.Close()

	return i.respParse(resp)
}

func (i *ImgBB) prepareRequest(ctx context.Context, img *Image) (*http.Request, error) {
	pipeReader, pipeWriter := io.Pipe()

	mpWriter := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer mpWriter.Close()

		err := mpWriter.WriteField("key", i.key)
		if err != nil {
			return
		}

		err = mpWriter.WriteField("type", "file")
		if err != nil {
			return
		}

		err = mpWriter.WriteField("action", "upload")
		if err != nil {
			return
		}

		if len(img.expiration) > 0 {
			err = mpWriter.WriteField("expiration", img.expiration)
			if err != nil {
				return
			}
		}

		part, err := mpWriter.CreateFormFile("image", img.name)
		if err != nil {
			return
		}

		if _, err = io.Copy(part, bytes.NewReader(img.file)); err != nil {
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, i.endpoint, pipeReader)
	if err != nil {
		return nil, Error{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			ErrInfo: ErrInfo{
				Message: fmt.Sprintf("new request: %v", err),
			},
		}
	}

	req.Header.Add("Content-Type", mpWriter.FormDataContentType())
	req.Header.Add("Host", host)
	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)

	return req, nil
}

func (i *ImgBB) respParse(resp *http.Response) (*Response, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Error{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			ErrInfo: ErrInfo{
				Message: fmt.Sprintf("read response body: %v", err),
			},
		}
	}

	if resp.StatusCode == http.StatusOK {
		var res Response
		if err := json.Unmarshal(data, &res); err != nil {
			return nil, Error{
				StatusCode: http.StatusInternalServerError,
				StatusText: http.StatusText(http.StatusInternalServerError),
				ErrInfo: ErrInfo{
					Message: fmt.Sprintf("json unmarshal: %v", err),
				},
			}
		}

		return &res, nil
	}

	var errRes Error
	if err := json.Unmarshal(data, &errRes); err != nil {
		return nil, Error{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			ErrInfo: ErrInfo{
				Message: fmt.Sprintf("json unmarshal: %v", err),
			},
		}
	}

	return nil, errRes
}
