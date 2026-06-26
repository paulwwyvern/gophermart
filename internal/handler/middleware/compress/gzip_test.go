package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func echo(w http.ResponseWriter, r *http.Request) {
	buf, _ := io.ReadAll(r.Body)

	w.WriteHeader(200)
	w.Write(buf)
}

func TestCompress_Uncompressed(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 uncompressed",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.want, w.Body.String())
		})
	}
}

func TestCompress_Gzip_Input(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress gzip input",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer

			z := gzip.NewWriter(&buf)
			_, err := z.Write([]byte(tt.body))
			require.NoError(t, err)

			err = z.Close()
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Encoding", "gzip")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.want, w.Body.String())
		})
	}
}

func TestCompress_Gzip_Output(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress gzip output",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			r.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			z, err := gzip.NewReader(w.Body)
			require.NoError(t, err)

			b, err := io.ReadAll(z)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestCompress_Gzip_Both(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress gzip both",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer

			zw := gzip.NewWriter(&buf)
			_, err := zw.Write([]byte(tt.body))
			require.NoError(t, err)

			err = zw.Close()
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Encoding", "gzip")
			r.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			zr, err := gzip.NewReader(w.Body)
			require.NoError(t, err)

			b, err := io.ReadAll(zr)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(b))
		})
	}
}
