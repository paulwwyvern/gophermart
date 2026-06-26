package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	z *gzip.Writer
}

func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{w, gzip.NewWriter(w)}
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.z.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {

	w.Header().Set("Content-Encoding", "gzip")

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Close() error {
	return w.z.Close()
}

type gzipReader struct {
	r io.ReadCloser
	z *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{r: r, z: z}, nil
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.z.Read(p)
}

func (r *gzipReader) Close() error {
	if err := r.z.Close(); err != nil {
		return err
	}

	return nil
}
