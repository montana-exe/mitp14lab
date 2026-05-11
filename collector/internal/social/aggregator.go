package social

import (
	"math"
	"sort"
	"time"
)

type bucket struct {
	start      time.Time
	end        time.Time
	topic      string
	count      int64
	positive   int64
	negative   int64
	neutral    int64
	min        float64
	max        float64
	sum        float64
	engagement int64
	authors    map[string]struct{}
}

type Aggregator struct {
	window time.Duration
	data   map[string]*bucket
}

func NewAggregator(window time.Duration) *Aggregator {
	if window <= 0 {
		window = time.Minute
	}
	return &Aggregator{window: window, data: make(map[string]*bucket)}
}

func (a *Aggregator) Add(post Post) {
	start := post.Timestamp.UTC().Truncate(a.window)
	key := start.Format(time.RFC3339Nano) + "|" + post.Topic
	b, ok := a.data[key]
	if !ok {
		b = &bucket{
			start:   start,
			end:     start.Add(a.window),
			topic:   post.Topic,
			min:     post.SentimentScore,
			max:     post.SentimentScore,
			authors: make(map[string]struct{}),
		}
		a.data[key] = b
	}
	b.count++
	b.sum += post.SentimentScore
	b.engagement += post.Engagement
	b.min = math.Min(b.min, post.SentimentScore)
	b.max = math.Max(b.max, post.SentimentScore)
	b.authors[post.AuthorID] = struct{}{}
	switch {
	case post.SentimentScore > 0.15:
		b.positive++
	case post.SentimentScore < -0.15:
		b.negative++
	default:
		b.neutral++
	}
}

func (a *Aggregator) Metrics() []WindowMetric {
	items := make([]WindowMetric, 0, len(a.data))
	for _, b := range a.data {
		avg := 0.0
		if b.count > 0 {
			avg = math.Round((b.sum/float64(b.count))*1000) / 1000
		}
		items = append(items, WindowMetric{
			WindowStart:     b.start,
			WindowEnd:       b.end,
			Topic:           b.topic,
			PostCount:       b.count,
			PositiveCount:   b.positive,
			NegativeCount:   b.negative,
			NeutralCount:    b.neutral,
			MinSentiment:    b.min,
			MaxSentiment:    b.max,
			AvgSentiment:    avg,
			TotalEngagement: b.engagement,
			UniqueAuthors:   int64(len(b.authors)),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].WindowStart.Equal(items[j].WindowStart) {
			return items[i].Topic < items[j].Topic
		}
		return items[i].WindowStart.Before(items[j].WindowStart)
	})
	return items
}
