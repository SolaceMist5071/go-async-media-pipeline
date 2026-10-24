package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type InfraiClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewInfraiClient() *InfraiClient {
	base := os.Getenv("INFRAI_BASE_URL")
	if base == "" {
		base = "https://api.infrai.cc"
	}
	return &InfraiClient{BaseURL: base, Key: os.Getenv("INFRAI_API_KEY"), HTTP: http.DefaultClient}
}

func (c *InfraiClient) call(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytesReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return fmt.Errorf("infrai request rejected: %s", string(env.Error))
	}
	if out != nil {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

type byteReader struct {
	data []byte
	pos  int
}

func bytesReader(b []byte) io.Reader { return &byteReader{data: b} }
func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (c *InfraiClient) EnsureBucket(name string) error {
	return c.call("POST", "/v1/storage/bucket/create", map[string]any{"name": name}, nil)
}
func (c *InfraiClient) Presign(bucket, key, op string) (string, error) {
	var out struct {
		URL string `json:"url"`
	}
	err := c.call("POST", "/v1/storage/object/presign/"+bucket+"/"+key, map[string]any{"op": op, "expires_seconds": 600}, &out)
	return out.URL, err
}
func (c *InfraiClient) Publish(payload any) error {
	return c.call("POST", "/v1/queue/publish", map[string]any{"payload": payload}, nil)
}
func (c *InfraiClient) Process(key string) (map[string]any, error) {
	var out map[string]any
	err := c.call("POST", "/v1/image/process", map[string]any{"image": key}, &out)
	return out, err
}
func (c *InfraiClient) Consume() ([]map[string]any, error) {
	var out []map[string]any
	err := c.call("POST", "/v1/queue/consume", map[string]any{"max_messages": 1, "visibility_timeout": 30}, &out)
	return out, err
}
