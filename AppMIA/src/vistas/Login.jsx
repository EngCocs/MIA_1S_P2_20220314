import { useState } from "react";
import "./Login.css";
import { useNavigate } from "react-router-dom";

function Login() {
  const [partitionID, setPartitionID] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [backendIp, setBackendIp] = useState("http://localhost:4000");
  const navigate = useNavigate();

  const handleLogin = async (e) => {
    e.preventDefault(); // Evita que el form recargue la página

    const comando = `login -user=${username} -pass=${password} -id=${partitionID}`;
    try {
      const response = await fetch(`${backendIp}/ejecutar`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ comandos: comando }),
      });

      const result = await response.json();
      const salida = result.salida || "";

     

      if (salida.includes("LOGIN Error: Usuario no existe") ||
          salida.includes("contraseña incorrecta") ||
          salida.includes("No se encontró la partición")) {
        alert("Usuario o contraseña o partición incorrectos.");
      } else if (salida.includes("logueado correctamente")) {
        alert(" Sesión iniciada correctamente.");
        setOutput((prev) => prev + "\n" + salida);
        navigate("/"); // redirige a la consola

      }else if (salida.includes("Ya hay un usuario logueado")) {
        alert("Ya hay una sesión activa. Cierra sesión antes de iniciar otra.");
      }
    } catch (error) {
      console.error("Error al conectar con el backend:", error);
      alert(" No se pudo conectar con el backend.");
    }
  };

  return (
    <div className="login-container">
      <form className="login-form" onSubmit={handleLogin}>
        <h2 className="login-title">Login</h2>

        <input
          type="text"
          placeholder="ID Partición"
          value={partitionID}
          onChange={(e) => setPartitionID(e.target.value)}
        />
        <input
          type="text"
          placeholder="Usuario"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <input
          type="password"
          placeholder="Contraseña"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit">Iniciar sesión</button>

        {/* Campo para IP del backend */}
        <input
          type="text"
          placeholder="http://IP:4000"
          value={backendIp}
          onChange={(e) => setBackendIp(e.target.value)}
          className="ip-field"
        />

        
      </form>
    </div>
  );
}

export default Login;



  