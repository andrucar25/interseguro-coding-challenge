# Exploración: refactorización estructural de `go-api`

## Qué se busca construir

Refactorizar la organización interna de `go-api` para separar responsabilidades y hacer más claro el flujo de la API, manteniendo el comportamiento funcional existente:

- recibir una matriz mediante el endpoint de Go/Fiber;
- validar y factorizar la matriz con QR económico;
- solicitar las estadísticas de `Q` y `R` al Node API;
- devolver `Q`, `R` y las estadísticas al cliente;
- conservar los contratos HTTP, los códigos de error y la configuración por variables de entorno salvo que una decisión posterior los cambie explícitamente.

El alcance es estructural y de legibilidad. No se pretende reescribir el algoritmo QR ni introducir una arquitectura empresarial completa.

## Actores

- Cliente HTTP: envía `POST /qr` con `{ "matrix": number[][] }`.
- Handler HTTP de Go/Fiber: traduce JSON a tipos internos, invoca el flujo de aplicación y traduce errores a respuestas HTTP.
- Lógica QR: valida la matriz y calcula `Q` y `R` sin depender de HTTP o Fiber.
- Adaptador de estadísticas: llama al endpoint Node (`POST /api/v1/statistics`) y valida su respuesta.
- Node API: servicio downstream que calcula las estadísticas.
- Configuración/composition root: lee el entorno, crea las dependencias y arranca Fiber.

## Flujo principal

1. El proceso carga puerto, URL del Node API, CORS y timeout desde el entorno (con defaults donde ya existan).
2. El punto de composición crea el cliente HTTP de estadísticas y conecta las dependencias.
3. Fiber recibe `POST /qr`.
4. El handler decodifica el cuerpo y ejecuta la operación de QR.
5. La operación valida la matriz y produce `Q` y `R`.
6. El adaptador envía `{ q, r }` al Node API y obtiene las estadísticas.
7. El handler responde con `{ q, r, statistics }`, o con el error HTTP correspondiente.

## Hallazgos sobre la estructura actual

- `main.go` mezcla lectura de entorno, creación del cliente HTTP y arranque del servidor.
- `httpapi/` contiene tanto el armado de la aplicación como DTOs, handler y mapeo de errores.
- `qr/` ya es un paquete aislado y razonablemente cohesivo para el cálculo puro.
- El cliente de estadísticas es una frontera externa clara y debe permanecer separado del cálculo QR y del transporte HTTP.
- La interfaz del colaborador de estadísticas en el paquete HTTP permite tests, pero el handler actual también coordina directamente el cálculo QR y el llamado downstream.
- La separación de un handler y un punto de configuración está justificada por sus responsabilidades distintas, no por la cantidad de archivos.

## Lectura del ejemplo de `/Users/andres/Downloads/go-api`

El ejemplo separa `config`, `handler`, `model`, `usecase` y `external/statistics`, y añade tests por capa. Es una referencia útil para límites y dependencias, pero no requiere copiar todos esos paquetes.

Su `model` concentra tipos y errores compartidos, mientras que `usecase` orquesta QR y estadísticas. Para este servicio pequeño conviene evaluar si esa capa de caso de uso aporta claridad real; crear además `model`, `usecase`, DTOs separados y múltiples interfaces podría ser más estructura de la necesaria si solo existe un flujo.

## POO, entidades y nivel de abstracción

Go no necesita POO clásica para este caso: no hay herencia, clases, ORM ni ciclo de vida de entidades persistentes. `struct` con métodos puede representar datos o encapsular una dependencia, pero eso no obliga a adoptar un modelo orientado a objetos.

La matriz y el resultado de estadísticas son tipos de dominio/transferencia, no “entidades” en el sentido de DDD o persistencia. El cliente de estadísticas puede ser un `struct` con un método porque mantiene endpoint y cliente HTTP; eso es composición idiomática de Go. Una interfaz pequeña solo se justifica en el límite que se quiere sustituir en tests (estadísticas), no como interfaz para cada paquete.

El objetivo recomendado es separación por responsabilidades y dependencias explícitas, con funciones y tipos pequeños, evitando Clean Architecture, DDD, repositorios, factories o inyección de dependencias generalizada.

## Alcance mínimo candidato

- Extraer la carga/validación de variables de entorno a `config` si se confirma que se desea centralizarla.
- Crear `handler` para router, DTOs, handler de QR y mapeo de errores HTTP.
- Mantener el cálculo en un paquete QR puro (o renombrarlo solo si mejora la semántica sin mover responsabilidades).
- Mantener el cliente Node en un paquete de adaptador externo, con sus payloads de wire separados de los tipos de dominio solo si evita mezclar contratos.
- Añadir una pequeña capa de aplicación/orquestación únicamente si ayuda a que el handler no conozca el detalle de QR + estadísticas; no asumir que hace falta un paquete `usecase` completo.
- Dejar `main` como composition root: cargar configuración, construir dependencias y arrancar el router.
- Conservar y reubicar los tests junto a la responsabilidad que prueban; agregar pruebas de configuración o de orquestación solo si la refactorización las introduce.

## Casos límite y comportamiento que debe preservarse

- JSON malformado, campo `matrix` ausente o valores no numéricos/no finitos.
- Matriz vacía, filas vacías, filas irregulares y matriz más ancha que alta.
- Fallo numérico interno de la factorización.
- Timeout, error de red, status no exitoso, JSON incompleto, tipos incorrectos o valores no finitos en la respuesta Node.
- CORS y preflight de `POST /qr`.
- No mutar la matriz de entrada durante QR.
- Errores deben seguir diferenciando `400`, `502`, `504` y `500` según el contrato actual.
- Configuración ausente o inválida debe fallar al arrancar, sin secretos ni URLs hardcodeadas.

## Dependencias y restricciones

- Go y Fiber ya usados por `go-api`.
- Cliente HTTP estándar existente o cliente Fiber del ejemplo; la elección debe ser consistente y no introducir dependencia innecesaria.
- Contrato vigente con Node: `/api/v1/statistics`, payload `{q,r}` y campos de estadísticas actuales.
- Variables existentes `PORT`, `NODE_API_URL` y `CORS_ALLOWED_ORIGINS`; cualquier nueva variable (por ejemplo timeout configurable) requiere decisión explícita.
- Debe respetarse `gofmt`, `go vet`, `go test` y `go build`.

## Fuera de alcance

- Cambiar el algoritmo QR, la precisión numérica o el significado de sus resultados.
- Cambiar endpoints, payloads, nombres de campos, códigos HTTP o el contrato Node sin una tarea de API separada.
- Crear entidades persistentes, base de datos, repositorios, ORM, autenticación o autorización.
- Introducir Clean Architecture/DDD, una jerarquía de clases o interfaces por cada dependencia.
- Cambiar `node-api`, Docker, Terraform, Cloud Run o autenticación entre servicios.
- Añadir directorios futuros del monorepo que no sean necesarios para esta refactorización.

## Preguntas resueltas para la planificación

- **¿Es necesaria POO?** No. Se recomienda composición, funciones y structs pequeños; no hay necesidad de herencia ni entidades persistentes.
- **¿Es necesaria `handler`?** Sí, si el objetivo es separar transporte HTTP de lógica y adaptadores.
- **¿Es necesario `config`?** Es razonable porque centraliza defaults y validación de entorno; su inclusión no debe añadir configuración nueva sin aprobación.
- **¿Hay que copiar el ejemplo?** No. Se puede adoptar la separación de responsabilidades sin reproducir todos sus paquetes.
- **¿Debe conservarse el contrato actual?** Sí, salvo aprobación explícita para cambiarlo.

## Bloqueos o decisiones pendientes

1. Confirmar si el alcance debe ser solo reorganización de archivos/packages o si también se acepta una capa pequeña de aplicación para encapsular “factorizar + consultar estadísticas”.
2. Confirmar si `config` debe validar únicamente las variables ya existentes o también incorporar timeout configurable (`STATISTICS_TIMEOUT_MS`) como en el ejemplo.
3. Resolver el estado de trabajo previo antes de implementar: `git status` muestra `go-api/statistics/client.go` y su test eliminados y `go-api/external/statistics/` sin seguimiento. Actualmente los imports de `main.go`/`httpapi` siguen apuntando a `go-api/statistics`, por lo que `go test ./...` falla por paquete inexistente (además, los tests HTTP que crean listeners están restringidos en este entorno). Esta condición debe reconciliarse con la intención del cambio, no sobrescribirse ciegamente.
4. El repositorio mantiene una instrucción histórica que dice que solo `node-api/` puede existir, aunque `go-api/` ya está presente y es el objeto explícito de esta solicitud; la tarea debe tratar `go-api` como alcance aprobado.
