# Development Plan - frontend-docker-compose

## Summary

Completar la prueba local con Docker Compose de los tres servicios existentes:

```text
Browser (http://localhost:5173)
  -> frontend nginx container (:80)
  -> Go API (http://localhost:8080)
  -> Node API (http://node-api:8080, red Docker por defecto)
```

El frontend React/Vite se construirá una vez con la URL pública de Go como
argumento de build (`VITE_API_URL=http://localhost:8080`) y nginx servirá solo
los estáticos resultantes. Se añadirá un footer de baja jerarquía visual con el
texto exacto solicitado. Como el `POST /qr` con JSON genera una preflight desde
el navegador, Go incorporará el soporte CORS mínimo, configurable por variable
de entorno y limitado al origen local explícito en Compose.

No se modificarán QR, estadísticas, contratos Node, infraestructura, GCP,
Firebase, autenticación, Terraform ni CI/CD.

## Files to Create

- `frontend/Dockerfile`
  - Usar un build multi-stage con una imagen oficial versionada de Node (en
    línea con la versión Alpine ya usada por el repositorio, por ejemplo
    `node:24.16.0-alpine3.22`) y runtime oficial versionado de nginx (por
    ejemplo `nginx:1.28.0-alpine3.22`).
  - En la etapa de build: definir `WORKDIR`, copiar primero `package.json` y
    `package-lock.json`, ejecutar `npm ci`, copiar las fuentes necesarias,
    declarar `ARG VITE_API_URL`, exponerlo al proceso de build con
    `ENV VITE_API_URL=${VITE_API_URL}`, y ejecutar `npm run build`.
  - En runtime: copiar exclusivamente `/app/dist` desde la etapa de build a
    `/usr/share/nginx/html`, declarar `EXPOSE 80` y usar el comando por defecto
    de nginx. No copiar fuentes, `node_modules`, archivos `.env` ni usar Node,
    Vite dev server, PM2 o `npm run dev` para servir contenido.
  - No se requiere configuración nginx adicional: la aplicación actual no
    define rutas SPA del lado cliente que necesiten fallback a `index.html`.

- `frontend/.dockerignore`
  - Excluir `node_modules`, `dist`, `.env`, `.env.*`, logs, `coverage`,
    artefactos temporales y de editor/OS.
  - Incluir una excepción `!.env.example` tras la regla `.env.*` para que el
    ejemplo pueda seguir versionado y disponible en el contexto si corresponde.
  - No excluir `package.json`, `package-lock.json`, `index.html`, los
    `tsconfig*`, `vite.config.ts` ni `src/`, porque son necesarios para
    `npm ci` y `npm run build`.

## Files to Modify

- `frontend/src/App.tsx`
  - Añadir el elemento semántico `footer` al final de la aplicación, fuera de
    la tarjeta de la calculadora, con el texto literal:
    `Coding Challenge Interseguro · Developed by Andrés De la Barra Vásquez`.
  - No añadir enlaces, iconos, logos ni lógica de aplicación.

- `frontend/src/styles.css`
  - Ajustar únicamente el layout de `.page-shell` para alojar contenido y pie
    a lo largo de la altura de la ventana, conservando la tarjeta y el diseño
    actual.
  - Añadir una clase específica para un footer pequeño, centrado/coherente,
    con color atenuado, espaciado contenido y jerarquía inferior al contenido
    principal. Mantener las reglas responsive existentes y evitar un rediseño.

- `docker-compose.yaml`
  - Mantener este nombre de archivo existente; no crear ni renombrar a
    `compose.yaml`, ya que Docker Compose reconoce `docker-compose.yaml` y el
    rename no aporta funcionalidad.
  - Preservar sin cambios innecesarios los servicios `node-api` y `go-api`,
    sus builds, puertos y la URL interna `NODE_API_URL=http://node-api:8080`.
  - Añadir `CORS_ALLOWED_ORIGINS=http://localhost:5173` al entorno de `go-api`.
  - Añadir el servicio `frontend` con `build.context: ./frontend`,
    `build.args.VITE_API_URL: http://localhost:8080`, publicación `5173:80` y
    `depends_on: [go-api]` únicamente para expresar el orden básico.
  - Usar la red default creada por Compose; no declarar redes, aliases, IPs,
    `network_mode`, healthchecks ni orquestación de readiness adicional.

- `go-api/main.go`
  - Leer la nueva configuración `CORS_ALLOWED_ORIGINS` desde el entorno y
    pasarla al constructor de la aplicación HTTP junto con el cliente de
    estadísticas. El nombre es nuevo solo porque la aplicación no tiene una
    variable CORS existente que reutilizar.
  - Mantener `NODE_API_URL`, el puerto, el timeout y el flujo Go→Node actuales.

- `go-api/httpapi/app.go`
  - Extender el constructor de forma directa y mínima para recibir los
    orígenes permitidos e instalar el middleware CORS de Fiber v3 antes de
    registrar `POST /qr`.
  - Configurar `AllowOrigins` con el valor recibido, junto con los métodos y
    headers estrictamente necesarios para el navegador (`POST`, `OPTIONS` y
    `Content-Type`). No usar `*`, credenciales, nuevos endpoints ni paquetes
    propios; el middleware es parte del framework ya utilizado.

- `go-api/httpapi/qr_handler_test.go`
  - Adaptar las llamadas existentes al constructor si cambia su firma.
  - Añadir pruebas HTTP pequeñas y deterministas para: preflight `OPTIONS /qr`
    desde `http://localhost:5173` (respuesta CORS permitida) y `POST /qr` con
    ese origen (cabecera `Access-Control-Allow-Origin` exacta). Incluir un
    origen distinto que no reciba dicha cabecera, para evitar una regresión a
    un permiso abierto.
  - Conservar las pruebas actuales de QR, validación y errores downstream; no
    añadir una prueba de QR o Node nueva para una feature que no cambia esas
    responsabilidades.

- `go-api/.env.example`
  - Documentar `CORS_ALLOWED_ORIGINS` con un valor local de ejemplo
    (`http://localhost:5173`) junto a las variables existentes, sin crear ni
    versionar un archivo `.env` real.

## Execution Order

1. Inspeccionar nuevamente los Dockerfiles, los scripts del frontend y la API
   de middleware CORS de Fiber v3 para aplicar las firmas exactas sin añadir
   dependencias.
2. Crear `.dockerignore` y el Dockerfile multi-stage del frontend; asegurar que
   la etapa runtime recibe solamente `dist/` y que el build consume el ARG
   `VITE_API_URL` antes de `npm run build`.
3. Integrar el footer literal en `App.tsx` y el mínimo CSS necesario para su
   posición y jerarquía visual.
4. Añadir CORS configurable en el límite HTTP de Go, cablearlo desde `main.go`
   y actualizar el ejemplo de variables; implementar las pruebas de preflight,
   origen permitido y origen no permitido.
5. Actualizar conservadoramente `docker-compose.yaml` con el servicio
   frontend, build arg, puerto y origen CORS, preservando las configuraciones
   backend existentes.
6. Ejecutar las validaciones de código y Compose. Tras levantar el stack,
   comprobar manualmente el navegador y los estados de interacción definidos
   abajo; detener temporalmente solo el contenedor Go para probar el error de
   red y volver a iniciarlo para dejar el stack saludable.

## Architecture Decisions

- **La URL se compila para el navegador, no para la red Docker.** Vite reemplaza
  `import.meta.env.VITE_API_URL` durante `npm run build`; Compose inyectará
  `http://localhost:8080` como build argument. Nunca se compilará
  `http://go-api:8080`, porque ese DNS solo existe dentro de Docker y el
  navegador del host no puede resolverlo.
- **nginx es el runtime estático.** El build permanece reproducible con
  `npm ci` y `package-lock.json`; el runtime final no contiene Node,
  dependencias ni código fuente. La alternativa de servir con Node/Vite sería
  más pesada y no corresponde a producción estática.
- **CORS explícito y portable.** `CORS_ALLOWED_ORIGINS` desacopla el origen del
  código Go. Compose establece solo `http://localhost:5173`; en un futuro otro
  host (por ejemplo Firebase Hosting) podrá configurar su propio origen sin
  cambio de fuente. No se habilita `*` y no se agregan credenciales porque el
  flujo no las usa.
- **No cambia el contrato ni la topología interna.** `/qr`, sus payloads y
  códigos se preservan. Go resuelve Node mediante `node-api:8080`; frontend no
  conoce DNS Docker. `depends_on` expresa orden, no salud, lo cual es suficiente
  porque no hay requests downstream durante el arranque.
- **Se conserva `docker-compose.yaml`.** Es el archivo existente y soportado
  por Compose; crear una segunda variante o renombrarlo aumentaría el riesgo
  sin beneficio.
- **Sin configuración nginx extra.** Dado que no existe routing cliente que
  requiera fallback, la configuración predeterminada sirve los assets de Vite
  con menor superficie de mantenimiento.

## Validation Criteria

### Static and project checks

- Desde `frontend/`, ejecutar los scripts reales disponibles:

  ```bash
  npm run lint
  npm run build
  ```

  No hay script `typecheck` separado: `npm run build` ejecuta `tsc -b` antes de
  `vite build`.

- Desde `go-api/`, aplicar y comprobar formato, y ejecutar:

  ```bash
  find . -type f -name '*.go' -exec gofmt -w {} +
  test -z "$(find . -type f -name '*.go' -exec gofmt -l {} +)"
  go vet ./...
  go test ./...
  go build ./...
  ```

- Desde la raíz, verificar la composición y las imágenes:

  ```bash
  docker compose config
  docker compose build
  docker compose up --build
  ```

  Confirmar en la salida que nginx, Go y Node arrancan; `depends_on` no debe
  interpretarse como una garantía de readiness.

### Manual end-to-end checks

- Abrir `http://localhost:5173`: la UI y el footer exacto, incluidas las
  tildes en **Andrés** y **Vásquez**, son visibles; no aparecen enlaces,
  iconos ni cambios visuales ajenos.
- Verificar que Node sigue publicado en `http://localhost:3000` y Go en
  `http://localhost:8080`; comprobar que el `OPTIONS` y `POST /qr` desde el
  origen `http://localhost:5173` reciben CORS para ese origen, sin wildcard.
- En la UI cargar una matriz válida (por ejemplo la matriz de ejemplo) y
  confirmar que se muestran Q, R, maximum, minimum, average, sum y el
  resultado diagonal. La request del navegador debe ir a
  `http://localhost:8080/qr`; la ejecución de Go debe resolver Node mediante
  `http://node-api:8080`.
- Probar entrada inválida (vacía, no numérica, filas irregulares o más columnas
  que filas) y confirmar el mensaje de validación existente sin request.
- Durante un envío confirmar el estado loading y que la UI evita requests
  duplicadas.
- Detener temporalmente `go-api` (por ejemplo `docker compose stop go-api`),
  enviar una matriz válida desde la UI y confirmar el error de red existente;
  reiniciar Go (`docker compose start go-api`) y repetir una petición exitosa.

## External Dependencies

- Docker daemon y Docker Compose plugin disponibles para construir y levantar
  el stack.
- Acceso a las imágenes oficiales versionadas de Node y nginx, y al registro
  npm durante `npm ci` en el build Docker.
- Las dependencias actuales del frontend y Fiber v3; no se agregarán paquetes,
  manifests ni servicios externos.
- Un navegador local para la comprobación visual y de flujo real. Si no hay
  automatización de navegador disponible, la inspección será manual,
  complementada por las pruebas HTTP CORS y los logs/requests de Compose.

## Blocked Tasks

- No hay bloqueadores de diseño. La validación de construcción y del flujo
  real queda condicionada a que el daemon Docker esté operativo y pueda
  descargar imágenes y dependencias.
- La implementación permanece bloqueada deliberadamente hasta revisión humana
  y aprobación explícita de este `design.md`, según el flujo del repositorio.

## Execution Results

### Completed files

- Created `frontend/Dockerfile` with a multi-stage Node build and nginx static
  runtime, and `frontend/.dockerignore` with the approved exclusions.
- Updated `frontend/src/App.tsx` and `frontend/src/styles.css` with the
  requested semantic footer and minimal page-shell layout.
- Updated `docker-compose.yaml` with the frontend service and explicit Go CORS
  origin while preserving the Node and Go service configuration.
- Updated `go-api/main.go`, `go-api/httpapi/app.go`, and
  `go-api/.env.example` to configure restrictive CORS from
  `CORS_ALLOWED_ORIGINS`.
- Updated `go-api/httpapi/qr_handler_test.go` for the constructor signature and
  allowed-preflight, allowed-POST, and rejected-origin CORS coverage.

### Validation results

- `frontend`: `npm run lint` and `npm run build` passed.
- `go-api`: `gofmt` verification, `go vet ./...`, `go test ./...`, and
  `go build ./...` passed.
- `docker compose config`, `docker compose build`, and `docker compose up --build -d`
  passed. All three containers are running and expose frontend `5173`, Go
  `8080`, and Node `3000` as designed.
- Runtime HTTP checks confirmed nginx returns the frontend, the allowed CORS
  preflight returns only `POST, OPTIONS` and `Content-Type`, an allowed POST
  completes through Go to Node, and an untrusted origin receives no
  `Access-Control-Allow-Origin` header.
- The Go container was temporarily stopped to confirm the expected connection
  failure, then restarted; a final allowed POST succeeded and the stack was
  left healthy.

### Blocked validation

- No automated browser is available in this environment, so visual inspection
  of the footer and browser-only UI interaction states was not performed.

### Out-of-scope observations

- The example runtime tag `nginx:1.28.0-alpine3.22` is not published. The
  implementation uses the verified, pinned official tag `nginx:1.28.0-alpine`;
  no scope or architecture change was required.
