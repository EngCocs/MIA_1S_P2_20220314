import { useState, useEffect } from "react";
import "./Discos.css";
import discoImg from "../../../assets/discoC.png";
import getBackendIp from "../../../config";
function Discos({ onSeleccionarDisco }) {
  const [discos, setDiscos] = useState([]);

  useEffect(() => {
    const ip = getBackendIp();
    const obtenerDiscos = async () => {
      try {
        const response = await fetch(`${ip}/listar-disks`, {
          method: "GET",
          headers: { "Content-Type": "application/json" },
        });
        const result = await response.json();
        setDiscos(result || []);
      } catch (error) {
        console.error("Error al cargar discos:", error);
      }
    };

    obtenerDiscos();
  }, []);

  return (
    <div className="discos-container">
      <h1 className="main-title">Explorador de Discos</h1>
      <div className="discos-grid">
        {discos.length === 0 ? (
          <p className="no-discos">No hay discos creados todavía </p>
        ) : (
          discos.map((disco, i) => (
            <div key={i} className="disco-card" onClick={() => onSeleccionarDisco(disco)}>
              <img src={discoImg} alt={`Disco ${i + 1}`} className="disco-img" />
              <div className="disco-info">
                <h3>{disco.nombre}</h3>
                <p><strong>Capacidad:</strong> {formatBytes(disco.capacidad)}</p>
                <p><strong>Fit:</strong> {disco.fit}</p>
                <p><strong>Particiones:</strong> {disco.particiones?.length > 0 ? disco.particiones.map(p => p.nombre).join(", ") : "Sin particiones"}</p>

              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}


function formatBytes(bytes) {
  if (bytes >= 1048576) {
    return (bytes / 1048576).toFixed(2) + " MB";
  } else if (bytes >= 1024) {
    return (bytes / 1024).toFixed(2) + " KB";
  } else {
    return bytes + " bytes";
  }
}

export default Discos;



