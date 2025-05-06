import particionImg from "../../../assets/particion.png";
import "./Particiones.css";

function Particiones({ disco, volver, onSeleccionarParticion }) {
  return (
    <div className="particiones-container">
      <h1 className="main-title">Particiones de {disco.nombre}</h1>
      <button className="btn-volver" onClick={volver}> --- Volver ---</button>
      <div className="particiones-grid">
      {disco.particiones && disco.particiones.length > 0 ? (
  disco.particiones.map((particion, i) => {
    console.log("🔴 particion en el mapa:", particion);  // Verificar que particion tiene la propiedad id
    const info = typeof particion === 'object' ? particion : { nombre: particion, tamano: 0, fit: "N/A", estado: "I" };

    return (
      <div key={i} className="particion-card" onClick={() => onSeleccionarParticion(info)}>
        <img src={particionImg} alt="Particion" className="particion-img" />
        <div className="particion-info">
          <h3>{info.nombre}</h3>
          <p><strong>Tamaño:</strong> {formatBytes(info.tamano)}</p>
          <p><strong>Fit:</strong> {info.fit}</p>
          <p><strong>Estado:</strong> {info.estado === "I" ? "Inactiva" : "Activa"}</p>
        </div>
      </div>
    );
  })
) : (
  <p className="no-particiones">No hay particiones creadas.</p>
)}

      </div>
    </div>
  );
}

function formatBytes(bytes) {
  if (!bytes || isNaN(bytes)) return "Desconocido";
  if (bytes >= 1048576) {
    return (bytes / 1048576).toFixed(2) + " MB";
  } else if (bytes >= 1024) {
    return (bytes / 1024).toFixed(2) + " KB";
  } else {
    return bytes + " bytes";
  }
}

export default Particiones;

