import { useState, useEffect } from "react";
import folderImg from "../../../assets/folder.png";
import fileImg from "../../../assets/documento.png";
import "./Archivos.css";
import getBackendIp from "../../../config";
// Este componente permite explorar archivos y carpetas en una partición específica.
export default function Archivos({ particion, volver }) {
  const [contenido, setContenido] = useState([]);
  const [rutaActual, setRutaActual] = useState("/");
  const [contenidoArchivo, setContenidoArchivo] = useState(null);

  useEffect(() => {
    const ip = getBackendIp();
    const obtenerContenido = async () => {
      console.log("🔴 particion en useEffect:", particion);  // Verificar si particion tiene un valor
    console.log("🔴 particion.id:", particion ? particion.id : "undefined");  // Verificar el valor de particion.id

    if (!particion|| !particion.id) {
      console.error("❌ particion no está definida!");
      return;
    }
      try {
        const baseURL = `${ip}/explorar`;
        const url = `${baseURL}?path=${encodeURIComponent(rutaActual)}&particion=${particion.id}`;  // Enviar el ID de la partición

      console.log("🚀 URL de la solicitud:", url);  // Aquí imprimimos la URL para verificar que el ID está siendo pasado

        const response = await fetch(url);
        if (!response.ok) {
          const textoError = await response.text();
          console.error(" Error de backend:", textoError);
          setContenido([]);
          return;
        }

        const datos = await response.json();
        setContenido(Array.isArray(datos) ? datos : []);
        console.log("Contenido recibido:", datos);

      } catch (error) {
        console.error(" Error de conexión:", error);
        setContenido([]);
      }
    };

    obtenerContenido();
  }, [rutaActual, particion.id]);

  const entrarACarpeta = (nombre) => {
    setContenidoArchivo(null);
    setRutaActual(prev => 
      prev === "/" ? `/${nombre}` : `${prev}/${nombre}`
    );
  };

  const regresarCarpetaAnterior = () => {
    if (rutaActual === "/") return;
    setContenidoArchivo(null);
    const partes = rutaActual.split("/").filter(Boolean);
    partes.pop();
    setRutaActual(partes.length > 0 ? `/${partes.join("/")}` : "/");
  };

  const leerArchivo = async (nombre) => {
    const ip = getBackendIp();
    const rutaArchivo = rutaActual === "/" ? `/${nombre}` : `${rutaActual}/${nombre}`;
    try {
      const res = await fetch(`${ip}/leer?path=${encodeURIComponent(rutaArchivo)}`);

      const texto = await res.text();
      setContenidoArchivo({ nombre, texto });
    } catch (err) {
      console.error(" Error al leer archivo:", err);
      setContenidoArchivo({ nombre, texto: "[Error al leer el archivo]" });
    }
  };

  return (
    <div className="archivos-container">
      <h1 className="main-title">Sistema de Archivos: {rutaActual}</h1>

      <div className="botones-navegacion">
        <button className="btn-volver" onClick={volver}>
          --- Volver a Particiones ---
        </button>
        {rutaActual !== "/" && (
          <button className="btn-volver" onClick={regresarCarpetaAnterior}>
            --- Subir Carpeta ---
          </button>
        )}
      </div>
      
      <div className="archivos-grid">
        {contenido.length > 0 ? (
          contenido.map((item, i) => (
            <div 
              key={i}
              className="archivo-card"
              onClick={() => {
                if (item.tipo === "carpeta") {
                  entrarACarpeta(item.nombre);
                } else if (item.tipo === "archivo") {
                  leerArchivo(item.nombre);
                }
              }}
            >
              <img 
                src={item.tipo === "carpeta" ? folderImg : fileImg} 
                alt={item.tipo} 
                className="archivo-img"
              />
              <div className="archivo-info">
                <h3>{item.nombre}</h3>
                <p><strong>Tipo:</strong> {item.tipo === "carpeta" ? "Carpeta" : "Archivo"}</p>
                <p><strong>Permisos:</strong> {item.perm}</p>
                <p><strong>Fecha de Creación:</strong> {item.fechaCreacion}</p>
              </div>
            </div>
          ))
        ) : (
          <p className="no-archivos">No hay archivos ni carpetas en esta ruta.</p>
        )}
      </div>

      {contenidoArchivo && (
        <div className="archivo-leido">
          <button onClick={() => setContenidoArchivo(null)}>✖ Cerrar</button>
          <h2>📄 Contenido de: {contenidoArchivo.nombre}</h2>
          <pre>{contenidoArchivo.texto}</pre>
        </div>
      )}
    </div>
  );
}



  
