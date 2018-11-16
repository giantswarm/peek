package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

var gitCommit string = "n/a"
var help = flag.Bool("help", false, "Print usage and exit.")
var listenAddr = flag.String("listen", ":8080", "Listen address. Default: :8080")
var next = flag.String("url", "", "URL to call when handling a request.")

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	if flag.Arg(0) == "version" {
		fmt.Printf("peek: %s\n", gitCommit)
		return
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "All good®")
	})

	http.HandleFunc("/callme", func(w http.ResponseWriter, r *http.Request) {
		if next == nil || *next == "" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "You reached the end of the chain.")
			return
		}

		log.Println("Calling next server ", *next)

		resp, err := http.Get(*next)
		if err != nil {
			log.Printf("http.Get(%s) failed: %#q\n", *next, err)
			resp.Body.Close()
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "%#q", err)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(http.StatusOK)
		p := make([]byte, 4096)
		for {
			nRead, err := resp.Body.Read(p)
			if nRead <= 0 && err == io.EOF {
				break
			}

			nWritten, err := w.Write(p[:nRead])
			if err != nil {
				log.Printf("error while writing response: %#q\n", err)
				break
			}

			if nWritten < nRead {
				// Try once.
				nWritten, err = w.Write(p[nWritten:nRead])
				if err != nil {
					log.Printf("error while writing response: %#q\n", err)
					break
				}
			}
		}
	})

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
