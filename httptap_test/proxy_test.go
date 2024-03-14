package httptap_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/myhops/httptap"
	"github.com/myhops/httptap/tap"
)

func TestProxy(t *testing.T) {
	var responseBody = []byte("dasfadfaefacvv vasdfad asdfasd")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// upstream server
	us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(responseBody)
	}))
	defer us.Close()

	pr, err := httptap.New(us.URL, httptap.WithLogger(logger))
	if err != nil {
		t.Fatalf("error creating proxy: %s", err)
	}

	tapGet := httptap.TapFunc(func(_ context.Context, rr *httptap.RequestResponse) {
		t.Logf("tap GET called")
	})

	tapPost := httptap.TapFunc(func(_ context.Context, rr *httptap.RequestResponse) {
		t.Logf("tap PUT called")
	})

	// add a tap.
	pr.Tap([]string{"GET /"},
		tapGet,
		httptap.WithRequestBody(),
		httptap.WithResponseBody(false),
		httptap.WithLogAttrs(slog.String("path", "GET /")),
	)

	pr.Tap([]string{"POST /"},
		tapPost,
		httptap.WithRequestBody(),
		httptap.WithResponseBody(false),
		httptap.WithLogAttrs(slog.String("path", "POST /")),
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

func BenchmarkLogTap(b *testing.B) {

	var responseBody = []byte("dasfadfaefacvv vasdfad asdfasd")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// upstream server
	us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(responseBody)
	}))
	defer us.Close()

	pr, err := httptap.New(us.URL, httptap.WithLogger(logger))
	if err != nil {
		b.Fatalf("error creating proxy: %s", err)
	}

	tapAll := tap.NewLogTap(logger, slog.LevelInfo)

	pr.Tap([]string{"/"},
		tapAll,
		httptap.WithRequestBody(),
		httptap.WithResponseBody(),
		httptap.WithRequestJSON(),
		httptap.WithResponseJSON(),
		httptap.WithLogAttrs(slog.String("path", "POST /")),
	)

	// proxy server.
	ps := httptest.NewServer(pr)

	testFunc := func() {
		data := []byte("some text")
		ct := http.DetectContentType(data)
		req, _ := http.NewRequest(http.MethodPost, ps.URL, bytes.NewReader([]byte("Hallo put")))
		req.Header.Set("content-type", ct)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("get error: %s", err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b.Fatalf("received bad status code: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			b.Fatalf("error reading body")
		}
		if !bytes.Equal(responseBody, body) {
			b.Fatalf("received incorrect body")
		}
	}

	for i := 0; i < b.N; i++ {
		testFunc()
	}

}
