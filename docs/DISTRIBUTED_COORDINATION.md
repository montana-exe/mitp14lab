# Распределённая координация

Collectors регистрируются в etcd по ключу:

```text
/lab14/collectors/<collector-id>
```

В value сохраняются `collector_id`, `shard_index`, `shard_total`, NATS subject и endpoint координации.

## Sharding

Collector поддерживает три стратегии:

```text
hash:         hash(post_id) % shard_total == shard_index
topic:        hash(topic) % shard_total == shard_index
author-range: hash(author_id) % shard_total == shard_index
```

Если `shard-index = -1`, collector выводит индекс из `collector-id`. Это удобно для Kubernetes, где имя pod стабильно доступно через Downward API.

## Почему etcd

etcd отделяет coordination state от stream broker. NATS отвечает за доставку событий, etcd - за список активных collectors и параметры shard ownership.
