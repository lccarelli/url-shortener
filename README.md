# ✂️ URL Shortener

Acortador de URLs minimalista en **Go**, con almacenamiento en **Redis** y despliegue en **AWS Lambda** + **API Gateway**. Ideal para practicar:

- Arquitectura serverless (Lambda + Redis)
- Infraestructura como código (Terraform)
- Performance y carga con K6
- Observabilidad con Grafana (opcional)

---

## 🧱 Arquitectura

```
[ API Gateway ]
      ↓
[ AWS Lambda (Go) ]
      ↓
[ ElastiCache Redis ]
```

En local:
```
[ Docker Compose ]
→ shortener (Go API) + redis
```

---

## 📦 Funcionalidad

- 🧩 Acorta URLs con un hash determinista (`crc32 → base36`)
- 🔄 Recupera la URL original por clave corta
- 🧼 Elimina entradas
- ⏳ Claves expiran en 24hs

---

## 🚀 Endpoints

### `POST /shorten`

Acorta una URL

**Request:**
```json
{ "url": "https://example.com" }
```

**Response:**
```json
{ "short": "kf12oi" }
```

---

### `GET /{short}`

Redirecciona a la URL original

Ejemplo:
```
GET /kf12oi
→ 307 Temporary Redirect → https://example.com
```

---

### `GET /lookup/{short}`

Devuelve la URL original en formato JSON

**Response:**
```json
{ "url": "https://example.com" }
```

---

### `DELETE /{short}`

Elimina una entrada

---

## 🧪 Pruebas de carga (proyecto separado)

1. Usar [Grafana K6](https://k6.io/) en un proyecto aparte
2. Exportar métricas a InfluxDB
3. Visualizar resultados en un dashboard de Grafana

---

## 🐳 Uso local con Docker

```bash
make build     # compila imagen
make up        # levanta stack
make logs      # logs del servicio
make clean     # elimina recursos
```

---

## ☁️ Despliegue en AWS (Terraform)

Incluye:

- VPC + Subnet
- Redis ElastiCache
- Lambda (Go)
- API Gateway HTTP
- IAM roles mínimos

### 1. Empaquetar Lambda

```bash
make package
```

### 2. Deploy con Terraform

```bash
terraform init
terraform apply
```

> Al final verás `api_url = https://xxxx.execute-api...` para consumir.

---

## ⚙️ Variables

```hcl
variable "aws_region" {
  default = "us-east-1"
}
```

---

## 📥 Dependencias

- Go ≥ 1.21
- Docker + Compose
- Terraform ≥ 1.4
- AWS CLI con credenciales
- make

---

## 🧠 Notas técnicas

- El acortamiento es **idempotente**: misma URL → misma clave.
- Se usa `SETNX` para evitar sobrescribir claves.
- Las claves expiran en 24 horas (`EX 86400`).
- Redis policy: `allkeys-lru`, `maxclients=50000`.

---

## 📈 Extensiones posibles

- Auth y seguimiento por usuario
- Estadísticas de clicks por clave
- Slugs personalizados
- Persistencia permanente (sin expiración)
- Migrar a DynamoDB para full serverless

---

## 👩‍💻 Autora

Desarrollado por [Laura Carelli](https://github.com/lccarelli)

---
