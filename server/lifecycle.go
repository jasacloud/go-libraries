package server

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type svcLifecycle struct {
	isReady             *atomic.Value
	readyzProbeDuration time.Duration
}

func NewSvcLifecycle() *svcLifecycle {
	lc := &svcLifecycle{
		isReady: &atomic.Value{},
	}

	return lc
}

// SetDuration is set custom duration for ready probe lifecycle
func (slc *svcLifecycle) SetDuration(d time.Duration) *svcLifecycle {
	slc.readyzProbeDuration = d
	return slc
}

// Init is run lifecycle runtime
func (slc *svcLifecycle) Init() *svcLifecycle {
	slc.isReady.Store(false)
	go func() {
		log.Printf("Readyz probe is negative by default...")
		time.Sleep(slc.readyzProbeDuration)
		slc.isReady.Store(true)
		log.Printf("Readyz probe is positive.")
	}()
	return slc
}

// Healthz is a liveness probe.
func (slc *svcLifecycle) Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Readyz is a readiness probe.
func (slc *svcLifecycle) Readyz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if slc.isReady == nil || !slc.isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// LoadDefaultLifecycle is enable lifecycle liveness and readiness probe
func LoadDefaultLifecycle() {
	if Route != nil {
		lifecycle := NewSvcLifecycle().SetDuration(5 * time.Second).Init()
		Route.GET("/healthz", gin.WrapF(lifecycle.Healthz))
		Route.GET("/readyz", gin.WrapH(lifecycle.Readyz()))
		return
	}
	log.Fatal("server route not loaded, please init load server first")
}
