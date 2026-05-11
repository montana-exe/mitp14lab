package main

import (
	"context"
	"testing"
	"time"

	"mitp/lab14/collector/internal/social"
)

func TestBelongsToShardIsDeterministic(t *testing.T) {
	first := belongsToShard("post-42", 0, 3)
	second := belongsToShard("post-42", 0, 3)

	if first != second {
		t.Fatal("expected shard decision to be deterministic")
	}
}

func TestCollectRejectsInvalidShardConfig(t *testing.T) {
	_, err := collect(
		context.Background(),
		t.TempDir(),
		10,
		5,
		time.Minute,
		[]string{"ai"},
		CollectorConfig{ID: "test", ShardIndex: 3, ShardTotal: 2},
	)

	if err == nil {
		t.Fatal("expected invalid shard configuration to fail")
	}
}

func TestCollectAppliesShardFilter(t *testing.T) {
	allMetrics, err := collect(
		context.Background(),
		t.TempDir(),
		80,
		20,
		time.Minute,
		[]string{"ai"},
		CollectorConfig{ID: "all", ShardIndex: 0, ShardTotal: 1},
	)
	if err != nil {
		t.Fatalf("collect all shards: %v", err)
	}

	shardMetrics, err := collect(
		context.Background(),
		t.TempDir(),
		80,
		20,
		time.Minute,
		[]string{"ai"},
		CollectorConfig{ID: "shard", ShardIndex: 0, ShardTotal: 2},
	)
	if err != nil {
		t.Fatalf("collect shard: %v", err)
	}

	allPosts := sumPosts(allMetrics)
	shardPosts := sumPosts(shardMetrics)
	if shardPosts <= 0 || shardPosts >= allPosts {
		t.Fatalf("expected shard to process a non-empty subset, got shard=%d all=%d", shardPosts, allPosts)
	}
}

func sumPosts(metrics []social.WindowMetric) int64 {
	var total int64
	for _, metric := range metrics {
		total += metric.PostCount
	}
	return total
}
