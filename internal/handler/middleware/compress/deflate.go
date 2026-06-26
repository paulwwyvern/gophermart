package compress

import (
	"compress/flate"
	"io"
	"net/http"
)

type deflateResponseWriter struct {
	http.ResponseWriter
	z *flate.Writer
}

func newDeflateResponseWriter(w http.ResponseWriter) *deflateResponseWriter {
	z, _ := flate.NewWriter(w, flate.BestCompression)
	return &deflateResponseWriter{ResponseWriter: w, z: z}
}

func (w *deflateResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *deflateResponseWriter) Write(b []byte) (int, error) {
	return w.z.Write(b)
}

func (w *deflateResponseWriter) WriteHeader(code int) {
	w.Header().Set("Content-Encoding", "deflate")

	w.ResponseWriter.WriteHeader(code)
}

func (w *deflateResponseWriter) Close() error {
	return w.z.Close()
}

type deflateReader struct {
	r io.ReadCloser
	z io.ReadCloser
}

func newDeflateReader(r io.ReadCloser) *deflateReader {
	z := flate.NewReader(r)

	return &deflateReader{r: r, z: z}
}

func (r *deflateReader) Read(b []byte) (int, error) {
	return r.z.Read(b)
}

func (r *deflateReader) Close() error {
	return r.z.Close()
}
