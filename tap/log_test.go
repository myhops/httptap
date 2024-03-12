package tap

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/myhops/httptap"
)

func TestLogTap(t *testing.T) {
	var (
		responseBodyJSON = []byte(`{
			"from": "response body",
			"name": "Peter Zandbergen"
		}`)
		requestBodyJSON = []byte(`{
			"from": "request body",
			"name": "Peter Zandbergen"
		}`)
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// upstream server
	us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseBodyJSON)
	}))
	defer us.Close()

	pr, err := httptap.New(us.URL, httptap.WithLogger(logger))
	if err != nil {
		t.Fatalf("error creating proxy: %s", err)
	}

	tapGet := NewLogTap(logger, slog.LevelError)

	tapPost := NewLogTap(logger, slog.LevelInfo)

	// add a tap.
	pr.Tap("GET /",
		tapGet,
		httptap.WithRequestBody(),
		httptap.WithResponseBody(),
		httptap.WithRequestJSON(),
		httptap.WithResponseJSON(),
		httptap.WithLogAttrs(slog.String("path", "GET /")),
	)

	pr.Tap("POST /",
		tapPost,
		httptap.WithRequestBody(),
		httptap.WithResponseBody(),
		httptap.WithRequestJSON(),
		httptap.WithResponseJSON(),
		httptap.WithLogAttrs(slog.String("path", "POST /")),
	)

	// proxy server.
	ps := httptest.NewServer(pr)
	// Case 1
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
		if !bytes.Equal(responseBodyJSON, body) {
			t.Fatalf("received incorrect body")
		}
	}
	// Case 2
	{ // issue a request.
		// data := []byte("some text")
		ct := "application/json"
		resp, err := http.Post(ps.URL, ct, bytes.NewReader(requestBodyJSON))
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
		if !bytes.Equal(responseBodyJSON, body) {
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
		if !bytes.Equal(responseBodyJSON, body) {
			t.Fatalf("received incorrect body")
		}
	}
}
