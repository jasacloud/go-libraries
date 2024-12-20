package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	srv    *http.Server
	srvTLS *http.Server

	onShutdownFunc func()
)

// runAutoCert support 1-line LetsEncrypt HTTPS servers
func runAutoCert(r http.Handler, domain ...string) error {
	return http.Serve(autocert.NewListener(domain...), r)
}

// runWithManager support custom autocert manager
func runWithManager(r http.Handler, m *autocert.Manager) {
	srvTLS = &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
		Handler:   r,
	}

	go func() {
		if err := http.ListenAndServe(":http", m.HTTPHandler(http.HandlerFunc(redirect))); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()
	if err := srvTLS.ListenAndServeTLS("", ""); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server listen failed: %v", err)
	}
}

// redirect function
func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

// runTLS function
func runTLS(addr string, engine http.Handler, certFile, keyFile string) {
	debugPrint("Listening and serving HTTPS on %s\n", addr)

	srvTLS = &http.Server{Addr: addr, Handler: engine}
	if onShutdownFunc != nil {
		srv.RegisterOnShutdown(onShutdownFunc)
	}
	if err := srvTLS.ListenAndServeTLS(certFile, keyFile); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server listen failed: %v", err)
	}
}

// run function
func run(engine http.Handler, addr ...string) {
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)

	srv = &http.Server{Addr: address, Handler: engine}
	if onShutdownFunc != nil {
		srv.RegisterOnShutdown(onShutdownFunc)
	}
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server listen failed: %v", err)
	}
}

func onShutdown(f func()) {
	onShutdownFunc = f
}

func shutdown() {
	if srv != nil {
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.Printf("The service is shutting down with error: %v", err)
		} else {
			log.Print("The service is shutting down...")
		}
	}
	if srvTLS != nil {
		if err := srvTLS.Shutdown(context.TODO()); err != nil {
			log.Printf("The service is shutting down with error: %v", err)
		} else {
			log.Print("The service is shutting down...")
		}
	}
}

// debugPrint function
func debugPrint(format string, values ...interface{}) {
	if gin.IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(gin.DefaultWriter, "[GIN-debug] "+format, values...)
	}
}

// debugPrintError function
func debugPrintError(err error) {
	if err != nil {
		if gin.IsDebugging() {
			fmt.Fprintf(gin.DefaultErrorWriter, "[GIN-debug] [ERROR] %v\n", err)
		}
	}
}

// resolveAddress function
func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
