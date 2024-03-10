package httptap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestProxy(t *testing.T) {
	var responseBody = []byte("dasfadfaefacvv vasdfad asdfasd")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// upstream server
	us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(responseBody)
	}))
	defer us.Close()

	pr, err := New(us.URL, WithLogger(logger))
	if err != nil {
		t.Fatalf("error creating proxy: %s", err)
	}

	tapGet := TapFunc(func(_ context.Context, rr *RequestResponse) {
		t.Logf("tap GET called")
	})

	tapPost := TapFunc(func(_ context.Context, rr *RequestResponse) {
		t.Logf("tap PUT called")
	})

	// add a tap.
	pr.Tap("GET /",
		tapGet,
		WithRequestBody(),
		WithResponseBody(false),
		WithLogAttrs(slog.String("path", "GET /")),
	)

	pr.Tap("POST /",
		tapPost,
		WithRequestBody(),
		WithResponseBody(false),
		WithLogAttrs(slog.String("path", "POST /")),
	)

	// proxy server.
	ps := httptest.NewServer(pr)
	{
		// issue a request.
		resp, err := http.Get(ps.URL)
		if err != nil {
			t.Fatalf("get error: %s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Fatalf("received bad status code: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading body")
		}
		if !bytes.Equal(responseBody, body) {
			t.Fatalf("received incorrect body")
		}
	}
	{ // issue a request.
		data := []byte("some text")
		ct := http.DetectContentType(data)
		resp, err := http.Post(ps.URL, ct, bytes.NewReader(data))
		if err != nil {
			t.Fatalf("get error: %s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Fatalf("received bad status code: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading body")
		}
		if !bytes.Equal(responseBody, body) {
			t.Fatalf("received incorrect body")
		}
	}
	{ // issue a request.
		data := []byte("some text")
		ct := http.DetectContentType(data)
		req, _ := http.NewRequest(http.MethodPut, ps.URL, bytes.NewReader([]byte("Hallo put")))
		req.Header.Set("content-type", ct)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("get error: %s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Fatalf("received bad status code: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading body")
		}
		if !bytes.Equal(responseBody, body) {
			t.Fatalf("received incorrect body")
		}
	}
}
