package imgbb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	endpoint = "https://api.imgbb.com/1/upload"
	host     = "imgbb.com"
	origin   = "https://imgbb.com"
	referer  = "https://imgbb.com/"
)

var (
	// ErrFileSize - is an error for too large image size
	ErrFileSize = errors.New("Image is too large. Max image size is 32mb")
)

// Image - is a struct with image data to upload
type Image struct {
	Name       string
	Size       int
	Expiration string
	File       []byte
}

// NewImage - create a new Image
func NewImage(name string, expiration string, file []byte) *Image {
	return &Image{
		Name:       name,
		Size:       len(file),
		Expiration: expiration,
		File:       file,
	}
}

// ImgBBError - is a struct with upload error response
type ImgBBError struct {
	StatusCode int     `json:"status_code"`
	StatusText string  `json:"status_txt"`
	Err        errInfo `json:"error"`
}

type errInfo struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Context string `json:"context"`
}

// ImgBBResult - is a struct with upload success response
type ImgBBResult struct {
	Data       data `json:"data"`
	StatusCode int  `json:"status"`
	Success    bool `json:"success"`
}

type data struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	UrlViewer  string `json:"url_viewer"`
	Url        string `json:"url"`
	DisplayUrl string `json:"display_url"`
	Size       int    `json:"size"`
	Time       string `json:"time"`
	Expiration string `json:"expiration"`
	Image      info   `json:"image"`
	Thumb      info   `json:"thumb"`
	Medium     info   `json:"medium"`
	DeleteUrl  string `json:"delete_url"`
}

type info struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	Url       string `json:"url"`
}

// ImgBB - is a struct with ImgBB api key and http client
type ImgBB struct {
	Key    string
	Client *http.Client
}

// NewImgBB - create a new ImgBB
func NewImgBB(key string, timeout time.Duration) *ImgBB {
	client := &http.Client{
		Timeout: timeout,
	}

	return &ImgBB{
		Key:    key,
		Client: client,
	}
}

// Upload - is a function to upload image to ImgBB
func (i *ImgBB) Upload(img *Image) (*ImgBBResult, *ImgBBError) {
	if img.Size > 33554432 {
		return nil, &ImgBBError{
			StatusCode: http.StatusRequestEntityTooLarge,
			StatusText: http.StatusText(http.StatusRequestEntityTooLarge),
			Err: errInfo{
				Message: ErrFileSize.Error(),
			},
		}
	}

	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		field := "image"
		m.WriteField("key", i.Key)
		m.WriteField("type", "file")
		m.WriteField("action", "upload")
		if len(img.Expiration) > 0 {
			m.WriteField("expiration", img.Expiration)
		}

		part, err := m.CreateFormFile(field, img.Name)
		if err != nil {
			return
		}

		if _, err = io.Copy(part, bytes.NewReader(img.File)); err != nil {
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return nil, &ImgBBError{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			Err: errInfo{
				Message: err.Error(),
			},
		}
	}

	req.Header.Add("Content-Type", m.FormDataContentType())
	req.Header.Add("Host", host)
	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)

	resp, err := i.Client.Do(req)
	if err != nil {
		return nil, &ImgBBError{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			Err: errInfo{
				Message: err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	return i.respParse(resp)
}

func (i *ImgBB) respParse(resp *http.Response) (*ImgBBResult, *ImgBBError) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &ImgBBError{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			Err: errInfo{
				Message: err.Error(),
			},
		}
	}

	if resp.StatusCode == http.StatusOK {
		var res ImgBBResult
		if err := json.Unmarshal(data, &res); err != nil {
			return nil, &ImgBBError{
				StatusCode: http.StatusInternalServerError,
				StatusText: http.StatusText(http.StatusInternalServerError),
				Err: errInfo{
					Message: err.Error(),
				},
			}
		}

		return &res, nil
	}

	var res ImgBBError
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, &ImgBBError{
			StatusCode: http.StatusInternalServerError,
			StatusText: http.StatusText(http.StatusInternalServerError),
			Err: errInfo{
				Message: err.Error(),
			},
		}
	}

	return nil, &res
}
