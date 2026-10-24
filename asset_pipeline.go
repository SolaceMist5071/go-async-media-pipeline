package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Asset struct {
	ID        string `json:"id"`
	ObjectKey string `json:"object_key"`
	State     string `json:"state"`
}

type Pipeline struct {
	Client *InfraiClient
	Bucket string
}

func assetKey(id, filename string) string {
	return strings.TrimPrefix(id, "/") + "/" + strings.TrimPrefix(filename, "/")
}

func (p *Pipeline) Upload(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID       string `json:"id"`
		Filename string `json:"filename"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.ID == "" || in.Filename == "" {
		http.Error(w, "invalid asset", 400)
		return
	}
	key := assetKey(in.ID, in.Filename)
	url, err := p.Client.Presign(p.Bucket, key, "put")
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	asset := Asset{ID: in.ID, ObjectKey: key, State: "upload-url-issued"}
	if err = p.Client.Publish(asset); err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"asset": asset, "upload_url": url, "next": "PUT bytes to upload_url, then the worker processes the queued asset"})
}

func (p *Pipeline) Download(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/assets/")
	if id == "" {
		http.Error(w, "missing asset", 400)
		return
	}
	url, err := p.Client.Presign(p.Bucket, id, "get")
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"download_url": url})
}

func (p *Pipeline) WorkerOnce() error {
	msgs, err := p.Client.Consume()
	if err != nil || len(msgs) == 0 {
		return err
	}
	var a Asset
	raw, _ := json.Marshal(msgs[0]["payload"])
	if json.Unmarshal(raw, &a) != nil {
		return nil
	}
	_, err = p.Client.Process(a.ObjectKey)
	return err
}
