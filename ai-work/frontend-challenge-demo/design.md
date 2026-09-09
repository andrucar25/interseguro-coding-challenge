# Development Plan - frontend-challenge-demo

## Summary

Crear una demo estática de una sola página en `frontend/`, con React, Vite y
TypeScript estricto. La página permitirá introducir una matriz por texto,
validarla localmente, invocar solamente la API pública de Go y mostrar su
resultado QR económico y las estadísticas devueltas. No habrá navegación,
estado global, librerías de UI, cliente HTTP adicional, backend propio ni
configuración de Firebase.

El alcance de implementación queda limitado a `frontend/**`. Este documento
no modifica `go-api`, `node-api`, infraestructura ni recursos cloud.

### Contrato Go verificado

Inspeccionado en `go-api/httpapi/app.go`, `go-api/httpapi/qr_handler.go`,
`go-api/qr/qr.go` y sus pruebas:

| Caso | Contrato real |
| --- | --- |
| Método y ruta | `POST /qr` |
| Header | `Content-Type: application/json` |
| Request | `{ "matrix": number[][] }` |
| Éxito | `200` con `{ "q": number[][], "r": number[][], "statistics": { "maximum": number, "minimum": number, "sum": number, "average": number, "hasDiagonalMatrix": boolean } }` |
| Errores del handler | `{ "error": string, "message": string }` |
| Validación Go | `400 invalid_request`: JSON/matriz inválida, vacía, fila vacía, no rectangular, valores no finitos o más columnas que filas |
| Dependencia Node | `502 downstream_error` con `unable to obtain statistics`; `504 downstream_timeout` con `statistics service timed out` |
| Error interno | `500 internal_error` con `unable to factorize matrix` |

Para una entrada `m × n` válida, Go produce QR económico: Q tiene `m × n` y
R tiene `n × n`. No hay límite máximo de tamaño definido en el contrato. El
frontend no llamará a `/api/v1/statistics` ni conocerá la URL de Node.

## Files to Create

Todos los siguientes archivos serán creados bajo `frontend/` al aprobar e
implementar el diseño:

| Archivo | Responsabilidad |
| --- | --- |
| `frontend/package.json` | Scripts estándar de Vite y dependencias de React, Vite, TypeScript y ESLint del template. |
| `frontend/package-lock.json` | Lockfile generado por npm para instalaciones reproducibles. |
| `frontend/index.html` | Punto de montaje y metadatos mínimos de la página Vite. |
| `frontend/vite.config.ts` | Configuración mínima generada por el template React + TypeScript. |
| `frontend/tsconfig.json` | Referencia de proyectos TypeScript del template. |
| `frontend/tsconfig.app.json` | TypeScript estricto de la aplicación. |
| `frontend/tsconfig.node.json` | Tipado de la configuración Vite. |
| `frontend/eslint.config.js` | Única estrategia de lint, mantenida desde el template Vite; no se agregará Biome ni Prettier. |
| `frontend/.gitignore` | Exclusiones estándar de Vite, incluido `.env`; `.env.example` seguirá versionado. |
| `frontend/.env.example` | Solo `VITE_API_URL=`. No contiene URL desplegada ni secretos. |
| `frontend/src/main.tsx` | Montaje React y carga de estilos globales. |
| `frontend/src/App.tsx` | Orquestación de estado, submit y layout semántico de la única página. |
| `frontend/src/types.ts` | Tipos concretos del contrato Go (`QrResponse`, estadísticas y error seguro). |
| `frontend/src/api/calculateQr.ts` | Una función pequeña de `fetch` a Go, unión segura de la base y `/qr`, lectura segura de respuestas y clasificación de fallos. |
| `frontend/src/lib/parseMatrix.ts` | Función pura de parseo y validación local que devuelve `number[][]` o un único mensaje accionable. |
| `frontend/src/components/MatrixInput.tsx` | `label`, ayuda de sintaxis, textarea, botón de ejemplo, error local y controles de envío. |
| `frontend/src/components/MatrixTable.tsx` | Tabla accesible y desplazable para renderizar Q o R sin transformar los datos. |
| `frontend/src/components/Statistics.tsx` | Grid compacto de las cinco estadísticas reales del backend. |
| `frontend/src/styles.css` | Estilos propios, responsive y de foco; no design system ni CSS-in-JS. |

No se creará `frontend/README.md`: la aplicación y sus scripts son lo bastante
pequeños y el README raíz permanece fuera del alcance.

## Files to Modify

Ningún archivo existente será modificado. La feature no tocará `go-api/**`,
`node-api/**`, `infrastructure/**`, Docker Compose, Terraform, README raíz ni
guías/skills del repositorio.

Durante esta etapa de planificación se crea únicamente este archivo:
`ai-work/frontend-challenge-demo/design.md`.

## Execution Order

1. Generar el esqueleto React + Vite + TypeScript con el template estándar,
   conservando su ESLint como único linter. Ajustar exclusivamente sus archivos
   dentro de `frontend/` para la estructura listada arriba.
2. Añadir `frontend/.env.example` con exactamente `VITE_API_URL=` y excluir
   `.env`. La aplicación leerá `import.meta.env.VITE_API_URL` en el módulo de
   API, nunca una URL hardcodeada.
3. Definir los tipos mínimos del contrato Go y la función pura
   `parseMatrix`. Integrarla en el submit antes de cualquier request.
4. Implementar `calculateQr`: construir `POST {base-sin-slash-final}/qr`,
   enviar `{ matrix }`, validar de forma mínima el JSON exitoso y convertir
   errores HTTP/red/respuesta inesperada en mensajes seguros para la UI.
5. Implementar los cuatro componentes y el estado local de `App`; mantener
   resultados únicamente en memoria y conservar intactos los `number` que
   envía/recibe la API.
6. Aplicar CSS sencillo y responsive, incluyendo estado de foco visible y
   scroll horizontal para tablas de matrices anchas.
7. Ejecutar lint y build. Hacer la validación manual indicada abajo. El tramo
   de navegador hacia Cloud Run se registrará como bloqueado hasta aprobar y
   desplegar la feature backend `frontend-cors`.

## Architecture Decisions

### Estructura y dependencias

La estructura deliberadamente pequeña será:

```text
frontend/
├── .env.example
├── src/
│   ├── api/calculateQr.ts
│   ├── components/MatrixInput.tsx
│   ├── components/MatrixTable.tsx
│   ├── components/Statistics.tsx
│   ├── lib/parseMatrix.ts
│   ├── App.tsx
│   ├── main.tsx
│   ├── styles.css
│   └── types.ts
└── [configuración estándar de Vite]
```

Se usarán únicamente las dependencias que genera el template oficial Vite
React + TypeScript: `react`, `react-dom`, `vite`, `@vitejs/plugin-react`,
`typescript` y las dependencias de desarrollo de ESLint que ese template
requiere. `fetch` del navegador cubre la llamada HTTP. No se añadirán router,
Redux, Zustand, React Query, tablas, iconos, UI kit, CSS-in-JS, Firebase, ni
dependencias de test. Mantener ESLint del template da un control de calidad
estándar sin introducir una segunda herramienta de formato/lint.

### Formato de entrada, parseo y validación

`MatrixInput` ofrecerá un textarea cuyo texto auxiliar mostrará:

```text
One row per line. Separate values with spaces or commas.
12 -51 4
6 167 -68
-4 24 -41
```

La sintaxis aceptada será una fila por línea. Los valores se pueden separar
por uno o más espacios, comas opcionalmente rodeadas por espacios, o una
combinación de ambos: `1 2`, `1, 2` y `1, 2 3` son válidos. Se recortará el
texto completo para no tratar el salto de línea final habitual como una fila;
una línea vacía entre filas sí se rechazará como fila vacía accidental.
Comas iniciales/finales o consecutivas se rechazarán, en vez de ignorar
silenciosamente valores faltantes.

Cada token debe seguir una sintaxis decimal explícita: signo opcional,
enteros o decimales (incluido `.5`) y exponente opcional (`1e-3`). Tras
convertir con `Number`, también debe pasar `Number.isFinite`. Esto admite
enteros, negativos, cero, decimales y notación científica finita; rechaza
`NaN`, `Infinity`, tokens vacíos y texto no numérico. No se redondea ni se
modifica ningún número antes del `fetch`.

El parser devolverá el primer problema más útil, junto con el número de fila
cuando aplica, con este orden práctico:

1. entrada vacía/solo whitespace: `Enter at least one matrix row.`;
2. fila vacía interna: `Row N is empty.`;
3. separador inválido o token no numérico/no finito: `Row N contains an invalid number.`;
4. filas con longitudes distintas: `All rows must contain the same number of values.`;
5. matriz con más columnas que filas: `Enter at least as many rows as columns.`.

Así se reflejan exactamente las reglas matemáticas reales de Go: al menos una
fila y una columna, rectangularidad, valores finitos y `rows >= columns`. No
se impondrán límites arbitrarios de dimensión, ni restricciones de signo o de
enteros. El botón `Load example` cargará la matriz válida 3 × 3 verificada por
las pruebas de Go: `12 -51 4 / 6 167 -68 / -4 24 -41`.

### API, configuración y estado

`calculateQr` recibirá `number[][]` y obtendrá su base de
`import.meta.env.VITE_API_URL`. Rechazará con un error de configuración claro
una variable ausente, vacía o que no sea URL absoluta HTTP(S), antes de llamar
a la red. El módulo quitará únicamente los slash finales de la base y añadirá
`/qr`, por lo que una base Cloud Run normal con o sin slash final funciona. La
misma build sirve para frontend local → Go Cloud Run y para Firebase Hosting →
Go Cloud Run: solo cambia el valor de build `VITE_API_URL`.

`App` será el único propietario de estado React: texto de entrada, resultado,
error de UI y `isSubmitting`. El submit limpia el error anterior, parsea,
detiene el flujo si hay error local y, si es válido, inicia un único `fetch`.
Mientras `isSubmitting` es verdadero, el botón principal queda deshabilitado y
cambia a `Calculating...`; el botón de ejemplo también se deshabilita para no
cambiar la entrada durante la operación. El resultado anterior se conservará
hasta un éxito nuevo; un fallo no lo sustituye por datos inválidos.

### Manejo de errores

La UI mostrará un único mensaje junto al formulario mediante una región de
estado accesible (`role="alert"` para errores), sin `alert()`, stack traces,
objetos JavaScript ni detalles internos.

| Situación | Mensaje/criterio de UI |
| --- | --- |
| Validación local | El mensaje específico del parser. No hay request. |
| `400` con `{ message }` string | Mostrar el mensaje seguro que entrega el handler Go. |
| Otro `4xx` | `The request was rejected. Check the matrix and try again.` |
| `502`, `504` o `500` con mensaje contractual conocido | Mostrar el mensaje seguro de Go; para cualquier otro `5xx`, `The service could not complete the request. Please try again.` |
| Rechazo de `fetch` | `The request could not be completed. Check your connection and try again.` CORS también aparece aquí por restricciones del navegador. |
| `2xx` no JSON o con campos incompatibles/no finitos | `The service returned an unexpected response. Please try again.` |

Antes de renderizar una respuesta exitosa se comprobará que Q y R sean arrays
rectangulares de números finitos y que los cinco campos de estadísticas sean
de tipo y finitud esperados. Esto es una defensa de presentación, no una
segunda fuente de reglas de negocio.

### Presentación de resultados y números flotantes

Tras éxito, una sección `Results` mostrará `Q Matrix`, `R Matrix` y
`Statistics`. `MatrixTable` usará una tabla HTML con encabezado/`caption`
visualmente discreto, celdas con fuente monospace y un contenedor
`overflow-x: auto`; por ello las matrices anchas seguirán siendo legibles en
móvil. Q y R permanecen como `number[][]` sin mutación en state.

Para lectura, una única función de presentación aplicará `toFixed(6)` y quitará
ceros finales. Si `abs(value) < 0.0000005`, mostrará `0`; este umbral solo
oculta el ruido visual esperado de QR, como `1.2246467991473532e-16`. No se
altera el valor recibido, no se usa el valor formateado para cálculos y las
estadísticas no se recalculan en el cliente. Valores muy grandes/pequeños que
no encajen bien en fijo podrán presentarse con notación exponencial compacta,
también solo visualmente.

`Statistics` mostrará cinco cards pequeñas y sin gráficos: `Maximum`,
`Minimum`, `Average`, `Total sum` (campo `sum`) y `Diagonal matrix` (campo
`hasDiagonalMatrix`, presentado como `Yes`/`No`). Los números usarán la misma
función visual, sin cambiar los valores contractuales.

### Diseño visual, responsive y accesibilidad

Una sola página tendrá un fondo claro neutral cálido, una card centrada de
ancho moderado, un acento azul/gris sobrio, bordes suaves, sombra mínima,
espacio generoso y jerarquía simple: título, explicación breve, formulario y
resultados. No habrá dark mode, gradients, glassmorphism, iconos, animaciones
complejas, ilustraciones ni estética de dashboard.

El formulario y resultados usarán `main`, `section`, headings ordenados,
`label` asociado al textarea y botones nativos. El foco tendrá un outline
visible, los colores alcanzarán contraste razonable, el área de ayuda estará
asociada con `aria-describedby`, y loading/errores se anunciarán con regiones
vivas apropiadas. En desktop las estadísticas se ordenarán en grid; en tablet
y móvil pasarán a una o dos columnas. Los controles mantendrán tamaños táctiles
razonables y las tablas usarán scroll horizontal en vez de comprimir números.

### Tests y tooling

No se agregará una suite de tests frontend ni Vitest para unas pocas pruebas
de parser: implicaría dependencias y configuración adicionales que no aportan
suficiente valor a esta demo. `parseMatrix` seguirá siendo una función pura y
pequeña para que pueda probarse después sin refactor. La evidencia inicial será
TypeScript estricto, ESLint, build Vite y los escenarios manuales. El backend
ya mantiene pruebas independientes de QR y su contrato.

## Validation Criteria

Después de implementar, desde `frontend/` los comandos configurados serán:

```bash
npm install
npm run lint
npm run build
npm run dev
```

El script `build` estándar del template será `tsc -b && vite build`; por tanto
incluye type-check estricto y no se añadirá un script `typecheck` redundante.
`lint` será `eslint .`; no se añadirán Biome ni Prettier en este proyecto.

Validación manual propuesta:

1. Configurar una copia local no versionada de `.env` con un `VITE_API_URL` de
   Go y arrancar `npm run dev`; comprobar la página en `http://localhost:5173`.
2. Probar textarea vacío, línea vacía interna, token inválido, fila ragged y
   una matriz con más columnas que filas; confirmar que no se hace request.
3. Probar enteros, negativos, decimales, cero, comas, espacios y el botón
   `Load example`.
4. Enviar la matriz válida y verificar que el botón dice `Calculating...` y
   evita un segundo submit mientras espera.
5. Una vez resuelto CORS, comprobar que el navegador llama solo a
   `{VITE_API_URL}/qr` y que se ven Q, R y los cinco campos de estadísticas.
6. Simular/configurar URL inaccesible y confirmar el mensaje de red seguro;
   si hay acceso a respuestas Go controladas, verificar `400`, `502`, `504`,
   `500` y respuesta exitosa malformada.
7. Revisar teclado, foco visible, mensajes de error y vista de escritorio,
   tablet y móvil, incluyendo scroll horizontal de matrices anchas.

## External Dependencies

- Node.js y npm compatibles con la versión vigente de Vite elegida al generar
  el template.
- Una URL absoluta HTTP(S) de la API pública Go proporcionada mediante
  `VITE_API_URL`; no se versionará ningún valor de despliegue.
- El servicio Go Cloud Run y su dependencia Node deben estar disponibles para
  la validación end-to-end.
- La corrección previa del CORS descrita abajo.

## Blocked Tasks

### CORS: feature backend separada `frontend-cors`

El código actual de `go-api/httpapi/app.go` solo registra `POST /qr`. No tiene
middleware CORS, ruta `OPTIONS` ni headers `Access-Control-Allow-*`. Como un
`POST` cross-origin con `Content-Type: application/json` desde
`http://localhost:5173` hace preflight, el navegador probablemente bloqueará
el flujo local → Go Cloud Run, aunque el endpoint sea correcto.

Por protección de alcance, esta feature no modificará Go. Se requiere diseñar,
aprobar, implementar y desplegar `frontend-cors` como feature backend antes de
considerar aprobado el recorrido navegador → Go Cloud Run. Esa feature deberá
permitir explícitamente los orígenes locales necesarios y el origen Firebase
Hosting definitivo, los métodos/headers mínimos y `OPTIONS`; no debe usar un
origen comodín sin una decisión de seguridad explícita. Hasta entonces se
pueden validar el build, validación local y UI, pero no el éxito end-to-end
desde un navegador.

### Decisiones que requieren aprobación humana

1. Aprobar la creación del proyecto Vite React + TypeScript bajo `frontend/`
   con las dependencias estándar del template y ESLint como único linter.
2. Aprobar la sintaxis de textarea: espacios y comas, líneas internas vacías
   inválidas y notación científica finita aceptada.
3. Aprobar el formato visual de seis decimales, ceros finales eliminados y
   valores con magnitud menor de `0.0000005` mostrados como `0`, sin alterar
   state ni datos enviados/recibidos.
4. Aprobar no añadir tests frontend por ahora y basar la validación inicial en
   lint, build y pruebas manuales.
5. Aprobar que el tramo end-to-end queda bloqueado y que `frontend-cors` se
   trate como feature backend separada, sin cambios a Go en esta feature.
6. Tras estas decisiones, aprobar explícitamente este archivo antes de invocar
   Developer o crear/modificar cualquier archivo de `frontend/`.
