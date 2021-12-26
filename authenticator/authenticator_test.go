package authenticator

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ExampleService() {
	authService, err := New("https://getpocket.com/v3", "consumerKey", "http://localhost:8081/auth/callback")
	if err != nil {
		log.Fatalln(err)
	}

	accessTkn, username, err := authService.Authenticate(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Access token:", accessTkn)
	fmt.Println("Username:", username)
}

func TestNew_validation(t *testing.T) {
	cases := []struct {
		name          string
		givenArgs     []string
		givenChan     chan struct{}
		expectedError error
	}{
		{
			name: "missing host",
			givenArgs: []string{
				"",
				"consumerKey",
				"redirectURI",
			},
			givenChan:     make(chan struct{}),
			expectedError: ErrMissingHost,
		},
		{
			name: "missing consumer key",
			givenArgs: []string{
				"host",
				"",
				"redirectURI",
			},
			givenChan:     make(chan struct{}),
			expectedError: ErrMissingConsumerKey,
		},
		{
			name: "missing redirect URI",
			givenArgs: []string{
				"host",
				"consumerKey",
				"",
			},
			givenChan:     make(chan struct{}),
			expectedError: ErrMissingRedirectURI,
		},
		{
			name: "valid",
			givenArgs: []string{
				"host",
				"consumerKey",
				"redirectURI",
			},
			givenChan:     make(chan struct{}),
			expectedError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.givenArgs[0], tc.givenArgs[1], tc.givenArgs[2])
			if err != tc.expectedError {
				t.Errorf("expected error %s, got %s", tc.expectedError, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Test that the service is created with fields populated correctly.
	s, err := New("host", "consumerKey", "redirectURI")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if s.host != "host" {
		t.Errorf("expected host to be 'host', got '%s'", s.host)
	}

	if s.consumerKey != "consumerKey" {
		t.Errorf("expected consumerKey to be 'consumerKey', got '%s'", s.consumerKey)
	}

	if s.redirectURI != "redirectURI" {
		t.Errorf("expected redirectURI to be 'redirectURI', got '%s'", s.redirectURI)
	}

	if s.httpCli == nil {
		t.Error("expected httpCli to be non-nil")
	}

	if s.accessToken != "" {
		t.Errorf("expected accessToken to be empty, got '%s'", s.accessToken)
	}

	if s.username != "" {
		t.Errorf("expected username to be empty, got '%s'", s.username)
	}
}

func TestMakeRequest(t *testing.T) {
	// Arrange

	givenRequestMethod := http.MethodGet

	expectedContentType := "application/json; charset=UTF8"
	expectedStatusCode := http.StatusOK
	expectedResponse := "response"

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Assert

			if r.Method != givenRequestMethod {
				t.Errorf("expected request method to be '%s', got '%s'", givenRequestMethod, r.Method)
			}

			if r.Header.Get("Content-Type") != expectedContentType {
				t.Errorf("expected Content-Type to be '%s', got '%s'", expectedContentType, r.Header.Get("Content-Type"))
			}

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("response")); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		}),
	)
	defer ts.Close()

	givenRequest, err := http.NewRequest(givenRequestMethod, ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	s := &Service{}

	// Act

	s.host = ts.URL
	s.httpCli = http.DefaultClient

	observedResp, err := s.makeRequest(context.Background(), givenRequest, expectedStatusCode)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Assert

	if observedResp.StatusCode != expectedStatusCode {
		t.Errorf("expected status code %d, got %d", expectedStatusCode, observedResp.StatusCode)
	}

	observedBody, err := ioutil.ReadAll(observedResp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(observedBody) != expectedResponse {
		t.Errorf("expected response body to be '%s', got '%s'", expectedResponse, observedBody)
	}
}
