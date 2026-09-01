# Interseguro Challenge

Monorepo base para el coding challenge de Interseguro.

## Aplicaciones

- `go-api`: API en Go con Fiber.
- `node-api`: API en Node.js, TypeScript y Express.
- `frontend`: aplicación inicial de Next.js con TypeScript.

## Desarrollo local

Instala las dependencias de cada proyecto antes de iniciarlo.

```bash
cd go-api && go run .
cd node-api && npm run dev
cd frontend && npm run dev
```

Las APIs exponen `GET /health` y usan los puertos 3001 (Go) y 3002 (Node). El frontend utiliza el puerto 3000 por defecto.
