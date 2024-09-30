// Package pkg provides utilities for compressing and decompressing HTTP responses and requests.
// It includes a gzip-based compression writer and reader, which can transparently handle
// gzip compression for HTTP clients and servers.
package pkg

import (
	"compress/gzip"
	"io"
	"net/http"
)

// compressWriter implements the http.ResponseWriter interface and enables transparent
// compression of data being sent to the client. It also sets the appropriate HTTP headers.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter creates and returns a new compressWriter that wraps the provided
// http.ResponseWriter. It initializes the gzip.Writer for compressing the response data.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statuscode int) {
	if statuscode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statuscode)
}

// Close finalizes the gzip.Writer by closing it, ensuring that all buffered data is flushed
// and written to the client. Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader implements the io.ReadCloser interface and allows transparent decompression
// of data received from the client using gzip compression.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader creates and returns a new compressReader that wraps the provided io.ReadCloser.
// It initializes the gzip.Reader for decompressing the request data.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.zr.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
