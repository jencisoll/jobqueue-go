package database

import (
    "database/sql" // paquete estándar de Go para hablar con bases de datos
    "fmt"          // para formatear strings, como fmt.Sprintf
    "log"          // para imprimir mensajes en la consola

    // Este es un "driver" externo, le enseña a Go a hablar con PostgreSQL.
    // El guión bajo "_" significa: importa este paquete solo por sus efectos
    // secundarios (registra el driver), no lo usamos directamente.
    _ "github.com/lib/pq"
)

// DB es una variable global que guarda la conexión a la base de datos.
// Al ser global, cualquier parte del programa puede usarla.
// *sql.DB es un "puntero" a una conexión de base de datos.
// Los punteros en Go apuntan a la dirección de memoria donde vive el dato real.
var DB *sql.DB

// Config guarda los datos necesarios para conectarse a PostgreSQL.
// Es buena práctica no escribir contraseñas directo en el código,
// sino recibirlas como configuración.
type Config struct {
    Host     string // dirección del servidor, ej: "localhost"
    Port     string // puerto de PostgreSQL, por defecto "5432"
    User     string // usuario de la base de datos
    Password string // contraseña
    DBName   string // nombre de la base de datos
}

// Connect establece la conexión con PostgreSQL.
// En Go, las funciones pueden retornar múltiples valores.
// Aquí retorna un error si algo falla, o nil si todo está bien.
// nil significa "nada" o "vacío" en Go.
func Connect(cfg Config) error {
    // fmt.Sprintf es como un string con variables.
    // %s es el lugar donde se insertan los strings de cfg.
    // Esto construye algo como:
    // "host=localhost port=5432 user=admin password=secret dbname=jobqueue sslmode=disable"
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
    )

    // sql.Open no conecta todavía, solo prepara la conexión.
    // Retorna dos valores: la conexión (db) y un posible error (err).
    // El := en Go declara y asigna variables al mismo tiempo.
    db, err := sql.Open("postgres", connStr)

    // En Go, siempre verificamos los errores manualmente.
    // No hay try/catch como en otros lenguajes.
    // Si err no es nil, significa que algo falló.
    if err != nil {
        // fmt.Errorf crea un nuevo error con un mensaje personalizado.
        // %w "envuelve" el error original para no perder información.
        return fmt.Errorf("error abriendo conexión: %w", err)
    }

    // db.Ping() sí intenta conectarse de verdad al servidor.
    // Sirve para verificar que los datos de conexión son correctos.
    if err = db.Ping(); err != nil {
        return fmt.Errorf("error haciendo ping a la base de datos: %w", err)
    }

    // Asignamos la conexión exitosa a la variable global DB
    // para que todo el programa pueda usarla.
    DB = db

    log.Println("✅ Conectado a PostgreSQL correctamente")
    return nil // nil significa "no hubo error", todo salió bien
}

// CreateTable crea la tabla "jobs" en PostgreSQL si no existe todavía.
// Así cuando iniciamos el programa por primera vez, la tabla se crea sola.
func CreateTable() error {
    // Esta es una consulta SQL de varias líneas.
    // Las comillas invertidas ` en Go permiten strings multilínea.
    // IF NOT EXISTS evita error si la tabla ya existe.
    query := `
        CREATE TABLE IF NOT EXISTS jobs (
            id         TEXT PRIMARY KEY,     -- identificador único
            type       TEXT NOT NULL,        -- tipo de job
            status     TEXT NOT NULL,        -- estado actual
            created_at TIMESTAMP NOT NULL,   -- fecha de creación
            updated_at TIMESTAMP NOT NULL    -- fecha de última actualización
        )
    `

    // DB.Exec ejecuta una consulta SQL que no retorna filas (INSERT, CREATE, etc).
    // El guión bajo _ descarta el primer valor retornado que no necesitamos.
    _, err := DB.Exec(query)
    if err != nil {
        return fmt.Errorf("error creando tabla: %w", err)
    }

    log.Println("✅ Tabla jobs lista")
    return nil
}
