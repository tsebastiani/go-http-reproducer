package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"golang.org/x/net/http2"
)

var cpuProfileFile *os.File
var memProfileFile *os.File

func main() {
	// Create a server on port 8000
	// Exactly how you would run an HTTP/1.1 server

	mux := http.NewServeMux()
	mux.HandleFunc("/start-cpu", startCPUProfile)
	mux.HandleFunc("/stop-cpu", stopCPUProfile)
	mux.HandleFunc("/", handler)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	http2.ConfigureServer(srv, &http2.Server{MaxReadFrameSize: 16000})

	// Start the server with TLS, since we are running HTTP/2 it must be
	// run with TLS.
	// Exactly how you would run an HTTP/1.1 server with TLS connection.
	log.Printf("Serving on https://0.0.0.0:8000")
	log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
}

func startCPUProfile(w http.ResponseWriter, r *http.Request) {
	var err error
	cpuProfileFile, err = os.Create("cpu.prof")
	if err != nil {
		http.Error(w, "could not create CPU profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		http.Error(w, "could not start CPU profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "CPU profiling started")
}

func stopCPUProfile(w http.ResponseWriter, r *http.Request) {
	pprof.StopCPUProfile()
	cpuProfileFile.Close()
	fmt.Fprintln(w, "CPU profiling stopped")
}

func handler(w http.ResponseWriter, r *http.Request) {
	count := 0
	readStartTime := time.Now()
	defer func() {
		readEndTime := time.Now()
		duration := readEndTime.Sub(readStartTime)
		if duration.Seconds() > 10 {
			fmt.Println(fmt.Sprintf("GGMGGM17 HandleStream read loop count %d time %s ts %d.%09d", count, duration.String(), readEndTime.Unix(), readEndTime.Nanosecond()))
		}
	}()

	// TODO: minimize garbage, optimize recvBuffer code/ownership
	const readSize = 8196
	//	var bodyBuffer bytes.Buffer
	for buf := make([]byte, readSize); ; {
		count++
		httpReadStart := time.Now()
		n, err := r.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			} else {
				panic(fmt.Errorf("error while reading body %s", err.Error()))
			}
		}
		httpReadEnd := time.Now()
		httpReadDuration := httpReadEnd.Sub(httpReadStart)
		if httpReadDuration.Seconds() > 3 {
			fmt.Println(fmt.Sprintf("GGMGGM18 HandleStream read loop single iter count %d time %s ts %d.%09d req obj %#v", count, httpReadDuration.String(), httpReadEnd.Unix(), httpReadEnd.Nanosecond(), r))
		}
		if n > 0 {
			buf = buf[n:]
			//_, err := bodyBuffer.Write(buf)

			if err != nil {
				panic(err)
			}
		}

		allocBufStart := time.Now()
		if len(buf) == 0 {
			buf = make([]byte, readSize)
		}
		allocBufDuration := time.Now().Sub(allocBufStart)
		if allocBufDuration.Seconds() > 3 {
			fmt.Println(fmt.Sprintf("GGMGGM18 HandleStream read loop make buf single iter count %d time %s ts %d.%09d", count, httpReadDuration.String()))
		}
	}

}
