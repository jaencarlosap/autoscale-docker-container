package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	targetContainer = os.Getenv("TARGET_CONTAINER")
	targetHost      = os.Getenv("TARGET_HOST")
	targetPort      = os.Getenv("TARGET_PORT")
	dockerAPI        = os.Getenv("DOCKER_API")

	idleTimeout, _   = strconv.Atoi(getEnv("IDLE_TIMEOUT_SECONDS", "600"))
	checkInterval, _ = strconv.Atoi(getEnv("CHECK_INTERVAL_SECONDS", "30"))

	lastRequest time.Time
	activeReqs  int
	mu          sync.Mutex
	starting    bool
)

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func startContainer() {
	mu.Lock()
	if starting {
		mu.Unlock()
		return
	}
	starting = true
	mu.Unlock()

	req, _ := http.NewRequest("POST",
		dockerAPI+"/containers/"+targetContainer+"/start", nil)

	http.DefaultClient.Do(req)

	mu.Lock()
	starting = false
	mu.Unlock()
}

func stopContainer() {
	req, _ := http.NewRequest("POST",
		dockerAPI+"/containers/"+targetContainer+"/stop", nil)

	http.DefaultClient.Do(req)
}

func waitForDNS() {
	for i := 0; i < 20; i++ {
		_, err := net.LookupHost(targetHost)
		if err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	if targetPort == "" {
		targetPort = "80"
	}

	targetURL, _ := url.Parse("http://" + targetHost + ":" + targetPort)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	go func() {
		for {
			time.Sleep(time.Duration(checkInterval) * time.Second)

			mu.Lock()
			if activeReqs == 0 && !lastRequest.IsZero() {
				if time.Since(lastRequest) > time.Duration(idleTimeout)*time.Second {
					log.Println("Idle timeout reached → stopping container")
					stopContainer()
					lastRequest = time.Time{}
				}
			}
			mu.Unlock()
		}
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		activeReqs++
		lastRequest = time.Now()
		mu.Unlock()

		startContainer()
		waitForDNS()

		proxy.ServeHTTP(w, r)

		mu.Lock()
		activeReqs--
		mu.Unlock()
	})

	log.Println("Go starter listening on :8080")
	http.ListenAndServe(":8080", handler)
}
