package handler_test

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/v-starostin/go-metrics/internal/handler"
)

func TestDecompress(t *testing.T) {
	l := zerolog.New(os.Stdout)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	t.Run("content encoding gzip", func(t *testing.T) {
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		gzw.Write([]byte("test"))
		gzw.Close()
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		rr := httptest.NewRecorder()
		handler.Decompress(&l)(testHandler).ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, []byte("test"), resBody)
	})

	t.Run("no content encoding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("test")))
		rr := httptest.NewRecorder()
		handler.Decompress(&l)(testHandler).ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, []byte("test"), resBody)
	})
}

func TestCheckHash(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	t.Run("no HashSHA256 header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("test")))
		rr := httptest.NewRecorder()
		handler.CheckHash("key")(testHandler).ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("HashSHA256 header exists", func(t *testing.T) {
		var buf = &bytes.Buffer{}
		buf.Write([]byte("test"))
		req := httptest.NewRequest(http.MethodPost, "/", buf)
		buf2 := *buf
		h := hmac.New(sha256.New, []byte("key"))
		_, err := h.Write(buf2.Bytes())
		assert.NoError(t, err)
		d := h.Sum(nil)
		req.Header.Add("HashSHA256", hex.EncodeToString(d))
		rr := httptest.NewRecorder()
		handler.CheckHash("key")(testHandler).ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("HashSHA256 header exists, but hash values are not equal", func(t *testing.T) {
		var buf = &bytes.Buffer{}
		buf.Write([]byte("test"))
		req := httptest.NewRequest(http.MethodPost, "/", buf)
		h := hmac.New(sha256.New, []byte("key"))
		_, err := h.Write([]byte("test2"))
		assert.NoError(t, err)
		d := h.Sum(nil)
		req.Header.Add("HashSHA256", hex.EncodeToString(d))
		rr := httptest.NewRecorder()
		handler.CheckHash("key")(testHandler).ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
