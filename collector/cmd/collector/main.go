package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mitp/lab14/collector/internal/social"

	"github.com/nats-io/nats.go"
)

func main() {
	var (
		outDir        = flag.String("out", "data", "output directory")
		count         = flag.Int("count", 240, "number of synthetic posts to collect")
		batchSize     = flag.Int("batch", 50, "JSONL batch size")
		windowSize    = flag.Duration("window", time.Minute, "aggregation window")
		serve         = flag.Bool("serve", false, "start Arrow HTTP server after collection")
		addr          = flag.String("addr", ":8080", "HTTP server address")
		topicsFlag    = flag.String("topics", "ai,fintech,travel,gaming", "comma-separated topics")
		collectorID   = flag.String("collector-id", envOrDefault("COLLECTOR_ID", "collector-local"), "collector identity for distributed coordination")
		shardIndex    = flag.Int("shard-index", envIntOrDefault("SHARD_INDEX", 0), "zero-based shard index owned by this collector; -1 derives it from collector-id")
		shardTotal    = flag.Int("shard-total", envIntOrDefault("SHARD_TOTAL", 1), "total shard count across collectors")
		shardStrategy = flag.String("shard-strategy", envOrDefault("SHARD_STRATEGY", "hash"), "shard strategy: hash, topic, or author-range")
		etcdEndpoint  = flag.String("etcd-endpoint", os.Getenv("ETCD_ENDPOINT"), "etcd HTTP endpoint, for example http://etcd:2379")
		natsURL       = flag.String("nats-url", os.Getenv("NATS_URL"), "NATS broker URL, for example nats://nats:4222")
		subject       = flag.String("stream-subject", envOrDefault("STREAM_SUBJECT", "social.windows"), "NATS subject for window metrics")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	topics := splitTopics(*topicsFlag)
	cfg := CollectorConfig{
		ID:            *collectorID,
		ShardIndex:    *shardIndex,
		ShardTotal:    *shardTotal,
		ShardStrategy: strings.ToLower(strings.TrimSpace(*shardStrategy)),
		EtcdEndpoint:  *etcdEndpoint,
		NATSURL:       *natsURL,
		StreamSubject: *subject,
	}
	if err := registerCollector(ctx, cfg); err != nil {
		log.Printf("etcd registration skipped: %v", err)
	}
	metrics, err := collect(ctx, *outDir, *count, *batchSize, *windowSize, topics, cfg)
	if err != nil {
		log.Fatalf("collect: %v", err)
	}
	if err := publishWindows(ctx, cfg, metrics); err != nil {
		log.Fatalf("publish windows: %v", err)
	}
	if *serve {
		if err := serveArrow(ctx, *addr, metrics); err != nil {
			log.Fatalf("serve arrow: %v", err)
		}
	}
}

type CollectorConfig struct {
	ID            string `json:"id"`
	ShardIndex    int    `json:"shard_index"`
	ShardTotal    int    `json:"shard_total"`
	ShardStrategy string `json:"shard_strategy"`
	EtcdEndpoint  string `json:"etcd_endpoint"`
	NATSURL       string `json:"nats_url"`
	StreamSubject string `json:"stream_subject"`
}

func collect(ctx context.Context, outDir string, count, batchSize int, window time.Duration, topics []string, cfg CollectorConfig) ([]social.WindowMetric, error) {
	if cfg.ShardTotal < 1 {
		return nil, fmt.Errorf("shard-total must be positive, got %d", cfg.ShardTotal)
	}
	cfg.ShardIndex = normalizeShardIndex(cfg)
	if cfg.ShardIndex < 0 || cfg.ShardIndex >= cfg.ShardTotal {
		return nil, fmt.Errorf("shard-index must be in [0,%d), got %d", cfg.ShardTotal, cfg.ShardIndex)
	}
	if cfg.ShardStrategy == "" {
		cfg.ShardStrategy = "hash"
	}
	simulator := social.NewSimulator(42, topics)
	aggregator := social.NewAggregator(window)
	postWriter := social.NewJSONLWriter[social.Post](outDir+"/posts.jsonl", batchSize, 500*time.Millisecond)
	windowWriter := social.NewJSONLWriter[social.WindowMetric](outDir+"/windows.jsonl", batchSize, 500*time.Millisecond)

	start := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			log.Print("shutdown requested, flushing current buffers")
			i = count
			continue
		default:
		}
		post := simulator.Next(i, start.Add(time.Duration(i)*5*time.Second))
		if !assignedToShard(post, cfg) {
			continue
		}
		aggregator.Add(post)
		if err := postWriter.Add(post); err != nil {
			return nil, err
		}
	}
	if err := postWriter.Flush(); err != nil {
		return nil, err
	}
	metrics := aggregator.Metrics()
	for _, metric := range metrics {
		if err := windowWriter.Add(metric); err != nil {
			return nil, err
		}
	}
	if err := windowWriter.Flush(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func normalizeShardIndex(cfg CollectorConfig) int {
	if cfg.ShardIndex >= 0 {
		return cfg.ShardIndex
	}
	if cfg.ShardTotal <= 1 {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(cfg.ID))
	return int(hasher.Sum32() % uint32(cfg.ShardTotal))
}

func assignedToShard(post social.Post, cfg CollectorConfig) bool {
	switch cfg.ShardStrategy {
	case "hash":
		return belongsToShard(post.PostID, cfg.ShardIndex, cfg.ShardTotal)
	case "topic":
		return belongsToShard(strings.ToLower(post.Topic), cfg.ShardIndex, cfg.ShardTotal)
	case "author-range":
		return belongsToShard(post.AuthorID, cfg.ShardIndex, cfg.ShardTotal)
	default:
		return belongsToShard(post.PostID, cfg.ShardIndex, cfg.ShardTotal)
	}
}

func belongsToShard(key string, shardIndex, shardTotal int) bool {
	if shardTotal <= 1 {
		return true
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return int(hasher.Sum32()%uint32(shardTotal)) == shardIndex
}

func registerCollector(ctx context.Context, cfg CollectorConfig) error {
	if cfg.EtcdEndpoint == "" {
		return nil
	}
	payload := map[string]string{
		"key":   base64.StdEncoding.EncodeToString([]byte("/lab14/collectors/" + cfg.ID)),
		"value": base64.StdEncoding.EncodeToString(mustJSON(cfg)),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := strings.TrimRight(cfg.EtcdEndpoint, "/") + "/v3/kv/put"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("etcd returned %s", resp.Status)
	}
	log.Printf("collector %s registered in etcd with shard %d/%d", cfg.ID, cfg.ShardIndex, cfg.ShardTotal)
	return nil
}

func publishWindows(ctx context.Context, cfg CollectorConfig, metrics []social.WindowMetric) error {
	if cfg.NATSURL == "" || len(metrics) == 0 {
		return nil
	}
	nc, err := nats.Connect(cfg.NATSURL, nats.Name("lab14-"+cfg.ID), nats.Timeout(5*time.Second))
	if err != nil {
		return err
	}
	defer nc.Close()
	for _, metric := range metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		payload, err := json.Marshal(metric)
		if err != nil {
			return err
		}
		if err := nc.Publish(cfg.StreamSubject, payload); err != nil {
			return err
		}
	}
	if err := nc.FlushTimeout(5 * time.Second); err != nil {
		return err
	}
	log.Printf("published %d window metrics to %s", len(metrics), cfg.StreamSubject)
	return nil
}

func mustJSON(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return data
}

func serveArrow(ctx context.Context, addr string, metrics []social.WindowMetric) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/arrow", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		filtered := filterByTopic(metrics, topic)
		var buf bytes.Buffer
		if err := social.WriteArrow(filtered, &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.apache.arrow.stream")
		_, _ = w.Write(buf.Bytes())
	})

	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("arrow server listening on %s", addr)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func filterByTopic(metrics []social.WindowMetric, topic string) []social.WindowMetric {
	if topic == "" {
		return metrics
	}
	result := make([]social.WindowMetric, 0, len(metrics))
	for _, metric := range metrics {
		if metric.Topic == topic {
			result = append(result, metric)
		}
	}
	return result
}

func splitTopics(value string) []string {
	parts := strings.Split(value, ",")
	topics := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			topics = append(topics, part)
		}
	}
	return topics
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func init() {
	log.SetOutput(os.Stdout)
}
