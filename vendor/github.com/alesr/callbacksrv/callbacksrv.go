package callbacksrv

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Serve implements a http server prepared to serve a single callback request on demand.
// When the callback request is received, the server will send a notification to the notifyChan channel,
// close the channel sinalizing that no more requests will be served, and shuts down itself gracefully.
func Serve(notifyChan chan struct{}, quitChan chan os.Signal) {
	router := gin.Default()

	router.GET("/callback", func(c *gin.Context) {
		c.String(http.StatusOK, "Authorization Granted. Bye!")

		notifyChan <- struct{}{}
		close(notifyChan)

		quitChan <- os.Interrupt
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		<-quitChan
		if err := server.Close(); err != nil {
			log.Fatal("server Close:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Println("server closed under request")
		} else {
			log.Fatal("server closed unexpect")
		}
	}
}
