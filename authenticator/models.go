package authenticator

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}

	accessTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	accessTokenResponse struct {
		Val      string `json:"access_token"`
		Username string `json:"username"`
	}
)
