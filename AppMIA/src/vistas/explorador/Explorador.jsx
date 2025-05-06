import { useState } from "react";
import Discos from "./discos/Discos";
import Particiones from "./particiones/Particiones";
 import Archivos from "./capetasArchivos/Archivos";
import "./Explorador.css";  

export default function Explorador() {
  const [vista, setVista] = useState("discos");
  const [discoSeleccionado, setDiscoSeleccionado] = useState(null);
  const [particionSeleccionada, setParticionSeleccionada] = useState(null);
  const seleccionarDisco = (disco) => {
    setDiscoSeleccionado(disco);
    setVista("particiones");
  };
  const seleccionarParticion = (particion) => {
    console.log("🔴 particion seleccionada:", particion);  // Verificar que particion tiene los datos correctos
    // Si la partición no tiene un id, se lo puedes asignar aquí
  const particionConId = { ...particion, id: particion.nombre }; // Por ejemplo, usando el nombre como id
    setParticionSeleccionada(particionConId);
    setVista("archivos");
  };
  const volverADiscos = () => {
    setVista("discos");
    setDiscoSeleccionado(null);
  };
  const volverAParticiones = () => {
    setVista("particiones");
    setParticionSeleccionada(null);
  };

  return (
    <div className="explorador-container">
      {vista === "discos" && (
        <Discos onSeleccionarDisco={seleccionarDisco} />
      )}
      {vista === "particiones" && discoSeleccionado && (
        <Particiones disco={discoSeleccionado} volver={volverADiscos} onSeleccionarParticion={seleccionarParticion} />
      )}
      {vista === "archivos" && particionSeleccionada && (
        <Archivos
          particion={particionSeleccionada}
          volver={volverAParticiones}
        />
      )}
    </div>
  );
}

  