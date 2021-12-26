package authenticator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alesr/callbacksrv"
	"github.com/alesr/pocketoauth2/httputil"
)

const (
	authURLTemplate      string = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"
	endpointAuthorize    string = "/oauth/authorize"
	endpointRequestToken string = "/oauth/request"

	xErrorHeader string = "X-Error"

	// 1 minute since we need to wait for the user to authorize request token in the browser
	defaultAuthCtxTimeout time.Duration = time.Minute
)

var (
	_ Authenticator = (*Service)(nil)

	// Enumerate the possible errors that can be returned by the authenticator.

	ErrMissingConsumerKey = errors.New("missing consumer key")
	ErrMissingHost        = errors.New("missing host")
	ErrMissingRedirectURI = errors.New("missing redirect URI")
)

type Authenticator interface {
	Authenticate(ctx context.Context) (string, string, error)
}

type Service struct {
	// Dependencies

	httpCli       *http.Client
	userAuthzChan chan struct{}

	// Configuration

	host        string
	redirectURI string

	// Credentials

	consumerKey string
	accessToken string
	username    string
}

func New(host string, consumerKey string, redirectURI string) (*Service, error) {
	// Input Validation

	if host == "" {
		return nil, ErrMissingHost
	}

	if consumerKey == "" {
		return nil, ErrMissingConsumerKey
	}

	if redirectURI == "" {
		return nil, ErrMissingRedirectURI
	}

	return &Service{
		httpCli:     httputil.BaseClient(),
		host:        host,
		consumerKey: consumerKey,
		redirectURI: redirectURI,
	}, nil

}

func (s *Service) Authenticate(ctx context.Context) (string, string, error) {
	if s.accessToken != "" && s.username != "" {
		return s.accessToken, s.username, nil
	}

	authzCtx, cancel := context.WithTimeout(ctx, defaultAuthCtxTimeout)
	defer cancel()

	requestToken, err := s.obtainRequestToken(authzCtx)
	if err != nil {
		return "", "", err
	}

	authzURL := fmt.Sprintf(authURLTemplate, requestToken, s.redirectURI)

	userAuthzChan := make(chan struct{}, 1)
	doneChan := make(chan os.Signal, 1)

	callbacksrv.Serve(userAuthzChan, doneChan)

	fmt.Printf("\n\nAwaiting user authorization:\n\t%s\n\n", authzURL)

	<-s.userAuthzChan

	fmt.Print("Authorization granted!\n\n")

	accessToken, err := s.obtainAccessToken(authzCtx, requestToken)
	if err != nil {
		return "", "", err
	}

	s.accessToken = accessToken.Val
	s.username = accessToken.Username

	return accessToken.Val, accessToken.Username, nil
}

func (s *Service) ClearCredentials() {
	s.accessToken = ""
	s.username = ""
}

// obtainRequestToken obtains a request token from Pocket.
func (s *Service) obtainRequestToken(ctx context.Context) (string, error) {
	reqTokenInput := requestTokenRequest{
		ConsumerKey: s.consumerKey,
		RedirectURI: s.redirectURI,
	}

	b, err := json.Marshal(reqTokenInput)
	if err != nil {
		return "", fmt.Errorf("could not marshal request body: %s", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.host+endpointRequestToken, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("could not create request: %s", err)
	}

	resp, err := s.makeRequest(ctx, req, http.StatusOK)
	if err != nil {
		return "", fmt.Errorf("could not make request for request token: %s", err)
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %s", err)
	}
	defer resp.Body.Close()

	values, err := url.ParseQuery(string(payload))
	if err != nil {
		return "", fmt.Errorf("could not parse response body: %s", err)
	}

	requestToken := values.Get("code")
	if requestToken == "" {
		return "", errors.New("missing request token code in api response")
	}
	return requestToken, nil
}

//	obtainAccessToken obtains an access token from Pocket given a request token.
func (s *Service) obtainAccessToken(ctx context.Context, requestToken string) (*accessTokenResponse, error) {
	accessTokenReq := accessTokenRequest{
		ConsumerKey: s.consumerKey,
		Code:        requestToken,
	}

	b, err := json.Marshal(accessTokenReq)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request body: %s", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.host+endpointAuthorize, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}

	resp, err := s.makeRequest(ctx, req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("could not make request for access token: %s", err)
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %s", err)
	}
	defer resp.Body.Close()

	values, err := url.ParseQuery(string(payload))
	if err != nil {
		return nil, fmt.Errorf("could not parse response body: %s", err)
	}

	accessToken, username := values.Get("access_token"), values.Get("username")
	if accessToken == "" {
		return nil, errors.New("empty access token in API response")
	}

	return &accessTokenResponse{
		Val:      accessToken,
		Username: username,
	}, nil
}

func (s *Service) makeRequest(ctx context.Context, req *http.Request, expectedStatusCode int) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json; charset=UTF8")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %s", err)
	}

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf(
			"unexpected status code '%s': %s",
			http.StatusText(resp.StatusCode),
			resp.Header.Get(xErrorHeader),
		)
	}
	return resp, nil
}
