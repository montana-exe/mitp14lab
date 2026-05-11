# Kubernetes и HPA

Манифесты находятся в `k8s/` и применяются через Kustomize:

```bash
kubectl apply -k k8s
```

Состав:

- `etcd` deployment/service для coordination state.
- `nats` deployment/service для streaming broker.
- `collector` deployment/service с resource requests/limits.
- `analyzer` deployment для stream consumer.
- `dashboard` deployment/service для Streamlit.
- `collector-hpa` autoscaling/v2 HPA.

## HPA

`collector-hpa` масштабирует Go collectors:

```text
minReplicas: 2
maxReplicas: 6
target CPU utilization: 65%
```

Для production-кластера нужен metrics-server. В учебном окружении HPA-манифест демонстрирует готовность deployment к горизонтальному масштабированию.
