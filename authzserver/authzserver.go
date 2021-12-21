package authzserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const (
	redirectPath string = "/auth/callback"
	port         string = "8080"
)

func Serve(userAuthzChan chan struct{}) error {
	router := mux.NewRouter()

	router.HandleFunc(redirectPath, requestTokenAuthz(userAuthzChan)).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not shutdown server: %s", err)
	}

	return nil
}

// RequestTokenAuthz implements a http handler receiving
// an user request to authorize a request token.
func requestTokenAuthz(userAuthzChan chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userAuthzChan <- struct{}{}
		close(userAuthzChan)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Authorization granted. Bye!"))
	}
}
