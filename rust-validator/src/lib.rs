use serde_json::Value;

pub fn validate_record_json(line: &str) -> Result<(), String> {
    let value: Value = serde_json::from_str(line).map_err(|err| format!("invalid json: {err}"))?;
    if value.get("post_id").is_some() {
        validate_post(&value)
    } else if value.get("window_start").is_some() {
        validate_window(&value)
    } else {
        Err("record is neither social post nor window metric".to_string())
    }
}

fn validate_post(value: &Value) -> Result<(), String> {
    require_string(value, "timestamp")?;
    require_string(value, "post_id")?;
    require_string(value, "author_id")?;
    require_string(value, "platform")?;
    require_string(value, "topic")?;
    require_string(value, "language")?;
    validate_sentiment(value, "sentiment_score")?;
    for field in ["likes", "shares", "replies", "views", "engagement"] {
        require_non_negative_i64(value, field)?;
    }
    Ok(())
}

fn validate_window(value: &Value) -> Result<(), String> {
    require_string(value, "window_start")?;
    require_string(value, "window_end")?;
    require_string(value, "topic")?;
    for field in [
        "post_count",
        "positive_count",
        "negative_count",
        "neutral_count",
        "total_engagement",
        "unique_authors",
    ] {
        require_non_negative_i64(value, field)?;
    }
    for field in ["min_sentiment", "max_sentiment", "avg_sentiment"] {
        validate_sentiment(value, field)?;
    }
    Ok(())
}

fn require_string(value: &Value, field: &str) -> Result<(), String> {
    match value.get(field).and_then(Value::as_str) {
        Some(text) if !text.trim().is_empty() => Ok(()),
        _ => Err(format!("{field} must be a non-empty string")),
    }
}

fn require_non_negative_i64(value: &Value, field: &str) -> Result<(), String> {
    match value.get(field).and_then(Value::as_i64) {
        Some(number) if number >= 0 => Ok(()),
        _ => Err(format!("{field} must be a non-negative integer")),
    }
}

fn validate_sentiment(value: &Value, field: &str) -> Result<(), String> {
    match value.get(field).and_then(Value::as_f64) {
        Some(score) if (-1.0..=1.0).contains(&score) => Ok(()),
        _ => Err(format!("{field} must be between -1 and 1")),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn validates_social_post() {
        let line = r#"{
            "timestamp":"2026-05-10T12:00:00Z",
            "post_id":"post-1",
            "author_id":"author-1",
            "platform":"microblog",
            "topic":"ai",
            "language":"ru",
            "sentiment_score":0.42,
            "likes":10,
            "shares":2,
            "replies":1,
            "views":200,
            "engagement":13
        }"#;

        assert!(validate_record_json(line).is_ok());
    }

    #[test]
    fn rejects_invalid_sentiment() {
        let line = r#"{
            "timestamp":"2026-05-10T12:00:00Z",
            "post_id":"post-1",
            "author_id":"author-1",
            "platform":"microblog",
            "topic":"ai",
            "language":"ru",
            "sentiment_score":2.0,
            "likes":10,
            "shares":2,
            "replies":1,
            "views":200,
            "engagement":13
        }"#;

        assert!(validate_record_json(line).is_err());
    }

    #[test]
    fn validates_window_metric() {
        let line = r#"{
            "window_start":"2026-05-10T12:00:00Z",
            "window_end":"2026-05-10T12:01:00Z",
            "topic":"ai",
            "post_count":5,
            "positive_count":3,
            "negative_count":1,
            "neutral_count":1,
            "min_sentiment":-0.2,
            "max_sentiment":0.8,
            "avg_sentiment":0.3,
            "total_engagement":100,
            "unique_authors":4
        }"#;

        assert!(validate_record_json(line).is_ok());
    }
}
