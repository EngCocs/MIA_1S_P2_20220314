import { useState, useEffect } from "react";
import "./Consola.css";
import { useNavigate } from "react-router-dom";
import { useConsole } from "../contexto/ConsoleContex.jsx";

function Consola() { 
  const { code, setCode, output, setOutput, backendIp, setBackendIp } = useConsole();
  const [lines, setLines] = useState(["1"]);
  const [isLoggedIn, setIsLoggedIn] = useState(true); // Variable para controlar si el usuario está logueado
  const navigate = useNavigate();
  const [usandoIpManual, setUsandoIpManual] = useState(false);

  // Función para obtener la IP pública del backend
  const obtenerIpPublica = async () => {
    if (usandoIpManual) return; //  No sobrescribir si ya fue definida manualmente
  
    try {
      const response = await fetch("http://localhost:4000/obtener-ip"); // tu endpoint en Go
      if (!response.ok) throw new Error("No se pudo obtener la IP pública");
      const data = await response.json();
      setBackendIp(data.ip);  //  Establece la IP obtenida
    } catch (error) {
      console.error("Error al obtener la IP pública:", error);
    }
  };
  

  // Ejecutar la función para obtener la IP pública al montar el componente
  useEffect(() => {
    obtenerIpPublica();
  }, []);  // Se ejecuta solo una vez cuando el componente se monta

  const handleInputChange = (event) => {
    const value = event.target.value;
    setCode(value);
    setLines(value.split("\n").map((_, i) => i + 1));
  };

  const handleFileChange = (event) => {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => setCode(e.target.result);
    reader.readAsText(file);
  };

  const handleExecute = async () => {
    try {
      const response = await fetch(`${backendIp}/ejecutar`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ comandos: code }), 
      });

      const result = await response.json();
      setOutput(result.salida || "No se recibió salida.");
    } catch (error) {
      setOutput("Error al conectar con el backend: " + error.message);
    }
  };

  const handleClear = () => {
    setCode("");
    setLines(["1"]);
    setOutput("");
  };

  const handleLogout = async () => {
    try {
      const response = await fetch(`${backendIp}/logout`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        alert("Error al cerrar sesión.");
        return;
      }

      // Si la sesión se cierra correctamente
      setIsLoggedIn(false); // Actualizamos el estado de sesión
      alert("Sesión cerrada correctamente.");
      navigate("/login"); // Redirige al login
    } catch (error) {
      console.error("Error al cerrar sesión:", error);
      alert("No se pudo cerrar sesión.");
    }
  };

  console.log("Renderizando componente Consola");

  return (
    <div className="console-container">
      <h1 className="main-title">MIA PRO</h1>
      <div style={{ marginBottom: "10px" }}>
        <label>IP del Backend: </label>
        <input
          type="text"
          value={backendIp}
          onFocus={() => setUsandoIpManual(true)} // Al enfocar, se activa el modo manual
          onChange={(e) => setBackendIp(e.target.value)}
          placeholder="http://3.84.XXX.XXX:4000"
          style={{ width: "150px", padding: "5px" }}
        />
      </div>

      {/* Barra lateral con botones */}
      <div className="sidebar">
        <label className="btn primary">
          <i className="fas fa-folder-open"></i> Seleccionar Archivo
          <input type="file" className="hidden-file-input" onChange={handleFileChange} />
        </label>
        <div className="separator"></div>
        <button className="btn success" onClick={handleExecute}>
          <i className="fas fa-play"></i> Ejecutar
        </button>
        <div className="separator"></div>
        <button className="btn danger" onClick={handleClear}>
          <i className="fas fa-trash"></i> Limpiar
        </button>
        <div className="separator"></div>

        {/* Botón de Cerrar Sesión */}
        {isLoggedIn && (
          <button className="btn danger" onClick={handleLogout}>
            <i className="fas fa-sign-out-alt"></i> Cerrar Sesión
          </button>
        )}
      </div>

      {/* Consolas de Entrada y Salida */}
      <div className="textareas">
        {/* Entrada */}
        <div className="editor-container">
          <div className="line-numbers">
            {lines.map((num) => (
              <div key={num}>{num}</div>
            ))}
          </div>
          <textarea
            className="input-console"
            placeholder="Ingrese código aquí..."
            value={code}
            onChange={handleInputChange}
          />
        </div>

        {/* Salida */}
        <div className="editor-container">
          <textarea className="output-console" placeholder="Salida..." value={output} readOnly />
        </div>
      </div>

      {/* GitHub */}
      <a 
        href="https://github.com/EngCocs/MIA_1S2025_P1_202200314" 
        target="_blank" 
        rel="noopener noreferrer" 
        className="github-logo"
      >
        <i className="fab fa-github"></i>
      </a>
    </div>
  );
}

export default Consola;

