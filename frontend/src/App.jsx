// Importamos hooks de React:
// useState → guarda datos que cambian (como variables reactivas)
// useEffect → ejecuta código cuando el componente carga o algo cambia
import { useState, useEffect } from "react";
import "./App.css";

// La URL base de tu API — apunta a tu VPS
const API_URL = "http://161.132.50.109";

// Componente principal de la app
// En React, un componente es una función que retorna HTML (JSX)
export default function App() {

    // useState retorna dos cosas:
    // 1. El valor actual
    // 2. Una función para actualizarlo
    // Cuando actualizas el valor, React re-renderiza el componente automáticamente
    const [jobs, setJobs] = useState([]);         // lista de jobs
    const [jobType, setJobType] = useState("");   // texto del input
    const [loading, setLoading] = useState(false); // si está enviando un job
    const [error, setError] = useState("");        // mensaje de error

    // fetchJobs obtiene todos los jobs de la API
    const fetchJobs = async () => {
        try {
            // fetch es la forma nativa de hacer peticiones HTTP en el navegador
            const res = await fetch(`${API_URL}/jobs`);
            const data = await res.json();
            // Actualizamos el estado — React re-renderiza la lista automáticamente
            setJobs(data);
        } catch (err) {
            console.error("Error obteniendo jobs:", err);
        }
    };

    // useEffect se ejecuta cuando el componente carga por primera vez
    // y cada vez que cambian las dependencias del array [] (vacío = solo al inicio)
    useEffect(() => {
        // Cargamos los jobs inmediatamente al abrir la página
        fetchJobs();

        // setInterval llama a fetchJobs cada 2 segundos automáticamente
        // Así el dashboard se actualiza en "tiempo real" sin que el usuario haga nada
        const interval = setInterval(fetchJobs, 2000);

        // El return de useEffect es una función de limpieza
        // Se ejecuta cuando el componente se desmonta (se cierra la página)
        // Limpiamos el interval para no tener memory leaks
        return () => clearInterval(interval);
    }, []); // [] significa "ejecuta esto solo una vez al montar el componente"

    // createJob envía un POST a la API para crear un nuevo job
    const createJob = async () => {
        // Validamos que el input no esté vacío
        if (!jobType.trim()) {
            setError("Escribe un tipo de job");
            return;
        }

        setLoading(true); // mostramos spinner
        setError("");     // limpiamos errores anteriores

        try {
            const res = await fetch(`${API_URL}/jobs`, {
                method: "POST",
                // Le decimos a la API que mandamos JSON
                headers: { "Content-Type": "application/json" },
                // JSON.stringify convierte el objeto JS a texto JSON
                // { type: "resize_image" } → '{"type":"resize_image"}'
                body: JSON.stringify({ type: jobType }),
            });

            if (!res.ok) {
                throw new Error("Error creando job");
            }

            setJobType(""); // limpiamos el input
            fetchJobs();    // actualizamos la lista inmediatamente
        } catch (err) {
            setError("No se pudo crear el job. ¿Está la API corriendo?");
        } finally {
            // finally se ejecuta siempre, haya error o no
            setLoading(false);
        }
    };

    // statusColor retorna un color según el estado del job
    // Esto es una función auxiliar que usamos en el JSX de abajo
    const statusColor = (status) => {
        const colors = {
            pending:    "#f59e0b", // amarillo
            processing: "#3b82f6", // azul
            done:       "#10b981", // verde
            failed:     "#ef4444", // rojo
        };
        // Si el status no existe en el objeto, retorna gris
        return colors[status] || "#6b7280";
    };

    const statusEmoji = (status) => {
        const emojis = {
            pending:    "⏳",
            processing: "⚙️",
            done:       "✅",
            failed:     "❌",
        };
        return emojis[status] || "❓";
    };

    // Contamos cuántos jobs hay por estado para las métricas
    const metrics = {
        total:      jobs.length,
        pending:    jobs.filter(j => j.status === "pending").length,
        processing: jobs.filter(j => j.status === "processing").length,
        done:       jobs.filter(j => j.status === "done").length,
    };

    // JSX — es HTML dentro de JavaScript
    // Las expresiones JS van entre llaves {}
    return (
        <div className="app">

            {/* HEADER */}
            <header className="header">
                <h1>🚀 JobQueue Dashboard</h1>
                <p>API REST con Go · Worker Pool · PostgreSQL</p>
            </header>

            {/* MÉTRICAS */}
            <div className="metrics">
                <div className="metric-card">
                    <span className="metric-number">{metrics.total}</span>
                    <span className="metric-label">Total Jobs</span>
                </div>
                <div className="metric-card" style={{ borderColor: "#f59e0b" }}>
                    <span className="metric-number" style={{ color: "#f59e0b" }}>{metrics.pending}</span>
                    <span className="metric-label">Pendientes</span>
                </div>
                <div className="metric-card" style={{ borderColor: "#3b82f6" }}>
                    <span className="metric-number" style={{ color: "#3b82f6" }}>{metrics.processing}</span>
                    <span className="metric-label">Procesando</span>
                </div>
                <div className="metric-card" style={{ borderColor: "#10b981" }}>
                    <span className="metric-number" style={{ color: "#10b981" }}>{metrics.done}</span>
                    <span className="metric-label">Completados</span>
                </div>
            </div>

            {/* FORMULARIO */}
            <div className="form-card">
                <h2>Crear nuevo Job</h2>
                <div className="form-row">
                    {/*
            value={jobType} → el input muestra el valor del estado
            onChange → cada vez que el usuario escribe, actualizamos el estado
            e.target.value → el texto que escribió el usuario
          */}
                    <input
                        type="text"
                        placeholder="Tipo de job (ej: resize_image, send_email)"
                        value={jobType}
                        onChange={(e) => setJobType(e.target.value)}
                        // onKeyDown detecta teclas — si presiona Enter, crea el job
                        onKeyDown={(e) => e.key === "Enter" && createJob()}
                        className="input"
                    />
                    <button
                        onClick={createJob}
                        disabled={loading}
                        className="button"
                    >
                        {/* Mostramos texto diferente según si está cargando */}
                        {loading ? "Enviando..." : "Crear Job"}
                    </button>
                </div>
                {/* && en JSX significa "muestra esto solo si la condición es true" */}
                {error && <p className="error">{error}</p>}
            </div>

            {/* LISTA DE JOBS */}
            <div className="jobs-section">
                <h2>Jobs ({jobs.length}) <span className="live-badge">● LIVE</span></h2>

                {jobs.length === 0 ? (
                    <p className="empty">No hay jobs todavía. ¡Crea uno arriba!</p>
                ) : (
                    <div className="jobs-list">
                        {/* .map() recorre el array y retorna JSX por cada elemento */}
                        {jobs.map((job) => (
                            // key es requerido por React para identificar cada elemento de la lista
                            <div key={job.id} className="job-card">
                                <div className="job-left">
                                    <span className="job-emoji">{statusEmoji(job.status)}</span>
                                    <div>
                                        <p className="job-type">{job.type}</p>
                                        <p className="job-id">{job.id}</p>
                                    </div>
                                </div>
                                <div className="job-right">
                  <span
                      className="job-status"
                      style={{ backgroundColor: statusColor(job.status) }}
                  >
                    {job.status}
                  </span>
                                    <p className="job-date">
                                        {/* Convertimos la fecha ISO a formato legible */}
                                        {new Date(job.created_at).toLocaleTimeString()}
                                    </p>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

        </div>
    );
}