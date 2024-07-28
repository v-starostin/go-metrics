package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/crypto"
	"github.com/v-starostin/go-metrics/internal/pb"
)

var (
	BuildVersion string
	BuildData    string
	BuildCommit  string
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Msgf("Build version: %s", getValue(BuildVersion))
	logger.Info().Msgf("Build data: %s", getValue(BuildData))
	logger.Info().Msgf("Build commit: %s", getValue(BuildCommit))

	cfg, err := config.NewAgent()
	if err != nil {
		logger.Error().Err(err).Msg("Configuration error")
		return
	}
	conn, err := grpc.NewClient(cfg.ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error().Err(err).Msg("Did not connect")
	}
	defer conn.Close()
	client := pb.NewGoMetricsClient(conn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var publicKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		publicKey, err = crypto.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			logger.Error().Err(err).Msg("Error to load public key")
			return
		}
	}

	a := agent.New(&logger, client, cfg.ServerAddress, cfg.Key, publicKey)

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

	go a.CollectRuntimeMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)
	go a.CollectGopsutilMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)

	metrics := a.PrepareMetrics(ctx, time.Duration(cfg.ReportInterval)*time.Second)
	ip, err := getIP()
	if err != nil {
		logger.Error().Err(err).Msg("Error to get IP")
		return
	}

	for i := 0; i < cfg.RateLimit; i++ {
		go a.Retry(ctx, 3, func(ctx context.Context) error {
			return a.SendMetrics(ctx, metrics, ip)
		}, 1*time.Second, 3*time.Second, 5*time.Second)
	}

	<-ctx.Done()
	if ctx.Err() != nil {
		logger.Info().Msgf("Received shutdown signal, stopping work: %v", ctx.Err())
	}

	logger.Info().Msg("Waiting 5 seconds to complete pending operations...")
	time.Sleep(5 * time.Second)

	logger.Info().Msg("Finished collecting metrics and shutting down gracefully.")
}

func getValue(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func getIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			return ipNet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("no IP address found")
}
