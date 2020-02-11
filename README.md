

# Selectel Status Exporter

Прометеус экспортер для получения кол-ва средств на счете аккаунта Selectel.

## Как работает экспортер

Экспортер раз в час ходит по url `http://selectel.status.io/1.0/status/5980813dd537a2a7050004bd` получает в json формате инфу по статусу инфраструктуры датацентра  и отдает ее по url `/metrics` в формате прометеуса.


## Как запустить

Создаем `docker-compose.yml` файл:

```
version: '3'

services:
  exporter:
    build: .
    image: mxssl/selectel_Status_exporter
    ports:
      - "6789:80"
    restart: always
```

Далее запускаем экспортер:

```
docker-compose up -d
```

Проверить работу экспортера можно следующими командами:

```
docker-compose ps
docker-compose logs
```

Метрики доступны по url `your_ip:6789/metrics`

## Настройка для prometheus:

```
  - job_name: 'selectel_Status'
    scrape_interval: 60m
    static_configs:
      - targets: ['exporter_ip:6789']
```

## Дашборд для графаны:
