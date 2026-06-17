package compress

import (
	"bytes"
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompress_Deflate_Input(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress deflate input",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer

			z, _ := flate.NewWriter(&buf, flate.BestCompression)
			_, err := z.Write([]byte(tt.body))
			require.NoError(t, err)

			err = z.Close()
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Encoding", "deflate")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.want, w.Body.String())
		})
	}
}

func TestCompress_Deflate_Output(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress deflate output",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			r.Header.Set("Accept-Encoding", "deflate")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			z := flate.NewReader(w.Body)

			b, err := io.ReadAll(z)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestCompress_Deflate_Both(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Test #1 compress deflate both",
			body: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}

	handler := WithCompress()(http.HandlerFunc(echo))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer

			zw, _ := flate.NewWriter(&buf, flate.BestCompression)
			_, err := zw.Write([]byte(tt.body))
			require.NoError(t, err)

			err = zw.Close()
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Encoding", "deflate")
			r.Header.Set("Accept-Encoding", "deflate")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			zr := flate.NewReader(w.Body)

			b, err := io.ReadAll(zr)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(b))
		})
	}
}
