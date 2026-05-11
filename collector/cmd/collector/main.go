package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mitp/lab14/collector/internal/social"
)

func main() {
	var (
		outDir     = flag.String("out", "data", "output directory")
		count      = flag.Int("count", 240, "number of synthetic posts to collect")
		batchSize  = flag.Int("batch", 50, "JSONL batch size")
		windowSize = flag.Duration("window", time.Minute, "aggregation window")
		serve      = flag.Bool("serve", false, "start Arrow HTTP server after collection")
		addr       = flag.String("addr", ":8080", "HTTP server address")
		topicsFlag = flag.String("topics", "ai,fintech,travel,gaming", "comma-separated topics")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	topics := splitTopics(*topicsFlag)
	metrics, err := collect(ctx, *outDir, *count, *batchSize, *windowSize, topics)
	if err != nil {
		log.Fatalf("collect: %v", err)
	}
	if *serve {
		if err := serveArrow(ctx, *addr, metrics); err != nil {
			log.Fatalf("serve arrow: %v", err)
		}
	}
}

func collect(ctx context.Context, outDir string, count, batchSize int, window time.Duration, topics []string) ([]social.WindowMetric, error) {
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

func init() {
	log.SetOutput(os.Stdout)
}
