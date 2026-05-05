// "main" es el nombre especial del paquete de entrada.
// Todo programa en Go empieza por el paquete main.
package main

import (
    "database/sql"
    "encoding/json" // para convertir structs de Go a JSON y viceversa
    "jobqueue/internal/database"
    "jobqueue/internal/job"
    "jobqueue/internal/worker"
    "log"
    "net/http" // el paquete estándar de Go para crear servidores web
    "os"       // para leer variables de entorno
    "time"

    "github.com/google/uuid" // para generar IDs únicos tipo "a3f2-bc91-..."
)

// workerPool es la variable global que usaremos en los handlers.
var workerPool *worker.Pool

// main() es la función que Go ejecuta primero cuando arranca el programa.
// Es equivalente al "index.js" en Node.js.
func main() {
    // Conectamos a PostgreSQL usando variables de entorno.
    // os.Getenv("VARIABLE") lee variables del sistema operativo.
    // Esto es más seguro que escribir contraseñas en el código.
    err := database.Connect(database.Config{
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
    })
    if err != nil {
        // log.Fatal imprime el error y cierra el programa.
        // Si no podemos conectar a la DB, no tiene sentido continuar.
        log.Fatal("No se pudo conectar a la base de datos:", err)
    }

    // Creamos la tabla si no existe
    if err := database.CreateTable(); err != nil {
        log.Fatal("No se pudo crear la tabla:", err)
    }

    // Creamos el pool con 5 workers paralelos.
    // Eso significa que 5 jobs se procesarán al mismo tiempo.
    workerPool = worker.NewPool(database.DB, 5)
    workerPool.Start()

    // http.HandleFunc registra una función para una ruta específica.
    // Cuando alguien visita esa ruta, Go llama a esa función.
    //
    // Es equivalente a Express en Node.js:
    // app.get('/jobs', ...)   → Go: http.HandleFunc("/jobs", ...)
    http.HandleFunc("/jobs", jobsHandler)
    http.HandleFunc("/jobs/", jobByIDHandler) // la barra final captura /jobs/cualquier-id
    http.HandleFunc("/health", healthHandler)

    log.Println("🚀 Servidor corriendo en http://localhost:8080")

    // http.ListenAndServe inicia el servidor y escucha en el puerto 8080.
    // Esta función bloquea el programa para siempre (hasta que lo cerremos).
    // Si falla (ej: el puerto está ocupado), retorna un error.
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal("Error iniciando servidor:", err)
    }
}

// jobsHandler maneja las rutas GET /jobs y POST /jobs.
// En Go, un "handler" es una función que recibe la request y escribe la response.
// w = writer (donde escribimos la respuesta)
// r = request (lo que nos mandó el cliente)
func jobsHandler(w http.ResponseWriter, r *http.Request) {
    // r.Method contiene el método HTTP: "GET", "POST", "DELETE", etc.
    switch r.Method {

    case http.MethodPost: // POST /jobs → crear un nuevo job
        createJob(w, r)

    case http.MethodGet: // GET /jobs → listar todos los jobs
        listJobs(w, r)

    default:
        // Si alguien usa DELETE, PUT, etc en esta ruta, les decimos que no está permitido.
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
    }
}

// createJob crea un nuevo job, lo guarda en DB y lo envía al worker pool.
func createJob(w http.ResponseWriter, r *http.Request) {
    // Decodificamos el JSON que nos mandaron en el body de la request.
    // Esperamos algo como: { "type": "resize_image" }
    var input struct {
        Type string `json:"type"`
    }

    // json.NewDecoder lee el body de la request.
    // .Decode(&input) convierte ese JSON a nuestra struct.
    // &input significa "la dirección de memoria de input" (un puntero).
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "JSON inválido", http.StatusBadRequest)
        return
    }

    // Validamos que type no esté vacío
    if input.Type == "" {
        http.Error(w, "El campo 'type' es requerido", http.StatusBadRequest)
        return
    }

    // Creamos el job con todos sus datos
    now := time.Now()
    j := job.Job{
        ID:        uuid.New().String(), // genera un ID único como "a3f2bc91-..."
        Type:      input.Type,
        Status:    job.StatusPending,
        CreatedAt: now,
        UpdatedAt: now,
    }

    // Guardamos el job en PostgreSQL
    _, err := database.DB.Exec(
        "INSERT INTO jobs (id, type, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
        j.ID, j.Type, j.Status, j.CreatedAt, j.UpdatedAt,
    )
    if err != nil {
        http.Error(w, "Error guardando job", http.StatusInternalServerError)
        return
    }

    // Enviamos el job al worker pool para que lo procese en paralelo.
    // Esto no bloquea — el job se encola y el handler responde de inmediato.
    workerPool.Submit(j)

    // Respondemos con el job creado en formato JSON con status 201 Created.
    respondJSON(w, http.StatusCreated, j)
}

// listJobs obtiene todos los jobs de la DB y los retorna como JSON.
func listJobs(w http.ResponseWriter, r *http.Request) {
    // DB.Query ejecuta un SELECT y retorna múltiples filas.
    rows, err := database.DB.Query(
        "SELECT id, type, status, created_at, updated_at FROM jobs ORDER BY created_at DESC",
    )
    if err != nil {
        http.Error(w, "Error consultando jobs", http.StatusInternalServerError)
        return
    }
    // defer asegura que rows.Close() se ejecute cuando la función termine.
    // Es importante cerrar los recursos para no tener memory leaks.
    defer rows.Close()

    // Creamos un slice (lista dinámica) de jobs.
    // Lo inicializamos vacío con {} para que si no hay jobs,
    // retorne [] en JSON y no null.
    jobs := []job.Job{}

    // rows.Next() avanza a la siguiente fila, retorna false cuando se acaban.
    for rows.Next() {
        var j job.Job
        // rows.Scan extrae los valores de la fila actual a las variables de Go.
        // El orden debe coincidir con el SELECT de arriba.
        err := rows.Scan(&j.ID, &j.Type, &j.Status, &j.CreatedAt, &j.UpdatedAt)
        if err != nil {
            http.Error(w, "Error leyendo jobs", http.StatusInternalServerError)
            return
        }
        // Agregamos el job al slice
        jobs = append(jobs, j)
    }

    respondJSON(w, http.StatusOK, jobs)
}

// jobByIDHandler maneja GET /jobs/:id
func jobByIDHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // r.URL.Path contiene la ruta completa, ej: "/jobs/abc-123"
    // [len("/jobs/"):] recorta los primeros 7 caracteres para obtener solo "abc-123"
    id := r.URL.Path[len("/jobs/"):]
    if id == "" {
        http.Error(w, "ID requerido", http.StatusBadRequest)
        return
    }

    var j job.Job
    // DB.QueryRow es como Query pero para una sola fila.
    err := database.DB.QueryRow(
        "SELECT id, type, status, created_at, updated_at FROM jobs WHERE id = $1", id,
    ).Scan(&j.ID, &j.Type, &j.Status, &j.CreatedAt, &j.UpdatedAt)

    if err == sql.ErrNoRows {
        // sql.ErrNoRows es el error especial que Go retorna cuando no encuentra la fila
        http.Error(w, "Job no encontrado", http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, "Error buscando job", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, j)
}

// healthHandler es un endpoint simple para verificar que la API está viva.
// Los servicios de monitoreo llaman este endpoint periódicamente.
// Si responde 200, todo está bien. Si no responde, algo está mal.
func healthHandler(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, http.StatusOK, map[string]string{
        "status": "ok",
        "time":   time.Now().Format(time.RFC3339),
    })
}

// respondJSON es una función auxiliar que usamos en todos los handlers.
// Configura los headers correctos y convierte cualquier dato a JSON.
func respondJSON(w http.ResponseWriter, status int, data any) {
    // Le decimos al cliente que la respuesta es JSON
    w.Header().Set("Content-Type", "application/json")
    // Escribimos el código de estado HTTP (200, 201, 404, etc)
    w.WriteHeader(status)
    // json.NewEncoder escribe el JSON directo al ResponseWriter
    json.NewEncoder(w).Encode(data)
}
