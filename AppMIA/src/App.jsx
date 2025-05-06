// App.jsx (componente principal con rutas)
import { BrowserRouter as Router, Routes, Route, NavLink } from "react-router-dom";
import Consola from "./vistas/Consola";
import Explorer from "./vistas/explorador/Explorador";
import Login from "./vistas/Login";
import "./App.css";

export default function App() {
  return (
    <Router>
      <div className="navbar">
        <NavLink to="/" className={({ isActive }) => isActive ? "active" : ""}>Comandos</NavLink>
        <NavLink to="/explorer" className={({ isActive }) => isActive ? "active" : ""}>Explorador</NavLink>
        <NavLink to="/login" className={({ isActive }) => isActive ? "active" : ""}>LOGIN</NavLink>
      </div>
      <Routes>
        <Route path="/" element={<Consola />} />
        <Route path="/explorer" element={<Explorer />} />
        <Route path="/login" element={<Login />} />
      </Routes>
    </Router>
  );
}




