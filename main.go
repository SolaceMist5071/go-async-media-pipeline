package main

import (
	"log"
	"net/http"
)

func main() {
	c := NewInfraiClient()
	bucket := "media-pipeline-assets"
	p := &Pipeline{Client: c, Bucket: bucket}
	http.HandleFunc("/assets", p.Upload)
	http.HandleFunc("/assets/", p.Download)
	log.Println("media pipeline listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
