# QA — go-node-integration

## Resultado

La revisión no encontró defectos funcionales bloqueantes. La implementación
mantiene la secuencia aprobada `Fiber handler -> QR core -> statistics HTTP
client`, no modifica `node-api/`, y las validaciones Go completaron
satisfactoriamente.

## Revisión funcional y de contrato

| Área | Resultado | Evidencia |
| --- | --- | --- |
| QR existente | Conforme | El handler llama a `qr.Factorize` antes de cualquier llamada downstream; las pruebas verifican ortonormalidad, reconstrucción, triangularidad e inputs inválidos. |
| Request Go válido y respuesta final | Conforme | `POST /qr` devuelve `200` con `q`, `r` y los cinco campos de `statistics`; la prueba de flujo verifica que `Q × R` reconstruye la matriz. |
| Contrato Go -> Node | Conforme | `statistics.Client` hace `POST /api/v1/statistics`, usa `Content-Type: application/json` y serializa exactamente `{q,r}`. Las pruebas usan `httptest`, sin Node real. |
| Parsing y forwarding | Conforme | Se exigen los cinco campos, numéricos finitos y booleano; JSON malformado, trailing, incompleto o de tipo incorrecto se rechaza como downstream. El resultado válido se reexpone bajo `statistics`. |
| Node no disponible y non-2xx | Conforme | Los errores de red y todo non-2xx, incluido `400`, se categorizan como `ErrDownstream` y el handler responde `502` con el cuerpo seguro aprobado. |
| Timeout | Conforme en implementación y mapping | `http.Client` compartido se configura en `main` con `5 * time.Second`; deadline/context timeout se categoriza como `ErrTimeout` y el handler responde `504`. Véase el hallazgo de cobertura P2. |
| Response bodies | Conforme | El cliente difiere `response.Body.Close()` y una prueba con body rastreable confirma el cierre. |
| Configuración y URLs | Conforme | La única configuración es `NODE_API_URL`, leída en `main`; `statistics.New` exige URL HTTP(S) absoluta, sin query/fragment. No hay host, localhost ni URL de producción hardcodeada en código Go. |
| Separación de responsabilidades | Conforme | Fiber se limita a parsear/mapear respuestas, `qr` no conoce HTTP/Node, y `statistics.Client` encapsula la red con la interfaz mínima consumida por el handler. |
| Regresiones | Conforme | Las pruebas QR, HTTP API y statistics pasan. No se modificó Node ni se añadieron dependencias. |

## Errores HTTP revisados

| Condición | Resultado público esperado | Estado |
| --- | --- | --- |
| Input QR inválido | `400 invalid_request` | Conservado y probado |
| Fallo QR inesperado | `500 internal_error` | Conservado en el handler |
| Node inaccesible, non-2xx o JSON inválido | `502 downstream_error` | Implementado y probado |
| Timeout/deadline downstream | `504 downstream_timeout` | Implementado y probado mediante fake/deadline |
| Éxito QR + Node | `200 {q,r,statistics}` | Implementado y probado |

## Resultados automatizados

Ejecutados desde `go-api/`:

| Comando | Resultado |
| --- | --- |
| `test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"` | Pass |
| `go vet ./...` | Pass |
| `go test ./...` | Pass (resultado inicialmente desde caché) |
| `go test -count=1 ./...` | Pass fuera del sandbox: `httpapi`, `qr` y `statistics` |
| `go build ./...` | Pass |

La primera ejecución sin caché dentro del sandbox no pudo abrir el listener
IPv6 local que crea `httptest` (`operation not permitted`). Se repitió el mismo
comando fuera del sandbox y pasó; es una limitación del sandbox, no una falla
de la suite ni del servicio. No ejecuté el comando de aplicación `gofmt -w`
porque QA no puede modificar source; el chequeo de formato confirma que no era
necesario. No aplica validación Node: no hubo cambios en `node-api/`.

## Hallazgos

### P2 — falta prueba directa del timeout del cliente HTTP configurado

`TestCalculateCategorizesDeadlineAsTimeout` verifica un contexto que ya venció
antes de iniciar la solicitud. Esto prueba la categorización de deadline, pero
no ejerce el timeout del `http.Client` configurado en producción contra un
downstream que demora. El mapping `504` del handler se prueba con un fake.

No es un defecto de la implementación: `isTimeout` reconoce tanto
`context.DeadlineExceeded` como `net.Error.Timeout`, y el cliente de producción
tiene el timeout aprobado de cinco segundos. Como mejora de cobertura, añadir
una prueba del cliente con servidor `httptest` demorado y timeout corto de
prueba confirmaría el camino real del transporte sin ralentizar la suite.

## Checklist de regresión

- [x] QR continúa siendo puro e independiente de HTTP.
- [x] Entrada `/qr` válida llega a QR y llama al endpoint Node correcto.
- [x] Request Node contiene sólo `q` y `r` y usa JSON.
- [x] Respuesta Node válida se valida, parsea y reenvía bajo `statistics`.
- [x] `q` y `r` siguen presentes y reconstruyen la matriz de entrada.
- [x] Input inválido conserva `400 invalid_request`.
- [x] Unavailable, non-2xx y JSON downstream inválido se convierten en `502`.
- [x] Deadline downstream se convierte en `504`.
- [x] Body downstream se cierra; no se exponen URL, body ni matrices en errores.
- [x] `NODE_API_URL` es obligatorio, no hay URLs hardcodeadas y el cliente está separado del handler/core.
- [x] Tests unitarios de cliente y HTTP de Go son aislados de Node real.
- [x] Suite Go, vet, formato y build pasan.

## Cobertura relevante

La suite cubre propiedades matemáticas QR (incluidos valores negativos,
decimales, rango deficiente y ceros), validación de entrada HTTP, serialización
del contrato Node, cierre de body, red inaccesible, non-2xx, JSON downstream
malformado/trailing/incompleto/con tipo erróneo/no finito, y las respuestas
`200`/`400`/`502`/`504`. La única cobertura prioritaria pendiente es el caso
P2 del timeout real del transporte.
