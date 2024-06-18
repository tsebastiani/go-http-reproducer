package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

var (
	durationTotal        float64
	durationMin          float64
	durationMax          float64
	durationCounter      int
	durationTotalMutex   sync.Mutex
	durationCounterMutex sync.Mutex
	durationMinMaxMutex  sync.Mutex
)

func doRequest(client *http.Client, method string, serverURL string, payload *bytes.Reader) (float64, error) {
	// Start time measurement
	startTime := time.Now()

	// Create a POST request with the payload
	req, err := http.NewRequest(method, serverURL, payload)
	if err != nil {
		return 0.0, err
	}

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	// Read the response body (optional)
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0, err
	}

	// Calculate duration
	duration := time.Since(startTime)

	//log.Printf("Request finished with status %d and took %v", resp.StatusCode, duration)
	return duration.Seconds(), nil
}

// Function to perform the request and update counters
func doRequestWithWaitGroup(tlsConfig *tls.Config, serverURL string, payloadSize int, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create random payload
	payload := make([]byte, payloadSize)
	_, err := rand.Read(payload)
	if err != nil {
		panic(err)
	}

	// Configure transport to enable HTTP/2
	tr := &http2.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	duration, err := doRequest(client, http.MethodPost, serverURL, bytes.NewReader(payload))
	if err != nil {
		panic(err)
	}

	// Update counters within the goroutine
	durationTotalMutex.Lock()
	defer durationTotalMutex.Unlock()
	durationTotal += duration
	durationCounterMutex.Lock()
	defer durationCounterMutex.Unlock()
	durationCounter++
	durationMinMaxMutex.Lock()
	defer durationMinMaxMutex.Unlock()
	if duration <= durationMin || durationMin == 0 {
		durationMin = duration
	}
	if duration >= durationMax {
		durationMax = duration
	}
}

func main() {
	// Server address
	serverURL := "https://localhost:8000"

	// Payload size in bytes
	payloadSize := 10000

	// Number of requests to do
	requestsCount := 15000
	durationMin = 0.0
	durationMax = 0.0
	log.Print(requestsCount)
	// Create a pool with the server certificate since it is not signed
	// by a known CA
	caCert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatalf("Reading server certificate: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration with the certificate of the server
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	var wg sync.WaitGroup

	for i := 1; i <= requestsCount; i++ {
		wg.Add(1)
		go doRequestWithWaitGroup(tlsConfig, serverURL, payloadSize, &wg)
	}

	wg.Wait()

	log.Printf("Min 	Max 	Average")
	log.Printf("%f 		%f 		%f", durationMin, durationMax, durationTotal/float64(durationCounter))

}
