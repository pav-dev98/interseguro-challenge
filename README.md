# Interseguro Coding Challenge

Monorepo con dos APIs que reciben una matriz, calculan su factorización QR reducida y obtienen estadísticas de las matrices resultantes.

## Arquitectura

```mermaid
flowchart LR
    C[Cliente] --> G[Go API /qr]
    G --> Q[QR: Modified Gram-Schmidt]
    Q --> N[Node API /statistics]
    N --> S[Estadísticas de Q y R]
    S --> G
    G --> R[Respuesta con Q, R y statistics]
```

La Go API calcula `Q` y `R`, y envía ambas matrices por HTTP a la Node API usando `NODE_API_URL`. Node devuelve las estadísticas y Go construye la respuesta final.

## Estructura

```text
interseguro-challenge/
├── go-api/
│   ├── nodeclient/        # Cliente HTTP para Node API
│   ├── qr/                # QR y pruebas matemáticas
│   ├── Dockerfile
│   ├── main.go
│   └── main_test.go
├── node-api/
│   ├── src/
│   │   ├── statistics/    # Cálculo de estadísticas
│   │   ├── app.ts
│   │   └── index.ts
│   ├── test/
│   ├── Dockerfile
│   └── package.json
├── frontend/              # Scaffold inicial; sin funcionalidad del challenge
├── docker-compose.yml
└── README.md
```

## Tecnologías

- Go 1.22, Fiber v2 y `net/http` de la biblioteca estándar.
- Node.js, TypeScript, Express 5 y `tsx` para desarrollo.
- Node.js built-in test runner y `testing` de Go.
- Docker y Docker Compose.
- Next.js, React y TypeScript están presentes únicamente como scaffold inicial del frontend.

## Ejecución local

La forma recomendada de iniciar ambas APIs es Docker Compose:

```bash
docker compose up --build
```

Con la configuración actual, los servicios quedan disponibles en:

- Go API: `http://localhost:3001`
- Node API: `http://localhost:3002`

Para detenerlos:

```bash
docker compose down
```

También pueden ejecutarse individualmente. Primero inicie Node y después Go, configurando la URL interna para el entorno local:

```bash
cd node-api
npm install
npm run dev
```

```bash
cd go-api
NODE_API_URL=http://localhost:3002 go run .
```

## Variables de entorno

| Variable | Servicio | Uso | Valor local por defecto |
| --- | --- | --- | --- |
| `PORT` | Go y Node | Puerto de escucha asignado por el entorno cloud. | Go: `3001`; Node: `3002` |
| `NODE_API_URL` | Go | Base URL de Node API para `GET /health/dependencies` y `POST /qr`. | En Compose: `http://node-api:3002` |

`PORT` es opcional: Go usa `3001` cuando está vacío y Node usa `3002` cuando no recibe un puerto positivo válido. Para que `POST /qr` complete el flujo, `NODE_API_URL` debe estar configurada en Go.

## Endpoints

### Go API

| Método | Ruta | Descripción |
| --- | --- | --- |
| `GET` | `/health` | Estado básico de Go API. |
| `GET` | `/health/dependencies` | Verifica por HTTP la disponibilidad de Node API. Endpoint operativo/de dependencias. |
| `POST` | `/qr` | Calcula QR, solicita estadísticas a Node y devuelve el resultado combinado. |

`GET /health` responde:

```json
{"status":"active"}
```

`GET /health/dependencies` responde, cuando Node está disponible:

```json
{
  "status": "active",
  "dependencies": {
    "node-api": "active"
  }
}
```

### Node API

| Método | Ruta | Descripción |
| --- | --- | --- |
| `GET` | `/health` | Estado básico de Node API. |
| `GET` | `/internal/health` | Health check interno que identifica el servicio. |
| `POST` | `/statistics` | Calcula estadísticas de las matrices `q` y `r`. Lo consume Go API. |

`GET /health` responde:

```json
{"status":"active"}
```

`GET /internal/health` es un endpoint interno y responde:

```json
{"status":"active","service":"node-api"}
```

`POST /statistics` recibe:

```json
{
  "q": [[1, 0], [0, 1]],
  "r": [[2, 3], [0, 4]]
}
```

y responde:

```json
{
  "max": 4,
  "min": 0,
  "sum": 11,
  "average": 1.375,
  "hasDiagonalMatrix": true
}
```

## Ejemplo completo: `POST /qr`

```bash
curl -X POST http://localhost:3001/qr \
  -H 'Content-Type: application/json' \
  -d '{
    "matrix": [
      [1, 1],
      [1, 0]
    ]
  }'
```

Respuesta representativa:

```json
{
  "q": [
    [0.7071067811865475, 0.7071067811865477],
    [0.7071067811865475, -0.7071067811865474]
  ],
  "r": [
    [1.4142135623730951, 0.7071067811865475],
    [0, 0.7071067811865476]
  ],
  "statistics": {
    "max": 1.4142135623730951,
    "min": -0.7071067811865474,
    "sum": 4.242640687119286,
    "average": 0.5303300858899107,
    "hasDiagonalMatrix": false
  }
}
```

Los últimos decimales pueden variar por aritmética de punto flotante.

## Factorización QR

Go implementa desde cero una factorización QR reducida mediante **Modified Gram-Schmidt**. Para cada columna, elimina las proyecciones sobre las columnas previas de `Q`; las proyecciones se almacenan en `R` y el residual normalizado forma la siguiente columna de `Q`.

La entrada debe ser una matriz rectangular `m × n` con `m >= n`, sin filas vacías, valores no finitos ni filas de distinto tamaño. La salida usa `Q` de dimensión `m × n` y `R` de dimensión `n × n`.

La tolerancia es `1e-12`, escalada por `max(1, norma original de la columna)`. Un residual con norma menor o igual a ese umbral se trata como columna linealmente dependiente o numéricamente casi dependiente.

## Estadísticas

Node valida `q` y `r` como matrices no vacías, rectangulares y con números finitos. Recorre todos los valores de ambas matrices en una sola pasada para calcular:

- `max`: mayor valor de `Q ∪ R`.
- `min`: menor valor de `Q ∪ R`.
- `sum`: suma de todos los valores de `Q` y `R`.
- `average`: `sum / cantidad total de valores`.

`hasDiagonalMatrix` es `true` si `Q` o `R` es cuadrada y todos sus valores fuera de la diagonal principal tienen valor absoluto menor o igual a `1e-12`. Una matriz no cuadrada nunca se considera diagonal.

## Manejo de errores

| Servicio/ruta | Estado | Caso |
| --- | --- | --- |
| Go `POST /qr` | `400` | JSON inválido, `matrix` ausente, matriz vacía/irregular, fila vacía, valores no finitos o `m < n`. |
| Go `POST /qr` | `422` | Columnas dependientes o casi dependientes según la tolerancia. |
| Go `POST /qr` | `503` | `NODE_API_URL` no configurada, Node no disponible o timeout. |
| Go `POST /qr` | `502` | Node devuelve un status no exitoso, JSON inválido o respuesta de estadísticas incompleta/no finita. |
| Go `GET /health/dependencies` | `503` | `NODE_API_URL` no configurada, Node no disponible, timeout o status no exitoso. |
| Node `POST /statistics` | `400` | JSON inválido, `q`/`r` ausente, matriz vacía/irregular, fila vacía o valores no finitos. |

## Tests

### Go API

```bash
cd go-api
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Las pruebas cubren QR 2×2 y rectangular, reconstrucción `Q × R`, ortonormalidad de `Q`, triangularidad de `R`, validaciones de matriz, cliente HTTP a Node, health de dependencias y los errores de integración de `POST /qr` con servidores `httptest`.

### Node API

```bash
cd node-api
npm install
npm test
npm run build
```

Las pruebas cubren health interno, cálculo y validación de estadísticas, detección de matrices diagonales y solicitudes HTTP válidas e inválidas a `/statistics`.

## Docker y despliegue

Los dos servicios están contenerizados. La Go API utiliza una imagen multi-stage con `scratch` como imagen final, manteniendo una imagen mínima y los certificados CA necesarios para realizar conexiones HTTPS salientes.

Los servicios están desplegados en Render:

- Go API: <https://interseguro-go-api-d9s6.onrender.com>
- Node API: <https://interseguro-node-api-lr3j.onrender.com>

### Prueba rápida

```bash
curl https://interseguro-go-api-d9s6.onrender.com/health
curl https://interseguro-go-api-d9s6.onrender.com/health/dependencies
```

El segundo endpoint permite comprobar que Go API puede comunicarse correctamente con Node API.

## Decisiones técnicas y supuestos

- El enunciado menciona «rotación» en algunos apartados, pero la funcionalidad requerida especifica factorización QR; esta implementación toma QR como el requisito concreto.
- Las estadísticas se calculan sobre todos los valores de `Q` y `R` juntos; la comprobación diagonal se realiza individualmente sobre cada matriz.
- Se implementa QR reducido para matrices rectangulares con `m >= n`; las columnas dependientes se rechazan en lugar de producir una base incompleta.

## Opcionales no implementados

- Frontend: pendiente; actualmente solo existe el scaffold inicial de Next.js.
- JWT: no implementado actualmente.
