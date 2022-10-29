// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("HelloServer starting")
	http.HandleFunc("/", HelloServer)
	// TLS handled by Load Balancer
	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	// output is text/plain and not html
	// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter
	fmt.Fprintf(w, "Hello, %s!", path)
	if path != "healthcheck" {
		log.Printf("host = %s, path = %s\n", r.Host, path)
	}
}
