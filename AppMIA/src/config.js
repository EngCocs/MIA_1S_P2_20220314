// src/config.js
const getBackendIp = () => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("backendIp") || "http://localhost:4000";
    }
    return "http://localhost:4000"; // fallback para SSR o build
  };
  
  export default getBackendIp;
  