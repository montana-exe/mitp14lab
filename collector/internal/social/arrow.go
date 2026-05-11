package social

import (
	"io"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/ipc"
	"github.com/apache/arrow/go/v17/arrow/memory"
)

func WriteArrow(metrics []WindowMetric, out io.Writer) error {
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "window_start", Type: arrow.BinaryTypes.String},
		{Name: "window_end", Type: arrow.BinaryTypes.String},
		{Name: "topic", Type: arrow.BinaryTypes.String},
		{Name: "post_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "positive_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "negative_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "neutral_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "min_sentiment", Type: arrow.PrimitiveTypes.Float64},
		{Name: "max_sentiment", Type: arrow.PrimitiveTypes.Float64},
		{Name: "avg_sentiment", Type: arrow.PrimitiveTypes.Float64},
		{Name: "total_engagement", Type: arrow.PrimitiveTypes.Int64},
		{Name: "unique_authors", Type: arrow.PrimitiveTypes.Int64},
	}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	defer builder.Release()

	for _, metric := range metrics {
		builder.Field(0).(*array.StringBuilder).Append(metric.WindowStart.Format("2006-01-02T15:04:05Z07:00"))
		builder.Field(1).(*array.StringBuilder).Append(metric.WindowEnd.Format("2006-01-02T15:04:05Z07:00"))
		builder.Field(2).(*array.StringBuilder).Append(metric.Topic)
		builder.Field(3).(*array.Int64Builder).Append(metric.PostCount)
		builder.Field(4).(*array.Int64Builder).Append(metric.PositiveCount)
		builder.Field(5).(*array.Int64Builder).Append(metric.NegativeCount)
		builder.Field(6).(*array.Int64Builder).Append(metric.NeutralCount)
		builder.Field(7).(*array.Float64Builder).Append(metric.MinSentiment)
		builder.Field(8).(*array.Float64Builder).Append(metric.MaxSentiment)
		builder.Field(9).(*array.Float64Builder).Append(metric.AvgSentiment)
		builder.Field(10).(*array.Int64Builder).Append(metric.TotalEngagement)
		builder.Field(11).(*array.Int64Builder).Append(metric.UniqueAuthors)
	}
	record := builder.NewRecord()
	defer record.Release()

	writer := ipc.NewWriter(out, ipc.WithSchema(schema))
	if err := writer.Write(record); err != nil {
		_ = writer.Close()
		return err
	}
	return writer.Close()
}
