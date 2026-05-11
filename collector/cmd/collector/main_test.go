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

func TestNormalizeShardIndexDerivesFromCollectorID(t *testing.T) {
	first := normalizeShardIndex(CollectorConfig{ID: "collector-a", ShardIndex: -1, ShardTotal: 4})
	second := normalizeShardIndex(CollectorConfig{ID: "collector-a", ShardIndex: -1, ShardTotal: 4})

	if first != second {
		t.Fatal("expected derived shard index to be deterministic")
	}
	if first < 0 || first >= 4 {
		t.Fatalf("derived shard index out of range: %d", first)
	}
}

func TestAssignedToShardSupportsTopicStrategy(t *testing.T) {
	post := social.Post{PostID: "post-1", Topic: "ai", AuthorID: "author-001"}
	topicShard := normalizeShardIndex(CollectorConfig{ID: "topic-owner", ShardIndex: -1, ShardTotal: 3})

	got := assignedToShard(post, CollectorConfig{
		ID:            "topic-owner",
		ShardIndex:    topicShard,
		ShardTotal:    3,
		ShardStrategy: "topic",
	})
	want := belongsToShard("ai", topicShard, 3)

	if got != want {
		t.Fatalf("topic strategy mismatch: got %v want %v", got, want)
	}
}

func TestAssignedToShardSupportsAuthorRangeStrategy(t *testing.T) {
	post := social.Post{PostID: "post-1", Topic: "ai", AuthorID: "author-042"}
	authorShard := normalizeShardIndex(CollectorConfig{ID: "author-owner", ShardIndex: -1, ShardTotal: 4})

	got := assignedToShard(post, CollectorConfig{
		ID:            "author-owner",
		ShardIndex:    authorShard,
		ShardTotal:    4,
		ShardStrategy: "author-range",
	})
	want := belongsToShard("author-042", authorShard, 4)

	if got != want {
		t.Fatalf("author-range strategy mismatch: got %v want %v", got, want)
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
