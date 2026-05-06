import { useState, useEffect, useRef } from "react";
import "./App.css";

const API_URL = "http://161.132.50.109";

export default function App() {
    const [jobs, setJobs]           = useState([]);
    const [loading, setLoading]     = useState(false);
    const [error, setError]         = useState("");
    const [preview, setPreview]     = useState(null);
    const [apiOnline, setApiOnline] = useState(true);

    // Campos del formulario
    const [file, setFile]           = useState(null);
    const [width, setWidth]         = useState("800");
    const [height, setHeight]       = useState("600");
    const [watermark, setWatermark] = useState("");
    const [format, setFormat]       = useState("jpg");

    const fileInputRef = useRef(null);

    const fetchJobs = async () => {
        try {
            const res = await fetch(`${API_URL}/jobs`);
            const data = await res.json();
            setJobs(Array.isArray(data) ? data : []);
        } catch { setJobs([]); }
    };

    const checkHealth = async () => {
        try {
            const res = await fetch(`${API_URL}/health`);
            setApiOnline(res.ok);
        } catch { setApiOnline(false); }
    };

    useEffect(() => {
        fetchJobs();
        checkHealth();
        const interval  = setInterval(fetchJobs, 2000);
        const interval2 = setInterval(checkHealth, 5000);
        return () => { clearInterval(interval); clearInterval(interval2); };
    }, []);

    // Cuando el usuario elige una imagen, mostramos preview
    const handleFileChange = (e) => {
        const f = e.target.files[0];
        if (!f) return;
        setFile(f);
        setPreview(URL.createObjectURL(f));
        setError("");
    };

    const handleSubmit = async () => {
        if (!file) { setError("Selecciona una imagen primero"); return; }
        setLoading(true);
        setError("");

        // FormData para enviar archivo + campos
        const formData = new FormData();
        formData.append("image", file);
        formData.append("width", width);
        formData.append("height", height);
        formData.append("watermark", watermark);
        formData.append("format", format);

        try {
            const res = await fetch(`${API_URL}/upload`, {
                method: "POST",
                body: formData, // NO poner Content-Type, el navegador lo hace solo
            });
            if (!res.ok) throw new Error();
            // Limpiamos el formulario
            setFile(null);
            setPreview(null);
            setWatermark("");
            if (fileInputRef.current) fileInputRef.current.value = "";
            fetchJobs();
        } catch {
            setError("Error al subir la imagen. Verifica que la API esté corriendo.");
        } finally {
            setLoading(false);
        }
    };

    const downloadFile = (processedFile) => {
        const link = document.createElement("a");
        link.href = `${API_URL}/download/${processedFile}`;
        link.download = processedFile;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    const statusColor  = s => ({ pending:"#f59e0b", processing:"#3b82f6", done:"#10b981", failed:"#ef4444" }[s] || "#6b7280");
    const statusEmoji  = s => ({ pending:"⏳", processing:"⚙️", done:"✅", failed:"❌" }[s] || "❓");

    const metrics = {
        total:      jobs.length,
        pending:    jobs.filter(j => j.status === "pending").length,
        processing: jobs.filter(j => j.status === "processing").length,
        done:       jobs.filter(j => j.status === "done").length,
    };

    return (
        <div className="app">
            <header className="header">
                <h1>🚀 JobQueue Dashboard</h1>
                <p>API REST con Go · Worker Pool · PostgreSQL</p>
                <span className={`api-badge ${apiOnline ? "online" : "offline"}`}>
          {apiOnline ? "● API Online" : "● API Offline"}
        </span>
            </header>

            {/* MÉTRICAS */}
            <div className="metrics">
                <div className="metric-card"><span className="metric-number">{metrics.total}</span><span className="metric-label">Total Jobs</span></div>
                <div className="metric-card" style={{borderColor:"#f59e0b"}}><span className="metric-number" style={{color:"#f59e0b"}}>{metrics.pending}</span><span className="metric-label">Pendientes</span></div>
                <div className="metric-card" style={{borderColor:"#3b82f6"}}><span className="metric-number" style={{color:"#3b82f6"}}>{metrics.processing}</span><span className="metric-label">Procesando</span></div>
                <div className="metric-card" style={{borderColor:"#10b981"}}><span className="metric-number" style={{color:"#10b981"}}>{metrics.done}</span><span className="metric-label">Completados</span></div>
            </div>

            {/* FORMULARIO */}
            <div className="form-card">
                <h2>Procesar Imagen</h2>

                {/* Zona de drop de imagen */}
                <div className="upload-zone" onClick={() => fileInputRef.current.click()}>
                    {preview
                        ? <img src={preview} alt="preview" className="preview-img" />
                        : <div className="upload-placeholder"><span>🖼️</span><p>Click para seleccionar imagen</p><small>JPG, PNG, WEBP — máximo 10MB</small></div>
                    }
                </div>
                <input ref={fileInputRef} type="file" accept="image/*" onChange={handleFileChange} style={{display:"none"}} />

                {/* Opciones */}
                <div className="options-grid">
                    <div className="option">
                        <label>Ancho (px)</label>
                        <input type="number" value={width} onChange={e => setWidth(e.target.value)} className="input" />
                    </div>
                    <div className="option">
                        <label>Alto (px)</label>
                        <input type="number" value={height} onChange={e => setHeight(e.target.value)} className="input" />
                    </div>
                    <div className="option">
                        <label>Formato</label>
                        <select value={format} onChange={e => setFormat(e.target.value)} className="input">
                            <option value="jpg">JPG (comprimido)</option>
                            <option value="png">PNG (sin pérdida)</option>
                        </select>
                    </div>
                    <div className="option">
                        <label>Marca de agua</label>
                        <input type="text" placeholder="Opcional" value={watermark} onChange={e => setWatermark(e.target.value)} className="input" />
                    </div>
                </div>

                <button onClick={handleSubmit} disabled={loading || !file} className="button">
                    {loading ? "Subiendo..." : "⚡ Procesar Imagen"}
                </button>
                {error && <p className="error">{error}</p>}
            </div>

            {/* LISTA DE JOBS */}
            <div className="jobs-section">
                <h2>Jobs ({jobs.length}) <span className="live-badge">● LIVE</span></h2>
                {jobs.length === 0
                    ? <p className="empty">No hay jobs todavía. ¡Sube una imagen arriba!</p>
                    : <div className="jobs-list">
                        {jobs.map(job => (
                            <div key={job.id} className="job-card">
                                <div className="job-left">
                                    <span className="job-emoji">{statusEmoji(job.status)}</span>
                                    <div>
                                        <p className="job-type">{job.original_file?.split("_original")[0].slice(0,8)}... → {job.type?.toUpperCase()} {job.width}×{job.height}</p>
                                        {job.watermark && <p className="job-watermark">💧 {job.watermark}</p>}
                                        <p className="job-id">{job.id}</p>
                                    </div>
                                </div>
                                <div className="job-right">
                                    <span className="job-status" style={{backgroundColor: statusColor(job.status)}}>{job.status}</span>
                                    {job.status === "done" && (
                                        <button className="download-btn" onClick={() => downloadFile(job.processed_file)}>
                                            ⬇ Descargar
                                        </button>
                                    )}
                                    {job.status === "failed" && <p className="job-error">{job.error}</p>}
                                    <p className="job-date">{new Date(job.created_at).toLocaleTimeString()}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                }
            </div>
        </div>
    );
}