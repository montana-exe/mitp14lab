package social

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Simulator struct {
	rnd       *rand.Rand
	topics    []string
	languages []string
}

func NewSimulator(seed int64, topics []string) *Simulator {
	if len(topics) == 0 {
		topics = []string{"ai", "fintech", "travel", "gaming"}
	}
	return &Simulator{
		rnd:       rand.New(rand.NewSource(seed)),
		topics:    topics,
		languages: []string{"en", "ru", "es"},
	}
}

func (s *Simulator) Next(index int, at time.Time) Post {
	topic := s.topics[index%len(s.topics)]
	sentiment := s.sentiment(topic, index)
	likes := int64(5 + s.rnd.Intn(500))
	shares := int64(s.rnd.Intn(80))
	replies := int64(s.rnd.Intn(45))
	views := likes*int64(15+s.rnd.Intn(30)) + shares*20 + replies*10
	return Post{
		Timestamp:      at.UTC(),
		PostID:         fmt.Sprintf("post-%06d", index),
		AuthorID:       fmt.Sprintf("author-%03d", 1+s.rnd.Intn(120)),
		Platform:       "x-emulator",
		Text:           fmt.Sprintf("Synthetic discussion about #%s sentiment %.2f", topic, sentiment),
		Topic:          topic,
		Hashtags:       []string{"#" + topic, "#lab14", "#etl"},
		SentimentScore: sentiment,
		Likes:          likes,
		Shares:         shares,
		Replies:        replies,
		Views:          views,
		Engagement:     likes + shares*2 + replies*3,
		Language:       s.languages[index%len(s.languages)],
	}
}

func (s *Simulator) sentiment(topic string, index int) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(topic)))
	base := float64(int(h.Sum32()%100)-50) / 100
	wave := math.Sin(float64(index)/7) * 0.35
	noise := (s.rnd.Float64() - 0.5) * 0.35
	value := base + wave + noise
	return math.Round(math.Max(-1, math.Min(1, value))*1000) / 1000
}
