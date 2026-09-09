# Development Plan - go-api-structure-refactor

## Summary

Refactorizar únicamente la organización interna de `go-api` para que cada
responsabilidad sea evidente, sin cambiar el comportamiento observable de la
API ni del contrato con Node:

```text
main (composition root)
  ├─ config: lee la configuración existente
  ├─ handler: transporte Fiber, DTOs, CORS y respuesta HTTP
  ├─ usecase: una operación pequeña que coordina QR + estadísticas
  ├─ qr: cálculo y validación puros
  └─ external/statistics: adaptador HTTP hacia Node
```

No se usará POO clásica: Go no requiere clases, herencia, factories ni un
contenedor de inyección de dependencias. Habrá `struct` solamente cuando reúne
datos o una dependencia real (`config.Config`, `handler.QRHandler`,
`usecase.Service` y `statistics.Client`), junto con funciones para el cálculo
puro. Tampoco se crearán entidades: `qr.Matrix` y `statistics.Result` son
valores de cálculo/transferencia, sin identidad persistente ni ciclo de vida
propio.

Se justifica `config` porque separa la lectura de entorno del arranque, y
`handler` porque contiene exclusivamente la frontera HTTP. Se justifica un
único paquete `usecase` porque el flujo "factorizar y obtener estadísticas" es
la operación de aplicación que hoy mezcla el handler; no se copiarán del
ejemplo los paquetes `model`, `testutil`, `payload` ni archivos separados para
la implementación de Householder. El algoritmo permanecerá cohesionado en
`qr`.

El contrato que se preserva es `POST /qr` con `{ "matrix": number[][] }`, la
respuesta `{ "q", "r", "statistics" }`, CORS/preflight actual y los errores
actuales: `400 invalid_request`, `502 downstream_error`, `504
downstream_timeout` y `500 internal_error`. El adaptador continuará llamando
`POST /api/v1/statistics` con `{ "q", "r" }` y validando la respuesta actual.

## Files to Create

- `go-api/config/config.go`: declara `Config` y `Load` para reunir `PORT`,
  `NODE_API_URL` y `CORS_ALLOWED_ORIGINS`; conserva el puerto por defecto
  `8080` y no agrega variables de entorno. El timeout existente de cinco
  segundos se conserva como valor de configuración interno, no como una nueva
  variable configurable.
- `go-api/config/config_test.go`: prueba valores por defecto y lectura de las
  tres variables existentes mediante `t.Setenv`.
- `go-api/handler/router.go`: crea la aplicación Fiber, conserva la
  configuración CORS existente y registra `POST /qr`.
- `go-api/handler/qr_handler.go`: define `QRHandler`, recibe y decodifica el
  request, invoca la operación, y mapea errores al contrato HTTP actual. La
  interfaz mínima para sustituir la operación en tests se declara junto al
  handler consumidor.
- `go-api/handler/dto.go`: contiene únicamente los DTO JSON privados de
  entrada, salida y error, para que no se confundan con los tipos QR ni con el
  contrato saliente a Node.
- `go-api/usecase/factorize.go`: define el resultado de la operación y un
  `Service` pequeño que ejecuta `qr.Factorize` y después el colaborador de
  estadísticas. Su interfaz `Calculate` tendrá un único método y pertenecerá a
  este consumidor. Si QR falla, no se invocará Node; los errores del cliente se
  propagarán sin cambiar su clasificación.
- `go-api/usecase/factorize_test.go`: prueba que la operación entrega Q/R y
  estadísticas, entrega al colaborador las matrices calculadas, no lo llama
  ante error QR y propaga fallos de estadísticas.

## Files to Modify

- `go-api/main.go`: convertirlo en el composition root mínimo: cargar
  `config`, crear un único `net/http.Client` con el timeout ya existente,
  construir `external/statistics.Client`, `usecase.Service` y el router, y
  escuchar en la dirección configurada. No contendrá lógica de lectura de
  entorno, QR, CORS ni HTTP saliente.
- `go-api/.env.example`: no se modificará salvo que la implementación detecte
  que ya no documenta exactamente las mismas tres variables. No se añadirá
  `STATISTICS_TIMEOUT_MS` en este alcance.
- `go-api/external/statistics/client.go` y
  `go-api/external/statistics/client_test.go`: aceptar como destino canónico el
  movimiento ya presente en el worktree. Se actualizarán solo los imports y
  nombres necesarios por la nueva ubicación, manteniendo la implementación
  basada en `net/http`, timeout, contexto, validación de URL y validación del
  contrato Node.
- `go-api/httpapi/app.go`: mover a `go-api/handler/router.go` y adaptar su
  dependencia al handler; no dejar un segundo router.
- `go-api/httpapi/qr_handler.go`: mover y dividir de forma mínima entre
  `handler/qr_handler.go` y `handler/dto.go`; la coordinación QR + downstream
  se traslada a `usecase`.
- `go-api/httpapi/qr_handler_test.go`: mover a
  `go-api/handler/qr_handler_test.go`, conservando pruebas HTTP de JSON,
  estados y CORS mediante un falso de la operación. Las aserciones de
  orquestación pasan al test de `usecase`, evitando duplicar pruebas del
  cliente externo.
- `go-api/statistics/client.go` y `go-api/statistics/client_test.go`: no se
  restaurarán. Se completará su renombre a `go-api/external/statistics/` con
  `git mv`/staging equivalente, después de confirmar que el contenido no fue
  alterado accidentalmente.

`go-api/qr/qr.go` y `go-api/qr/qr_test.go` no se moverán ni reescribirán:
siguen siendo el paquete puro y testeable sin Fiber. No se crearán `model/`,
`entity/`, `repository/`, `domain/`, `factory/`, `testutil/` ni capas de Clean
Architecture.

## Execution Order

1. Antes de editar, inspeccionar y preservar el worktree: los dos archivos
   eliminados bajo `go-api/statistics/` y los dos no rastreados bajo
   `go-api/external/statistics/` son byte a byte equivalentes en el estado
   revisado. Tratarlo como un renombre pendiente, no como una eliminación que
   deba revertirse ni como código nuevo que deba sobrescribirse.
2. Crear `config` y sus pruebas con exactamente las variables y defaults ya
   existentes. Mantener la validación de URL en `statistics.New`, donde ya se
   aplica, para no duplicarla.
3. Crear `usecase` y sus pruebas. Extraer de la función HTTP solo el flujo
   secuencial QR → estadísticas; mantener `qr` como implementación matemática
   y `external/statistics` como frontera de red.
4. Mover `httpapi` a `handler`, separar router/DTO/handler y trasladar sus
   pruebas. Mantener los nombres JSON, los mensajes, los códigos, la ruta y
   la configuración CORS exactos.
5. Completar el renombre del cliente a `external/statistics`, actualizar
   imports en `main`, `handler` y `usecase`, y dejar de referenciar el paquete
   antiguo. No cambiar a Fiber client ni añadir dependencias: `net/http` ya
   satisface la necesidad.
6. Simplificar `main` a composición y arranque, usando las dependencias ya
   construidas una vez por proceso. El constructor del servicio rechazará un
   colaborador nulo para que el handler no tenga que defenderse de una
   dependencia faltante durante una petición.
7. Aplicar `gofmt`, ejecutar las validaciones y revisar el diff final para
   comprobar que los movimientos se registran como renombres y que no se
   incluyeron cambios fuera de `go-api` y este artefacto.

## Architecture Decisions

- **POO y entidades:** no se aplican. La composición de structs pequeños es
  idiomática en Go; no habrá clases ni herencia. Matriz y estadísticas son
  tipos de valor, no entidades DDD/persistentes.
- **`handler`: sí.** Es el único paquete que importa Fiber y conoce JSON,
  CORS, rutas y códigos HTTP. No calcula QR ni realiza HTTP saliente.
- **`config`: sí, pero pequeño.** Centraliza las variables existentes y sus
  defaults. No incorpora una capa de configuración genérica ni una nueva
  variable de timeout, porque eso cambiaría la superficie operativa sin ser
  necesario para la refactorización.
- **Una sola orquestación:** `usecase.Service` representa el flujo existente
  de una petición, no una arquitectura genérica. Su única interfaz es el
  puerto de estadísticas, necesaria para aislar la red en pruebas. No se
  definen interfaces para QR, config, router ni cada struct.
- **Tipos compartidos mínimos:** se conservan `qr.Matrix`,
  `statistics.Result`, `statistics.ErrDownstream` y `statistics.ErrTimeout`.
  Duplicarlos en un `model` solo para imitar el ejemplo añadiría una capa sin
  necesidad actual. El handler puede mapear los sentinels del adaptador al
  contrato HTTP existente; el use case los propaga intactos.
- **Cliente externo:** se mantiene `net/http` y una instancia compartida con
  timeout, contexto de la petición, cierre de body y validación defensiva de
  respuestas. El prefijo `external/` comunica que es un adaptador saliente,
  no una entidad de negocio.
- **Compatibilidad:** esta tarea es de estructura; no cambia QR económico,
  precisión/tolerancias, payloads, respuestas, CORS, Docker, Node, módulos ni
  dependencias.

## Validation Criteria

- Las pruebas de `qr` siguen demostrando las propiedades numéricas actuales:
  reconstrucción aproximada `Q × R`, ortonormalidad, triangularidad, entradas
  inválidas y que la entrada no se muta.
- Las pruebas de `usecase` demuestran la secuencia QR → estadísticas y la
  propagación de errores sin abrir un servidor HTTP.
- Las pruebas de `handler` conservan `POST /qr`, cuerpo inválido/ausente,
  matrices inválidas, `400/502/504/500`, JSON de respuesta y preflight/CORS.
- Las pruebas de `external/statistics` conservan método, ruta,
  `Content-Type`, payload `{q,r}`, campos obligatorios de respuesta,
  respuestas no 2xx/malformadas/no finitas, timeout, red y cierre de body.
- Desde `go-api/`, ejecutar en este orden:

  ```bash
  find . -type f -name '*.go' -exec gofmt -w {} +
  test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
  go vet ./...
  go test ./...
  go build ./...
  ```

  Si el sandbox vuelve a restringir listeners de `httptest`, documentar el
  bloqueo y ejecutar la misma suite en un entorno donde se permitan sockets;
  no se eliminarán esas pruebas para obtener un resultado verde artificial.

## External Dependencies

No se añaden dependencias ni se modifica `go.mod`/`go.sum`.

Se mantienen Fiber para el servidor y el cliente estándar `net/http` para la
llamada al Node API. La URL del downstream continúa llegando mediante
`NODE_API_URL`; no se introducen secretos, archivos `.env` ni URLs codificadas.

## Blocked Tasks

No hay bloqueo para aprobar este diseño. La decisión de alcance tomada es
conservar el timeout fijo actual de cinco segundos en vez de añadir
`STATISTICS_TIMEOUT_MS`; si se desea configurarlo por entorno, debe ser una
tarea/decisión explícita posterior porque cambia la configuración documentada.

La implementación debe tratar el worktree sucio descrito arriba como un
renombre pendiente y verificar de nuevo su equivalencia antes de hacer staging.
No debe ejecutar `git checkout`, `git reset` ni restaurar los archivos bajo
`go-api/statistics/`, pues podría destruir trabajo preexistente del usuario.

## Execution Results

Completed files:

- Created `go-api/config/config.go` and `config_test.go`; `Load` preserves the
  existing three environment variables, default port `8080`, and the fixed
  five-second timeout without adding a variable.
- Created `go-api/usecase/factorize.go` and `factorize_test.go`; the service
  runs QR before the statistics collaborator, skips it when QR fails, and
  propagates downstream errors.
- Moved the Fiber transport from `httpapi` into `go-api/handler/` as
  `router.go`, `qr_handler.go`, `dto.go`, and `qr_handler_test.go`. It retains
  the route, JSON payloads, CORS behavior, and `400`, `502`, `504`, and `500`
  mappings.
- Updated `go-api/main.go` as the composition root for config, one HTTP
  client, the external statistics client, the use case, and the router.
- Completed the pending canonical move from `go-api/statistics/` to
  `go-api/external/statistics/`. Before implementation, SHA-256 checks
  confirmed both moved files were byte-for-byte identical to their `HEAD`
  counterparts; their implementation and tests were not changed.

No planned file is blocked. `go-api/.env.example`, `go-api/qr/qr.go`,
`go-api/qr/qr_test.go`, `go.mod`, and `go.sum` were intentionally unchanged.

Validation from `go-api/`:

```text
find . -type f -name '*.go' -exec gofmt -w {} +              PASS
test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)" PASS
go vet ./...                                                   PASS
go test ./...                                                  PASS
go build ./...                                                 PASS
```

The first sandboxed `go test ./...` attempt could not open the localhost
listener required by `httptest` in `external/statistics`; the unchanged tests
were retained and the same suite passed when run with permission for local
sockets. No out-of-scope changes or follow-up architecture suggestions were
identified.
