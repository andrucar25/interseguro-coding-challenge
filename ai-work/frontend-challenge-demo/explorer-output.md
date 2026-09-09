# Exploración de alcance: `frontend-challenge-demo`

## Qué se construirá

Una demo frontend de una sola página, pequeña y estática, para introducir una matriz, validarla localmente, invocar únicamente el endpoint público de Go y presentar la factorización QR y sus estadísticas. El alcance solicitado está limitado a `frontend/**`; no incluye cambios en las APIs ni despliegue.

## Actores

- Usuario de la demo: escribe o carga una matriz, corrige errores y ejecuta el cálculo.
- Navegador: parsea/valida el texto, hace el `fetch` y renderiza loading, resultados o errores.
- Go/Fiber: API pública que valida/factoriza y obtiene estadísticas de Node.
- Node/Express: dependencia indirecta; el navegador no debe conocerla ni llamarla.

## Flujo principal verificado

1. El usuario introduce filas en un textarea o carga un ejemplo.
2. El frontend convierte las filas a `number[][]` y valida las reglas que también exige Go.
3. Envía `POST {VITE_API_URL}/qr` con `Content-Type: application/json` y body `{ "matrix": matrix }`.
4. Mientras espera, deshabilita el submit y muestra `Calculating...`.
5. En HTTP 2xx interpreta el resultado y muestra Q, R y estadísticas.
6. En validación local, fallo de red, error HTTP o respuesta inesperada muestra un mensaje seguro junto al formulario/resultados.

## Contrato real de Go

Fuente inspeccionada: `go-api/httpapi/app.go`, `go-api/httpapi/qr_handler.go`, `go-api/qr/qr.go`, `go-api/httpapi/qr_handler_test.go`.

### Request

- Método: `POST`.
- Ruta: `/qr`.
- Header esperado: `Content-Type: application/json`.
- Body: `{ "matrix": number[][] }`.
- La matriz se representa como filas de números JSON (`float64` en Go).

El frontend debe llamar a la base configurada en `import.meta.env.VITE_API_URL` y añadir `/qr` de forma segura. No debe llamar a `/api/v1/statistics` ni a Node.

### Success response

El handler responde 200 con:

```json
{
  "q": [[0.0]],
  "r": [[1.0]],
  "statistics": {
    "maximum": 1.0,
    "minimum": 0.0,
    "sum": 1.0,
    "average": 0.5,
    "hasDiagonalMatrix": true
  }
}
```

Los nombres reales de estadísticas son `maximum`, `minimum`, `sum`, `average` y `hasDiagonalMatrix`. Q es una matriz `m x n` y R es `n x n` para una entrada `m x n`; Go implementa QR económico.

### Error response y status codes

Los errores JSON tienen siempre esta forma cuando los produce el handler:

```json
{ "error": "...", "message": "..." }
```

- `400 invalid_request`: JSON/body inválido, matriz vacía, fila vacía, filas no rectangulares, matriz más ancha que alta (`rows < columns`) o valores no finitos. Para JSON o valores no numéricos el mensaje puede ser `request body must contain a valid matrix`; para errores matemáticos son los mensajes de `qr`.
- `502 downstream_error`: Go no pudo obtener estadísticas de Node.
- `504 downstream_timeout`: timeout al obtener estadísticas.
- `500 internal_error`: error interno/no clasificado.
- El frontend debe tratar cualquier otro status o JSON incompatible como respuesta inesperada y no mostrar objetos crudos ni stack traces.

El cliente HTTP de Go valida que las estadísticas existan, sean numéricas/booleanas y finitas, por lo que el frontend puede tipar y validar mínimamente la forma de la respuesta antes de renderizarla.

## Restricciones de QR que debe reflejar la validación local

Confirmadas en `go-api/qr/qr.go`:

- al menos una fila;
- cada fila con al menos una columna;
- todas las filas con la misma cantidad de columnas;
- cada valor finito (no NaN ni infinito);
- cantidad de filas mayor o igual a cantidad de columnas.

No hay límite máximo de filas/columnas configurado en el código inspeccionado. No se debe inventar un límite de tamaño en la demo; solo podrían existir límites operativos del navegador/Cloud Run fuera de este contrato.

La entrada de ejemplo `12 -51 4 / 6 167 -68 / -4 24 -41` es compatible: es una matriz 3 x 3 con valores finitos.

## Formato de input recomendado para planificar

Usar un textarea donde cada línea sea una fila. Aceptar separadores por espacios y/o comas, por ejemplo `1 2` y `1, 2`; normalizar comas a espacios antes de tokenizar. Ignorar líneas completamente vacías solo si se decide explícitamente que son separación accidental; para detectar errores útiles, una línea vacía entre filas debería producir un mensaje claro o ser tratada consistentemente. La alternativa más simple y predecible es rechazar filas vacías no finales y permitir que un textarea compuesto solo de whitespace produzca “Enter at least one matrix row.”

El parser debe conservar los signos, decimales y cero. Debe aceptar la sintaxis numérica que representa `Number` finito (incluyendo notación exponencial si el planner la mantiene), y rechazar tokens vacíos producidos por separadores ambiguos o tokens no numéricos con la fila indicada. No convertir valores a strings ni redondearlos antes del request.

## Casos límite y estados

- textarea vacío o solo whitespace;
- filas vacías accidentales;
- coma/separador que deja un token inválido;
- token no numérico;
- decimal, negativo, cero y notación científica válidos si son finitos;
- fila con distinta longitud;
- matriz `rows < columns`;
- request duplicado durante loading;
- `VITE_API_URL` ausente o con una URL mal formada (error de configuración visible, sin request);
- timeout/fallo de conexión/CORS, que el navegador expone generalmente como error de red sin detalles seguros;
- respuestas 400, 500, 502 y 504;
- status 2xx con JSON inválido o campos ausentes/forma de matriz inesperada;
- Q/R anchas, que deben poder desplazarse horizontalmente en móvil;
- valores QR muy pequeños o con muchos decimales: presentación redondeada únicamente para lectura, conservando el valor recibido en state.

## Dependencias y configuración

Stack obligatorio: React + Vite + TypeScript estricto. La opción de menor riesgo es conservar las dependencias estándar del template Vite y usar `fetch`, `useState` y CSS propio. No hay necesidad demostrada de router, state manager, cliente HTTP, tabla, iconos, UI kit o skeletons.

Crear `frontend/.env.example` con `VITE_API_URL=`. La URL no debe hardcodearse ni contener credenciales. El build debe poder apuntar localmente o a Go Cloud Run cambiando solo variables de build.

## CORS: hallazgo y bloqueo

`go-api/httpapi/app.go` crea Fiber y registra únicamente `POST /qr`; no se encontró middleware `cors`, manejo de `OPTIONS` ni headers `Access-Control-Allow-Origin`. El request JSON cross-origin desde `http://localhost:5173` normalmente dispara preflight, por lo que el flujo navegador → Go Cloud Run no está garantizado y probablemente falle por CORS aunque el endpoint funcione.

No modificar `go-api` en esta feature. Registrar una feature backend separada `frontend-cors` para permitir explícitamente los orígenes necesarios y responder preflight; su diseño debe definir local y Firebase Hosting. Hasta entonces, el frontend puede implementarse/buildarse, pero la validación manual end-to-end desde el navegador queda bloqueada por CORS.

## Fuera de alcance

- Cambios en `go-api`, `node-api`, infraestructura, Docker Compose, Terraform o recursos Cloud Run/Firebase.
- Llamadas directas del navegador a Node.
- Firebase Hosting o Firebase CLI.
- Autenticación, navegación, backend propio, API routes, persistencia, historial, charts o edición tipo spreadsheet.
- Nuevas reglas matemáticas no presentes en Go, como límite arbitrario de tamaño o restricciones de signo.
- Rounding de datos enviados o recibidos.

## Dependencias externas/bloqueadores

- Es necesario conocer `VITE_API_URL` en desarrollo/build.
- Go debe estar desplegado y accesible por HTTPS para probar la demo real.
- CORS en Go requiere trabajo separado antes de que `localhost:5173 → Go Cloud Run` funcione en un navegador.
- Si la API se expone bajo un prefijo o URL con trailing slash, el planner debe definir cómo unirlo con `/qr` sin producir una ruta incorrecta; el contrato actual del código solo confirma `/qr` relativo a la base.

## Preguntas resueltas y pendientes para planner/humano

Resueltas por inspección: endpoint `/qr`; request `{matrix}`; response `{q,r,statistics}`; nombres exactos de estadísticas; status/error shape; condición `rows >= columns`; ausencia de límite máximo; ausencia de CORS en el Go actual.

Pendientes de decisión de diseño, no de contrato backend: si se permiten líneas vacías internas o se rechazan; si se acepta notación científica (recomendado, siempre que `Number.isFinite`); número visual de decimales y tratamiento visual de casi-cero; mensaje exacto para URL de configuración faltante; estrategia concreta para CORS como trabajo posterior.

