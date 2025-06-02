# ✂️ URL Shortener

Acortador de URLs minimalista en **Go**, con almacenamiento en **Redis** y despliegue en **AWS Lambda** + **API Gateway**

---

## 🧱 Arquitectura

- In progress

---

## 📦 Funcionalidad

- Acorta URLs con un hash determinista (`crc32 → base36`)
- Recupera la URL original por clave corta
- Elimina entradas
- Claves expiran en 24hs

---

## 🧱 Propuesta de diseño de API (OpenAPI-style)

```bash
Método  Endpoint	      Descripción
POST    /shorten        Crea una URL corta desde una URL larga
GET     /{short}	      Redirecciona a la URL larga (para navegador)
GET     /stats/{short}	Devuelve estadísticas de la URL corta
DELETE  /{short}	      Borra una URL corta
```
---

## 🧪 Pruebas de carga (proyecto separado)

1. Usar repositorio https://github.com/lccarelli/url-shortener-load-test/tree/main


## ☁️ Despliegue en AWS (Terraform)

### 1. Empaquetar Lambda

```bash
make package
```

### 2. Deploy con Terraform

```bash
terraform init
terraform apply
```

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

## 📈 Extensiones posibles
In progress


---
