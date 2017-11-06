package httpcaching

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type Client struct {
	http.Client
	CacheDir string
	Flush    bool
}

type ReadCloser struct {
	*bytes.Reader
	body io.ReadCloser
}

func (rc ReadCloser) Close() error {
	return nil
}

func (rc *ReadCloser) UnmarshalBinary(b []byte) error {
	rc.Reader = bytes.NewReader(b)
	return nil
}

func (rc *ReadCloser) MarshalBinary() ([]byte, error) {
	b, err := ioutil.ReadAll(rc.body)
	rc.body.Close()
	rc.Reader = bytes.NewReader(b)
	return b, err
}

func init() {
	gob.Register(ReadCloser{})
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.CacheDir == "" || req.Method != "GET" {
		return c.Client.Do(req)
	}
	sum := sha1.Sum([]byte(req.URL.String()))
	hash := hex.EncodeToString(sum[:])
	dir := path.Join(c.CacheDir, hash[:2])
	filename := path.Join(dir, hash)
	if !c.Flush {
		if file, err := os.Open(filename); err == nil {
			resp := new(http.Response)
			err := gob.NewDecoder(file).Decode(resp)
			file.Close()
			return resp, err
		}
	}
	resp, err := c.Client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		return resp, err
	}
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return resp, err
		}
	}
	file, err := os.Create(filename + "~")
	defer file.Close()
	if err != nil {
		return resp, err
	}
	resp.TLS = nil
	resp.Body = &ReadCloser{body: resp.Body}
	if err := gob.NewEncoder(file).Encode(resp); err != nil {
		return resp, err
	}
	return resp, os.Rename(filename+"~", filename)
}
