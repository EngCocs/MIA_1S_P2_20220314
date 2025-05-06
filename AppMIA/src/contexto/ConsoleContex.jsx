import { createContext, useContext, useState } from "react";

// Crear el contexto
const ConsoleContext = createContext();

// Proveedor del contexto
export const ConsoleProvider = ({ children }) => {
  const [code, setCode] = useState("");
  const [output, setOutput] = useState("");
  const [backendIp, setBackendIp] = useState("http://localhost:4000");

  return (
    <ConsoleContext.Provider value={{ code, setCode, output, setOutput, backendIp, setBackendIp }}>
      {children}
    </ConsoleContext.Provider>
  );
};

// Hook para usar el contexto
export const useConsole = () => useContext(ConsoleContext);
