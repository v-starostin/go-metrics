package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/v-starostin/go-metrics/internal/model"
)

type LogFormatter struct {
	*zerolog.Logger
}

func (l *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := l.With().
		Str("URI", r.RequestURI).
		Str("method", r.Method).
		Logger()

	return &LogEntry{&logger}
}

type LogEntry struct {
	*zerolog.Logger
}

func (l *LogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	l.Info().
		Str("elapsed", elapsed.String()).
		Int("bytes", bytes).
		Int("status", status).
		Msg("Request handled")
}

func (l *LogEntry) Panic(v any, stack []byte) {
	l.Info().
		Interface("panic", v).
		Bytes("stack", stack).
		Msg("Panic handled")
}

func Decompress(l *zerolog.Logger) func(next http.Handler) http.Handler {
	gr := new(gzip.Reader)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encodingHeaders := r.Header.Values("Content-Encoding")
			if !slices.Contains(encodingHeaders, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			if err := gr.Reset(r.Body); err != nil {
				l.Error().Err(err).Msg("Reset gzip reader error")
			}
			defer gr.Close()

			r.Body = gr
			next.ServeHTTP(w, r)
		})
	}
}

func CheckHash(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashSHA256 := r.Header.Get("HashSHA256")
			if hashSHA256 == "" {
				next.ServeHTTP(w, r)
				return
			}
			h := hmac.New(sha256.New, []byte(key))
			b, err := io.ReadAll(r.Body)
			if err != nil {
				writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad Request"})
				return
			}
			if _, err = h.Write(b); err != nil {
				writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal Server Error"})
				return
			}
			d := h.Sum(nil)
			hh, err := hex.DecodeString(hashSHA256)
			if err != nil {
				writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal Server Error"})
				return
			}
			if !hmac.Equal(d, hh) {
				writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad Request"})
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(b))
			next.ServeHTTP(w, r)
		})
	}
}

func CheckHashInterceptor(key string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		hashSHA256 := md.Get("hashsha256")
		if len(hashSHA256) == 0 {
			return handler(ctx, req)
		}

		h := hmac.New(sha256.New, []byte(key))

		b, err := proto.Marshal(req.(proto.Message))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal request")
		}

		if _, err = h.Write(b); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to write hash")
		}

		d := h.Sum(nil)
		hh, err := hex.DecodeString(hashSHA256[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid hash")
		}

		if !hmac.Equal(d, hh) {
			return nil, status.Errorf(codes.InvalidArgument, "hash mismatch")
		}

		return handler(ctx, req)
	}
}

func CheckIP(subnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == "" {
				next.ServeHTTP(w, r)
				return
			}
			realIP := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil {
				writeResponse(w, http.StatusBadRequest, model.Error{Error: "X-Real-IP is empty"})
				return
			}
			_, ipNet, err := net.ParseCIDR(subnet)
			if err != nil {
				writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Wrong subnet format"})
			}
			if !ipNet.Contains(ip) {
				writeResponse(w, http.StatusForbidden, model.Error{Error: "Forbidden"})
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CheckIPInterceptor(subnet string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if subnet == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		realIP := md.Get("x-real-ip")
		if len(realIP) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "X-Real-IP header is missing")
		}

		ip := net.ParseIP(realIP[0])
		if ip == nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid IP address in X-Real-IP header")
		}

		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid subnet format")
		}

		if !ipNet.Contains(ip) {
			return nil, status.Errorf(codes.PermissionDenied, "IP address not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
