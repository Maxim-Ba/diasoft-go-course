# Документация по метрикам Calendar

## Обзор

Сервис экспортирует метрики в формате Prometheus на эндпоинте `GET /metrics`.
Метрики собираются автоматически через middleware и инструментирование бизнес-логики.

---

## HTTP-метрики

### `http_requests_total`
**Тип:** Counter  
**Labels:** `method`, `path`, `status`  
**Описание:** Общее количество HTTP-запросов к API.  
Позволяет отслеживать нагрузку на каждый эндпоинт, долю ошибок (4xx/5xx) и распределение трафика.

---

### `http_request_duration_seconds`
**Тип:** Histogram  
**Labels:** `method`, `path`  
**Buckets:** стандартные Prometheus (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)  
**Описание:** Распределение времени обработки HTTP-запросов в секундах.  
Позволяет выявить узкие места (медленные эндпоинты), отслеживать деградацию производительности.

---

## Бизнес-метрики событий

### `calendar_events_created_total`
**Тип:** Counter  
**Описание:** Общее количество успешно созданных событий.  
Показывает активность пользователей и темп роста данных.

### `calendar_events_updated_total`
**Тип:** Counter  
**Описание:** Общее количество успешно обновлённых событий.

### `calendar_events_deleted_total`
**Тип:** Counter  
**Описание:** Общее количество успешно удалённых событий.

---

### `calendar_events_fetched_total`
**Тип:** Counter  
**Labels:** `scope` (`day` | `week` | `month`)  
**Описание:** Общее количество запросов на получение списков событий с разбивкой по временному диапазону.

Помогает понять, какой вид выборки наиболее востребован.

---

## Метрики планировщика

### `calendar_scheduler_notifications_sent_total`
**Тип:** Counter  
**Описание:** Общее количество успешно отправленных уведомлений о событиях через Kafka.

### `calendar_scheduler_notifications_errors_total`
**Тип:** Counter  
**Описание:** Количество ошибок при отправке уведомлений (сбои Kafka, ошибки сериализации).  
Критичная метрика — резкий рост сигнализирует о проблемах с очередью сообщений.

---

### `calendar_scheduler_cleanups_total`
**Тип:** Counter  
**Описание:** Количество старых событий (>1 года), удалённых планировщиком при очистке.

### `calendar_scheduler_last_run_timestamp_seconds`
**Тип:** Gauge  
**Описание:** Unix timestamp последнего успешного запуска `ScanAndNotify`.  
Позволяет выявить зависание планировщика — если значение не обновляется, задача не выполняется.

---

## Инфраструктура мониторинга

### Запуск стека
```bash
docker-compose -f deployments/docker-compose.yaml up -d
```

| Сервис     | URL                        | Описание                     |
|------------|----------------------------|------------------------------|
| Calendar   | http://localhost:8082       | REST API                     |
| Metrics    | http://localhost:8082/metrics | Prometheus scrape endpoint |
| Prometheus | http://localhost:9090       | Сбор и хранение метрик       |
| Grafana    | http://localhost:3000       | Визуализация (admin/admin)   |

### Настройка Grafana
1. Открыть http://localhost:3000, войти с `admin` / `admin`.
2. **Configuration → Data Sources → Add data source → Prometheus**.
3. URL: `http://prometheus:9090`, нажать **Save & Test**.

