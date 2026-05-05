package worker

import (
    "database/sql"
    "jobqueue/internal/job" // importamos nuestro propio paquete job
    "log"
    "time"
)

// Pool es el corazón de nuestra API. Maneja un grupo de workers (trabajadores).
// Imagínalo como una fábrica con N empleados que procesan pedidos de una cola.
//
// Ejemplo visual:
//
//  [POST /jobs] → jobChannel → [worker 1] → procesa → guarda en DB
//                            → [worker 2] → procesa → guarda en DB
//                            → [worker 3] → procesa → guarda en DB
//
type Pool struct {
    // jobChannel es un "channel" de Go. Los channels son tuberías por donde
    // se pasan datos entre goroutines de forma segura.
    // make(chan job.Job, 100) crea un channel que puede guardar hasta 100 jobs
    // en espera sin bloquear al que los envía.
    jobChannel chan job.Job

    db *sql.DB // conexión a la base de datos

    // numWorkers es cuántos trabajadores paralelos tendremos.
    // Si ponemos 5, Go creará 5 goroutines procesando jobs al mismo tiempo.
    numWorkers int
}

// NewPool crea y retorna un nuevo Pool listo para usar.
// Esto se llama "constructor" aunque Go no tiene clases como otros lenguajes.
func NewPool(db *sql.DB, numWorkers int) *Pool {
    return &Pool{
        // Creamos el channel con capacidad para 100 jobs en cola
        jobChannel: make(chan job.Job, 100),
        db:         db,
        numWorkers: numWorkers,
    }
}

// Start arranca todos los workers en paralelo.
// Cada worker corre en su propia goroutine.
func (p *Pool) Start() {
    // Un bucle que se repite numWorkers veces.
    // i es el número del worker, solo para identificarlo en los logs.
    for i := 0; i < p.numWorkers; i++ {
        // "go" es la palabra mágica de Go.
        // Lanza una función en paralelo sin bloquear el programa.
        // Es como abrir una nueva pestaña en el navegador — todo corre al mismo tiempo.
        go p.runWorker(i)
    }
    log.Printf("✅ %d workers iniciados y escuchando jobs\n", p.numWorkers)
}

// Submit envía un job al channel para que sea procesado.
// Esta función es llamada cuando alguien hace POST /jobs.
func (p *Pool) Submit(j job.Job) {
    // Enviamos el job al channel.
    // Si el channel está lleno (100 jobs esperando), esta línea
    // esperará automáticamente hasta que haya espacio.
    p.jobChannel <- j
}

// runWorker es lo que hace cada worker.
// Escucha el channel y procesa jobs uno por uno.
// Esta función corre para siempre (hasta que el programa se cierre).
func (p *Pool) runWorker(id int) {
    log.Printf("Worker %d listo\n", id)

    // "range" sobre un channel es un bucle infinito que espera mensajes.
    // Cada vez que alguien hace Submit(), uno de los workers recibe el job aquí.
    // Si no hay jobs, el worker simplemente espera sin consumir CPU.
    for j := range p.jobChannel {
        log.Printf("Worker %d procesando job %s (tipo: %s)\n", id, j.ID, j.Type)

        // Cambiamos el estado a "processing" en la base de datos
        p.updateStatus(j.ID, job.StatusProcessing)

        // Aquí simulamos el trabajo real (redimensionar imagen, etc).
        // time.Sleep pausa la goroutine por 2 segundos sin bloquear las demás.
        // En producción real, aquí iría el código de procesamiento de imágenes.
        time.Sleep(2 * time.Second)

        // Cambiamos el estado a "done" cuando terminó
        p.updateStatus(j.ID, job.StatusDone)

        log.Printf("Worker %d terminó job %s ✅\n", id, j.ID)
    }
}

// updateStatus actualiza el estado de un job en PostgreSQL.
// Es una función auxiliar que usan los workers.
func (p *Pool) updateStatus(id string, status job.Status) {
    _, err := p.db.Exec(
        // $1 y $2 son placeholders seguros en PostgreSQL.
        // Evitan SQL injection (ataques donde alguien inyecta código SQL malicioso).
        // Go reemplaza $1 con status y $2 con id automáticamente.
        "UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3",
        status, time.Now(), id,
    )
    if err != nil {
        log.Printf("Error actualizando status del job %s: %v\n", id, err)
    }
}
