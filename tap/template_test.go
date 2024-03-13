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

func TestTemplateTap(t *testing.T) {
	tpl, err := NewTemplateTap(slog.Default(), DefaultTemplate, "", "json")
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	_ = tpl
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

	tapGet := tpl

	tapPost := tpl

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
		httptap.WithResponseBody(false),
		httptap.WithRequestJSON(),
		httptap.WithResponseJSON(),
		httptap.WithLogAttrs(slog.String("path", "POST /")),
	)

	// proxy server.
	ps := httptest.NewServer(pr)
	{
		// issue a request.
		resp, err := http.Get(ps.URL+"/hallo")
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

	// t.Error()
}


const (
	fromBodyTemplate = `{"time":"{{ .Data.Start.Format "2006-01-02T15:04:05.999999999Z07:00" }}","body_from":"{{ .Data.RespBodyJSON.from}}"}{{"\n"}}`
)

func TestTemplateTapBody(t *testing.T) {
	tpl, err := NewTemplateTap(slog.Default(), fromBodyTemplate, "data", "json")
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	_ = tpl
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

	tapGet := tpl

	tapPost := tpl

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
	{
		// issue a request.
		resp, err := http.Get(ps.URL+"/hallo")
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

	// t.Error()
}
