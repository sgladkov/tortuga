package web

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sgladkov/tortuga/internal/blockchain"
	"github.com/sgladkov/tortuga/internal/logger"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writerToUse := w
		if ContainsHeaderValue(r, "Content-Encoding", "gzip") {
			// change original request body to decode its content
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			logger.Log.Info("Use request decode")
			r.Body = gz
		} else {
			logger.Log.Info("Don't use request decode")
		}

		if ContainsHeaderValue(r, "Accept-Encoding", "gzip") {
			// change writer to wrapped writer with gzip encoding
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer func() {
				err = gz.Close()
				if err != nil {
					logger.Log.Warn("Failed to close gzip writer", zap.Error(err))
				}
			}()

			logger.Log.Info("Use reply compress")
			w.Header().Set("Content-Encoding", "gzip")
			writerToUse = gzipWriter{ResponseWriter: w, Writer: gz}
		} else {
			// use original writer
			logger.Log.Info("Don't use reply compress")
		}

		h.ServeHTTP(writerToUse, r)
	})
}

type (
	responseData struct {
		status int
		size   int
		body   []byte
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.body = append(r.responseData.body, b...)
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		logger.Log.Info("request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", responseData.status),
			zap.String("responce", string(responseData.body)),
		)
	})
}

func AuthorizationHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/private") {
			h.ServeHTTP(w, r)
			return
		}
		address := r.Header.Get("TRTG-Address")
		if len(address) == 0 {
			logger.Log.Warn("no public key in headers")
			http.Error(w, "no public key in headers", http.StatusBadRequest)
			return
		}
		nonceStr := r.Header.Get("TRTG-Nonce")
		if len(nonceStr) == 0 {
			logger.Log.Warn("no nonce in headers")
			http.Error(w, "no nonce in headers", http.StatusBadRequest)
			return
		}
		nonce, err := strconv.ParseUint(nonceStr, 10, 64)
		if err != nil {
			logger.Log.Warn("invalid nonce", zap.Error(err))
			http.Error(w, "invalid nonce: "+err.Error(), http.StatusBadRequest)
			return

		}
		signatureStr := r.Header.Get("TRTG-Signature")
		if len(signatureStr) == 0 {
			logger.Log.Warn("no signature in headers")
			http.Error(w, "no signature in headers", http.StatusBadRequest)
			return
		}
		signature, err := hex.DecodeString(signatureStr)
		if err != nil {
			logger.Log.Warn("invalid signature", zap.Error(err))
			http.Error(w, "invalid signature: "+err.Error(), http.StatusBadRequest)
			return
		}

		msg, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Warn("error reading body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dataToSign := r.URL.Path + "&" + nonceStr + "&" + string(msg)
		logger.Log.Info("data to sign", zap.String("data", dataToSign))
		restoredAddress, err := blockchain.RestoreAddressFromSignature([]byte(dataToSign), signature)
		if err != nil {
			logger.Log.Warn("failed to restore address", zap.Error(err))
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}
		if restoredAddress != address {
			logger.Log.Warn("invalid signature", zap.String("address", address),
				zap.String("restored", restoredAddress))
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		user, err := storage.GetUser(address)
		if err != nil {
			logger.Log.Warn("failed to get user from storage", zap.Error(err))
			http.Error(w, "failed to get user info", http.StatusBadGateway)
			return
		}
		if user.Nonce >= nonce {
			logger.Log.Warn("invalid nonce", zap.Uint64("stored", user.Nonce),
				zap.Uint64("received", nonce))
			http.Error(w, "invalid nonce", http.StatusForbidden)
			return
		}

		err = storage.UpdateUserNonce(address, nonce)
		if err != nil {
			logger.Log.Error("failed to update user nonce", zap.Error(err))
		}

		r.Body = io.NopCloser(bytes.NewBuffer(msg))
		h.ServeHTTP(w, r)
	})
}
