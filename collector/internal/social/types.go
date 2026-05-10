package social

import "time"

type Post struct {
	Timestamp      time.Time `json:"timestamp"`
	PostID         string    `json:"post_id"`
	AuthorID       string    `json:"author_id"`
	Platform       string    `json:"platform"`
	Text           string    `json:"text"`
	Topic          string    `json:"topic"`
	Hashtags       []string  `json:"hashtags"`
	SentimentScore float64   `json:"sentiment_score"`
	Likes          int64     `json:"likes"`
	Shares         int64     `json:"shares"`
	Replies        int64     `json:"replies"`
	Views          int64     `json:"views"`
	Engagement     int64     `json:"engagement"`
	Language       string    `json:"language"`
}

type WindowMetric struct {
	WindowStart     time.Time `json:"window_start"`
	WindowEnd       time.Time `json:"window_end"`
	Topic           string    `json:"topic"`
	PostCount       int64     `json:"post_count"`
	PositiveCount   int64     `json:"positive_count"`
	NegativeCount   int64     `json:"negative_count"`
	NeutralCount    int64     `json:"neutral_count"`
	MinSentiment    float64   `json:"min_sentiment"`
	MaxSentiment    float64   `json:"max_sentiment"`
	AvgSentiment    float64   `json:"avg_sentiment"`
	TotalEngagement int64     `json:"total_engagement"`
	UniqueAuthors   int64     `json:"unique_authors"`
}
