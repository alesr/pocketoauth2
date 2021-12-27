#POCKETOAUTH2

Implements a simple authentication client towards the Pocket API, including a http server for receiving callbacks for user request token authorization.


```go
func ExampleService() {
	authService, err := New("https://getpocket.com/v3", "consumerKey", "http://localhost:8080/auth/callback")
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
