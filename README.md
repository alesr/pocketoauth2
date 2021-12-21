#POCKETOAUTH2

Implements a simple authentication client towards the Pocket API, including a http server for receiving callbacks for user request token authorization.


```go
func ExampleService() {

    // Gets notified when user approves the user authorizes the request token.
	userAuthzChan := make(chan struct{},1)

    // Starts an HTTP server serving the redirect URI.
    // Runs by default on port 8080.
	go func(ch chan struct{}) {
		if err := authzserver.Serve(userAuthzChan); err != nil {
			log.Fatalln(err)
		}
	}(userAuthzChan)


	authService, err := New("https://getpocket.com/v3", "consumerKey", "http://localhost:8080/auth/callback", userAuthzChan)
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
```