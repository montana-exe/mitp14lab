package social

import (
	"testing"
	"time"
)

func TestAggregatorBuildsWindowMetrics(t *testing.T) {
	agg := NewAggregator(time.Minute)
	base := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	agg.Add(Post{Timestamp: base, Topic: "ai", AuthorID: "a1", SentimentScore: 0.4, Engagement: 10})
	agg.Add(Post{Timestamp: base.Add(10 * time.Second), Topic: "ai", AuthorID: "a2", SentimentScore: -0.3, Engagement: 20})
	agg.Add(Post{Timestamp: base.Add(time.Minute), Topic: "ai", AuthorID: "a1", SentimentScore: 0, Engagement: 5})

	metrics := agg.Metrics()
	if len(metrics) != 2 {
		t.Fatalf("got %d windows, want 2", len(metrics))
	}
	if metrics[0].PostCount != 2 || metrics[0].PositiveCount != 1 || metrics[0].NegativeCount != 1 {
		t.Fatalf("unexpected first metric: %+v", metrics[0])
	}
	if metrics[0].UniqueAuthors != 2 {
		t.Fatalf("unique authors = %d, want 2", metrics[0].UniqueAuthors)
	}
}

func TestSimulatorIsDeterministic(t *testing.T) {
	topics := []string{"ai"}
	first := NewSimulator(7, topics).Next(1, time.Unix(0, 0))
	second := NewSimulator(7, topics).Next(1, time.Unix(0, 0))
	if first.SentimentScore != second.SentimentScore || first.Engagement != second.Engagement {
		t.Fatalf("simulator is not deterministic: %+v != %+v", first, second)
	}
}
