package compress

import (
	"net/http"
	"strings"
)

// WithCompress разпаковывает запрос, если запрос был сжат и сжимает ответ, если пользователь поддерживает это
//
// Доступные алгоритмы сжатия:
//
// gzip, deflate
func WithCompress() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")
			supportDeflate := strings.Contains(acceptEncoding, "deflate")
			if supportGzip {
				cw := newGzipResponseWriter(w)
				ow = cw

				defer cw.Close()
			} else if supportDeflate {
				cw := newDeflateResponseWriter(w)
				ow = cw

				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			sendsDeflate := strings.Contains(contentEncoding, "deflate")

			if sendsGzip {
				cr, err := newGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = cr
				defer r.Body.Close()
			} else if sendsDeflate {
				cr := newDeflateReader(r.Body)
				r.Body = cr
				defer r.Body.Close()
			}

			h.ServeHTTP(ow, r)
		})
	}
}
