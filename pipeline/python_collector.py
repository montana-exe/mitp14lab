from __future__ import annotations

import asyncio
import math
import random
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone


@dataclass(frozen=True)
class SyntheticPost:
    timestamp: datetime
    post_id: str
    topic: str
    sentiment_score: float
    engagement: int


async def collect_posts(count: int = 240, topics: list[str] | None = None, seed: int = 42) -> list[SyntheticPost]:
    rnd = random.Random(seed)
    topics = topics or ["ai", "fintech", "travel", "gaming"]
    start = datetime(2026, 5, 10, 12, 0, tzinfo=timezone.utc)

    async def build(index: int) -> SyntheticPost:
        await asyncio.sleep(0)
        topic = topics[index % len(topics)]
        wave = math.sin(index / 7) * 0.35
        sentiment = max(-1, min(1, wave + (rnd.random() - 0.5) * 0.35))
        likes = 5 + rnd.randint(0, 500)
        shares = rnd.randint(0, 80)
        replies = rnd.randint(0, 45)
        return SyntheticPost(
            timestamp=start + timedelta(seconds=index * 5),
            post_id=f"py-post-{index:06d}",
            topic=topic,
            sentiment_score=round(sentiment, 3),
            engagement=likes + shares * 2 + replies * 3,
        )

    return await asyncio.gather(*(build(index) for index in range(count)))
