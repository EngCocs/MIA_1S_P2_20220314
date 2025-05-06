import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import { ConsoleProvider } from "./contexto/ConsoleContex.jsx";
console.log("Renderizando componente Consola");

createRoot(document.getElementById('root')).render(
  

  <StrictMode>
    <ConsoleProvider>
      <App />
    </ConsoleProvider>
  </StrictMode>
)
