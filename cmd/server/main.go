package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Go Auth Service is running!")
    })

    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", nil)
}