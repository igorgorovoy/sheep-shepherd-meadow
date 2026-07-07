# Demo deployments — multi-tier stacks

Приклади показують, як деплоїти «існуючі контейнери» у Shepherd.

> Повний посібник: [`docs/meadow-workflow.md`](../../docs/meadow-workflow.md) · [`docs/demo-deployments.md`](../../docs/demo-deployments.md)

## Важливо про платформу

| Середовище | Що працює |
|------------|-----------|
| **macOS** (host mode) | Образи `minimal`, `tinynginx` (локальний import). Не Docker Hub напряму. |
| **Linux** | `sheep pull nginx`, `postgres:16-alpine`, `wordpress:6` з OCI (Docker Hub / Meadow). |

Образ у маніфесті — це **ім'я локального образу Sheep**, не `docker run`. Спочатку:

```bash
export SHEEP_DATA_DIR=/path/to/.run/sheep   # або /var/lib/sheep на Linux
sheep pull postgres:16-alpine               # Linux
sheep bootstrap minimal                     # мінімальний rootfs
```

Потім `sheepctl apply -f ...`.

## Швидкий старт (macOS / demo UI)

```bash
export SHEPHERD_API=localhost:9876
export SHEEP_DATA_DIR="$(pwd)/.run/sheep"

./scripts/demo-mac.sh          # bootstrap + tinynginx + apply mac-demo/
./bin/sheepctl get pods
./bin/sheepctl get deployments
./bin/sheepctl get services
```

Відкрий dashboard: http://localhost:9876/ — побачиш postgres, wordpress, redis (симуляція) і tinynginx.

## Linux: справжні OCI-образи (WordPress + PostgreSQL)

```bash
export SHEEP_DATA_DIR=/var/lib/sheep
sudo sheep pull postgres:16-alpine
sudo sheep pull wordpress:6-apache

./scripts/demo-linux-oci.sh    # apply linux-oci/
```

Після старту перевір:

```bash
sheepctl get pods
sheepctl get services
# WordPress: service wordpress, Postgres: service postgres
```

`WORDPRESS_DB_HOST` у маніфесті вказує на **Service name** `postgres` — на Linux з мережею кластера pods бачать endpoints через Service controller.

## Структура

```
mac-demo/           # працює на Mac (minimal + tinynginx)
linux-oci/          # postgres + wordpress з Docker Hub (Linux)
```

## Окремі сервіси

| Файл | Опис |
|------|------|
| `mac-demo/deployment-postgres.json` | «БД» (симуляція на minimal) |
| `mac-demo/deployment-wordpress.json` | «WordPress» (симуляція) |
| `mac-demo/deployment-redis.json` | «Redis» (симуляція) |
| `mac-demo/deployment-tinynginx.json` | Справжній HTTP (tinynginx) |
| `mac-demo/service-*.json` | ClusterIP services |
| `linux-oci/deployment-postgres.json` | postgres:16-alpine |
| `linux-oci/deployment-wordpress.json` | wordpress:6-apache |

## Видалити demo

```bash
sheepctl delete deployment postgres-db
sheepctl delete deployment wordpress
sheepctl delete deployment redis-cache
sheepctl delete deployment tinynginx
sheepctl delete service postgres postgres wordpress redis tinynginx
```
