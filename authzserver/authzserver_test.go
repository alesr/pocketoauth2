package authzserver

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServe(t *testing.T) {
	ch := make(chan struct{}, 1)

	go func(ch chan struct{}) {
		_ = Serve(ch)
	}(ch)

	// prepare the request
	req, err := http.NewRequest("GET", "http://localhost:8080/auth/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	// send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Check the channel.
	<-ch

	// Check the response.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestRequestTokenAuthz(t *testing.T) {
	ch := make(chan struct{}, 1)

	handler := requestTokenAuthz(ch)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/auth/callback")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	<-ch

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code to be '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

	// check response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "Authorization granted. Bye!" {
		t.Errorf("expected response body to be '%s', got '%s'", "response", string(body))
	}
}
