from __future__ import annotations

import asyncio

from pipeline.python_collector import collect_posts


def test_python_collector_generates_requested_count() -> None:
    posts = asyncio.run(collect_posts(count=10, topics=["ai"]))

    assert len(posts) == 10
    assert {post.topic for post in posts} == {"ai"}
    assert all(-1 <= post.sentiment_score <= 1 for post in posts)
