# Exploración de alcance: `frontend-docker-compose`

## Qué se construirá

Una integración local de los tres servicios existentes para poder ejecutar la
aplicación completa con `docker compose up --build`: el frontend React/Vite se
compilará como archivos estáticos y se servirá con nginx; el navegador lo
abrirá en `http://localhost:5173`, llamará al Go API en
`http://localhost:8080`, y Go continuará llamando al Node API por el DNS
interno `http://node-api:8080`. También se añadirá al final de la UI un footer
sutil con el texto exacto:

`Coding Challenge Interseguro · Developed by Andrés De la Barra Vásquez`

La URL de Go del frontend será una configuración de build (`VITE_API_URL`),
no una URL hardcodeada en React. La configuración de Compose usará
`http://localhost:8080`, porque la request la hace el navegador del host y no
un proceso dentro de la red Docker.

## Actores

- Usuario: abre la UI, introduce una matriz y observa loading, validaciones,
  resultados o errores.
- Navegador: descarga los estáticos del contenedor frontend y ejecuta el
  `fetch` al Go API.
- Frontend/nginx: sirve únicamente `dist/` en el puerto interno `80`.
- Go/Fiber: expone `/qr`, factoriza la matriz y obtiene estadísticas de Node.
- Node/Express: calcula las estadísticas y permanece accesible también en el
  puerto local `3000` para inspección directa.
- Docker Compose: provee la red default y el orden básico de arranque, sin
  healthchecks ni redes personalizadas.

## Estado actual inspeccionado

- Existe `frontend/` con React + Vite + TypeScript. `App.tsx` contiene la
  composición principal y `styles.css` contiene toda la presentación actual;
  no hay footer.
- `frontend/src/api/calculateQr.ts` lee `import.meta.env.VITE_API_URL` y llama
  `POST {base}/qr`; ya valida la URL y la forma de la respuesta.
- Los scripts frontend disponibles son `build`, `lint`, `dev` y `preview`.
  No existe un script separado `typecheck`; `build` ejecuta `tsc -b` antes de
  `vite build`.
- Existe `docker-compose.yaml`, no `compose.yaml`. Docker Compose reconoce el
  nombre actual y ya define `node-api` (`3000:8080`) y `go-api`
  (`8080:8080`), con `NODE_API_URL=http://node-api:8080` y dependencia básica
  Go→Node.
- No existe actualmente Dockerfile ni `.dockerignore` para `frontend/`.
- `go-api/httpapi/app.go` crea la app Fiber y registra únicamente `POST /qr`.
  No hay middleware CORS, manejo explícito de preflight ni una variable de
  orígenes existente. El JSON del navegador con `Content-Type` provoca una
  preflight cross-origin, por lo que el flujo `localhost:5173`→`localhost:8080`
  no funcionará de forma confiable sin un cambio mínimo en Go.
- Node ya escucha en `0.0.0.0` y su Compose actual usa el puerto interno
  `8080`; no requiere conocer al frontend.

## Flujo principal

1. Compose construye el frontend con `VITE_API_URL=http://localhost:8080`.
2. La etapa de build instala con `npm ci`, ejecuta el build TypeScript/Vite y
   deja `dist/`; la etapa runtime sirve solo esos archivos con nginx en `80`.
3. Compose inicia los tres servicios en la red default y publica frontend en
   `5173:80`, Go en `8080:8080` y Node en `3000:8080`.
4. El usuario abre `http://localhost:5173`, carga una matriz válida y envía el
   formulario.
5. El navegador hace `POST http://localhost:8080/qr` y completa la preflight
   CORS contra el origen explícito `http://localhost:5173`.
6. Go calcula QR y llama internamente a `http://node-api:8080`; Node devuelve
   las estadísticas, Go responde al navegador y la UI muestra Q, R, maximum,
   minimum, average, sum y el resultado diagonal.
7. El usuario puede verificar Node directamente en `http://localhost:3000` y
   Go directamente en `http://localhost:8080`.

## Casos límite y estados a comprobar

- Footer visible al final, con texto y tildes exactos, sin links, iconos ni
  cambios visuales ajenos a su integración.
- Build sin `VITE_API_URL` interno de Docker; `go-api` no debe aparecer en el
  JavaScript compilado como hostname de API.
- Matriz válida: se muestran Q, R y las cinco piezas de estadísticas.
- Entrada vacía, filas irregulares, valores inválidos o matriz más ancha que
  alta: se conserva el mensaje de validación existente y no se hace request.
- Envío en curso: permanece el estado loading y no se duplican requests.
- Go detenido o inaccesible: el frontend muestra su error de red existente.
- Respuestas 400, 500, 502, 504 o JSON incompatible: se conservan los
  mensajes seguros existentes del cliente.
- CORS: la preflight y el POST deben permitir exactamente el origen local;
  no se debe resolver con `*` si puede configurarse explícitamente.
- Compose sin healthchecks avanzados: `depends_on` solo expresa orden básico;
  el flujo debe tolerar que la API downstream tarde en estar lista.
- Los assets nginx deben incluir rutas estáticas, no `node_modules`, fuentes,
  `.env.local`, logs, coverage ni archivos temporales.

## Dependencias

- Docker daemon y Docker Compose/Compose plugin disponibles.
- Acceso durante build a imágenes oficiales versionadas de Node y nginx y al
  registro npm para ejecutar `npm ci`.
- `frontend/package-lock.json` debe seguir siendo la fuente de instalación
  reproducible.
- Go y Node Dockerfiles existentes deben continuar funcionando sin cambios de
  contrato.
- Para la comprobación desde navegador, Go debe aceptar CORS para
  `http://localhost:5173`; como hoy no existe soporte, será necesario el
  cambio mínimo de configuración/middleware en `go-api`, manteniendo el
  origen externo mediante environment variable.
- Las validaciones previstas son los scripts existentes del frontend,
  `docker compose config`, `docker compose build`, arranque de Compose y
  comprobación manual/HTTP del flujo.

## Archivos potencialmente afectados

Dentro del alcance declarado:

- Crear `frontend/Dockerfile` y `frontend/.dockerignore`.
- Modificar `frontend/src/App.tsx` y `frontend/src/styles.css` solo para
  integrar el footer.
- Modificar el archivo existente `docker-compose.yaml` para añadir frontend,
  su build argument, publicación `5173:80` y dependencias básicas.
- Modificar únicamente la configuración/arranque HTTP de `go-api` si es
  estrictamente necesario para CORS; no tocar QR, estadísticas ni el contrato
  de `/qr`.

El nombre solicitado `compose.yaml` no coincide con el repositorio: el archivo
real es `docker-compose.yaml`. Se considera resuelto manteniendo el archivo
actual para evitar un rename innecesario.

## Fuera de alcance

- Firebase y Firebase Hosting.
- Terraform, `infrastructure/**`, GCP, Cloud Run, Artifact Registry y URLs
  cloud.
- GitHub Actions o cualquier CI/CD.
- JWT, autenticación service-to-service o secretos.
- Cambios a QR, estadísticas, Node contracts, endpoints o payloads.
- Redes Docker personalizadas, IPs estáticas, host networking, aliases
  innecesarios o health orchestration avanzada.
- Vite dev server, `npm run dev`, PM2 o Node como runtime de estáticos.
- Nuevas dependencias de frontend, rediseño visual, iconos, logos o links.
- Ocultar Node del host: `localhost:3000` seguirá publicado para pruebas.

## Preguntas resueltas

- Build argument del frontend: `VITE_API_URL`.
- Valor para Compose: `http://localhost:8080`.
- DNS interno Go→Node: `http://node-api:8080`.
- Puertos: frontend `localhost:5173`→contenedor `80`; Go
  `localhost:8080`→`8080`; Node `localhost:3000`→`8080`.
- Runtime frontend preferido: nginx oficial versionado; copiar solo `dist/`.
- Dependencias: frontend depende de Go y Go de Node para orden básico; no se
  requiere readiness orchestration.
- Footer: texto exacto, discreto y de menor jerarquía visual.
- Se mantiene `docker-compose.yaml`; no se crea un segundo Compose ni se
  renombra el archivo existente.

## Bloqueadores o decisiones pendientes

- La implementación debe decidir el nombre de la variable CORS porque Go no
  tiene actualmente una equivalente a `CORS_ALLOWED_ORIGINS`. Debe quedar
  configurable y permitir explícitamente `http://localhost:5173`; añadir un
  nombre nuevo solo es necesario porque no existe uno que reutilizar.
- La verificación real de `docker compose build`, `docker compose up` y el
  flujo en navegador depende de que el daemon Docker esté disponible. El
  análisis estático no confirma todavía que las imágenes puedan descargarse ni
  que el runtime nginx arranque.
- La comprobación “Go detenido” requiere detener temporalmente el contenedor o
  aislarlo y confirmar el mensaje de red; no cambia el alcance de código.
- Si el entorno no ofrece navegador automatizable, la validación end-to-end
  deberá ser manual en `http://localhost:5173`, complementada con requests
  HTTP y la inspección visual del footer.
