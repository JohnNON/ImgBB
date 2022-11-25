package imgbb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testImg = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0}

func Test_Upload_Success(t *testing.T) {
	resp := `{
		"data": {
			"id": "2ndCYJK",
			"title": "c1f64245afb2",
			"url_viewer": "https://ibb.co/2ndCYJK",
			"url": "https://i.ibb.co/w04Prt6/c1f64245afb2.gif",
			"display_url": "https://i.ibb.co/98W13PY/c1f64245afb2.gif",
			"width": 1,
			"height": 1,
			"size": 42,
			"time": "1552042565",
			"expiration":"0",
			"image": {
				"filename": "c1f64245afb2.gif",
				"name": "c1f64245afb2",
				"mime": "image/gif",
				"extension": "gif",
				"url": "https://i.ibb.co/w04Prt6/c1f64245afb2.gif"
			},
			"thumb": {
				"filename": "c1f64245afb2.gif",
				"name": "c1f64245afb2",
				"mime": "image/gif",
				"extension": "gif",
				"url": "https://i.ibb.co/2ndCYJK/c1f64245afb2.gif"
			},
			"medium": {
				"filename": "c1f64245afb2.gif",
				"name": "c1f64245afb2",
				"mime": "image/gif",
				"extension": "gif",
				"url": "https://i.ibb.co/98W13PY/c1f64245afb2.gif"
			},
			"delete_url": "https://ibb.co/2ndCYJK/670a7e48ddcb85ac340c717a41047e5c"
		},
		"success": true,
		"status": 200
	}`

	img := NewImage("name", "", testImg)

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				fmt.Fprintln(w, resp)
			},
		),
	)
	defer ts.Close()

	apiClient := New(*ts.Client(), "secret-key", WithEndpoint(ts.URL))

	expect := &ImgBBResponse{
		Data: Data{
			ID:         "2ndCYJK",
			Title:      "c1f64245afb2",
			UrlViewer:  "https://ibb.co/2ndCYJK",
			Url:        "https://i.ibb.co/w04Prt6/c1f64245afb2.gif",
			DisplayUrl: "https://i.ibb.co/98W13PY/c1f64245afb2.gif",
			Width:      1,
			Height:     1,
			Size:       42,
			Time:       "1552042565",
			Expiration: "0",
			Image: Info{
				Filename:  "c1f64245afb2.gif",
				Name:      "c1f64245afb2",
				Mime:      "image/gif",
				Extension: "gif",
				Url:       "https://i.ibb.co/w04Prt6/c1f64245afb2.gif",
			},
			Thumb: Info{
				Filename:  "c1f64245afb2.gif",
				Name:      "c1f64245afb2",
				Mime:      "image/gif",
				Extension: "gif",
				Url:       "https://i.ibb.co/2ndCYJK/c1f64245afb2.gif",
			},
			Medium: Info{
				Filename:  "c1f64245afb2.gif",
				Name:      "c1f64245afb2",
				Mime:      "image/gif",
				Extension: "gif",
				Url:       "https://i.ibb.co/98W13PY/c1f64245afb2.gif",
			},
			DeleteUrl: "https://ibb.co/2ndCYJK/670a7e48ddcb85ac340c717a41047e5c",
		},
		Success:    true,
		StatusCode: http.StatusOK,
	}

	actual, err := apiClient.Upload(img)

	assert.NoError(t, err)
	assert.Equal(t, expect, actual)
}

func Test_Upload_ImgBBError(t *testing.T) {
	resp := `{
		"error": {
			"message": "error message",
			"code": 999,
			"context": "error context"
		},
		"status_code": 500,
		"status_txt": "internal error"
	}`

	img := NewImage("name", "", testImg)

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				fmt.Fprintln(w, resp)
			},
		),
	)
	defer ts.Close()

	apiClient := New(*ts.Client(), "secret-key", WithEndpoint(ts.URL))

	expect := ImgBBError{
		StatusCode: http.StatusInternalServerError,
		StatusText: "internal error",
		Err: ErrInfo{
			Code:    999,
			Message: "error message",
			Context: "error context",
		},
	}

	_, err := apiClient.Upload(img)

	assert.Equal(t, expect, err)
}

func Test_Upload_ClientInternalServerError(t *testing.T) {
	resp := `bad response format`

	img := NewImage("name", "", testImg)

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				fmt.Fprintln(w, resp)
			},
		),
	)
	defer ts.Close()

	apiClient := New(*ts.Client(), "secret-key", WithEndpoint(ts.URL))

	_, err := apiClient.Upload(img)

	assert.ErrorIs(t, err, ImgBBError{
		StatusCode: http.StatusInternalServerError,
		StatusText: http.StatusText(http.StatusInternalServerError),
	})
}

func Test_Upload_EmptyImage(t *testing.T) {
	img := NewImage("name", "", []byte{})

	apiClient := New(http.Client{}, "secret-key")

	_, err := apiClient.Upload(img)

	assert.ErrorIs(t, err, ImgBBError{
		StatusCode: http.StatusBadRequest,
		StatusText: http.StatusText(http.StatusBadRequest),
	})
}

func Test_Upload_OversizeImage(t *testing.T) {
	img := &Image{
		name:       "name",
		size:       len(testImg) * 10000000,
		expiration: "",
		file:       testImg,
	}

	apiClient := New(http.Client{}, "secret-key")

	_, err := apiClient.Upload(img)

	assert.ErrorIs(t, err, ImgBBError{
		StatusCode: http.StatusBadRequest,
		StatusText: http.StatusText(http.StatusBadRequest),
	})
}

func Test_Upload_ErrorUnmarshalFail(t *testing.T) {
	resp := `bad error format`

	img := NewImage("name", "", testImg)

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				fmt.Fprintln(w, resp)
			},
		),
	)
	defer ts.Close()

	apiClient := New(*ts.Client(), "secret-key", WithEndpoint(ts.URL))

	_, err := apiClient.Upload(img)

	assert.ErrorIs(t, err, ImgBBError{
		StatusCode: http.StatusInternalServerError,
		StatusText: http.StatusText(http.StatusInternalServerError),
	})
}
