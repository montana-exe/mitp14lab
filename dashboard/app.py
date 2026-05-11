from __future__ import annotations

import os
from pathlib import Path

import plotly.express as px
import polars as pl
import requests
import streamlit as st

from pipeline.arrow_client import fetch_arrow

PARQUET_PATH = Path(os.getenv("PARQUET_PATH", "reports/social_windows.parquet"))
ARROW_URL = os.getenv("ARROW_URL", "")


def load_data(topic: str | None) -> pl.DataFrame:
    if ARROW_URL:
        try:
            return fetch_arrow(ARROW_URL, topic=topic)
        except requests.RequestException as exc:
            st.error(f"Arrow endpoint is unavailable: {exc}")
    if not PARQUET_PATH.exists():
        st.warning(f"Parquet file not found: {PARQUET_PATH}")
        return pl.DataFrame()
    df = pl.read_parquet(PARQUET_PATH)
    if topic:
        df = df.filter(pl.col("topic") == topic)
    return df


st.set_page_config(page_title="Lab14 Social Pipeline", layout="wide")
st.title("Social Network Monitoring Pipeline")

base_df = load_data(None)
if base_df.is_empty():
    st.stop()

topics = base_df.select("topic").unique().sort("topic")["topic"].to_list()
topic = st.sidebar.selectbox("Topic", ["all", *topics])
selected_topic = None if topic == "all" else topic
df = load_data(selected_topic)

total_posts = int(df["post_count"].sum())
total_engagement = int(df["total_engagement"].sum())
avg_sentiment = round(float(df["avg_sentiment"].mean()), 3)

col1, col2, col3 = st.columns(3)
col1.metric("Posts", total_posts)
col2.metric("Engagement", total_engagement)
col3.metric("Average sentiment", avg_sentiment)

pdf = df.to_pandas()
st.plotly_chart(
    px.line(pdf, x="window_start", y="avg_sentiment", color="topic", markers=True, title="Sentiment trend"),
    use_container_width=True,
)
st.plotly_chart(
    px.bar(pdf, x="topic", y="total_engagement", color="topic", title="Engagement by topic"),
    use_container_width=True,
)
st.dataframe(pdf, use_container_width=True)
