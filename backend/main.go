package main

import (
	"encoding/json"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	"backend/Structs"
	"backend/Herramientas"
	"backend/permiso"
	"backend/session"

	Comandos "backend/commands"
	DM "backend/commands"
	US "backend/commands/AdministacionUserAndGrups"
	MK "backend/commands/AdministacionCarpArch"
)

type RequestBody struct {
	Comandos string `json:"comandos"`
}

type ResponseBody struct {
	Salida string `json:"salida"`
}

func main() {
	http.HandleFunc("/ejecutar", ejecutarHandler)
	http.HandleFunc("/listar-disks", listarDiscosHandler)
	http.HandleFunc("/explorar", explorarParticionHandler)
	http.HandleFunc("/leer", leerArchivoHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/obtener-ip", obtenerIPPublicaHandler)


	fmt.Println("Servidor escuchando en http://localhost:4000")
	http.ListenAndServe("0.0.0.0:4000", nil)

}

func ejecutarHandler(w http.ResponseWriter, r *http.Request) {
	// Permitir CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Leer el cuerpo JSON
	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al leer cuerpo", http.StatusBadRequest)
		return
	}

	lineas := strings.Split(req.Comandos, "\n")
	var resultado strings.Builder

	for _, linea := range lineas {
		comando := strings.Split(linea, "#")[0]
		comando = strings.TrimSpace(comando)
		if comando != "" {
			resultado.WriteString("\n*********************************************************************************************\n")
			resultado.WriteString("Linea en ejecucion: " + comando + "\n")
			resultado.WriteString(analizar(comando) + "\n")
		}
	}

	// Responder al frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponseBody{Salida: resultado.String()})
}

func analizar(entrada string) string {
	parametros := strings.Split(entrada, " -")
	comando := strings.ToLower(parametros[0])
	var salida strings.Builder

	switch comando {
	case "execute":
		salida.WriteString("[ERROR] 'execute' no está permitido desde la interfaz web\n")
	case "mkdisk":
		if len(parametros) > 1 {
			out, err := DM.Mkdisk(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "fdisk":
		if len(parametros) > 1 {
			out, err := DM.Fdisk(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mount":
		if len(parametros) > 1 {
			out, err := DM.Mount(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "rmdisk":
		if len(parametros) > 1 {
			out, err := DM.Rmdisk(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mounted":
		if len(parametros) == 1 {
			out, err := DM.Mounted(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mkfs":
		if len(parametros) > 1 {
			out, err := DM.Mkfs(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "login":
		if len(parametros) > 1 {
			out, err := US.Login(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "logout":
		if len(parametros) == 1 {
			out, err := US.Logout(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mkgrp":
		if len(parametros) > 1 {
			out, err := US.Mkgrp(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "rmgrp":
		if len(parametros) > 1 {
			out, err := US.Rmgrp(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mkusr":
		if len(parametros) > 1 {
			out, err := US.Mkusr(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "rmusr":
		if len(parametros) > 1 {
			out, err := US.Rmusr(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "chgrp":
		if len(parametros) > 1 {
			out, err := US.Chgrp(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mkfile":
		if len(parametros) > 1 {
			out, err := MK.Mkfile(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "cat":
		if len(parametros) > 1 {
			out, err := MK.Cat(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "mkdir":
		if len(parametros) > 1 {
			out, err := MK.Mkdir(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "rep":
		if len(parametros) > 1 {
			out, err := Comandos.Rep(parametros)
			salida.WriteString(out)
			if err != nil {
				salida.WriteString(fmt.Sprintf("Error: %v\n", err))
			}
		} else {
			salida.WriteString("MKDISK ERROR: parámetros no encontrados\n")
		}
	case "exit":
		salida.WriteString("Salida exitosa\n")
		os.Exit(0)
	default:
		salida.WriteString("Comando no reconocible\n")
	}

	return salida.String()
}

func listarDiscosHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    carpeta := "/home/ubuntu/Calificacion_MIA/Discos" // Ajusta si tu ruta es distinta
    archivos, err := ioutil.ReadDir(carpeta)
    if err != nil {
        http.Error(w, "Error al leer los discos", http.StatusInternalServerError)
        return
    }

    type Particion struct {
        Nombre string `json:"nombre"`
        Tamano int32  `json:"tamano"`
        Fit    string `json:"fit"`
        Estado string `json:"estado"`
    }

    type Disco struct {
        Nombre      string      `json:"nombre"`
        Capacidad   int32       `json:"capacidad"`
        Fit         string      `json:"fit"`
        Particiones []Particion `json:"particiones"`
    }

    var discos []Disco

    for _, archivo := range archivos {
        if !archivo.IsDir() && filepath.Ext(archivo.Name()) == ".mia" {
            fullPath := filepath.Join(carpeta, archivo.Name())

            fileMBR, err := os.OpenFile(fullPath, os.O_RDWR, 0644)
            if err != nil {
                continue
            }

            var mbr Structs.MBR
            if err := binary.Read(fileMBR, binary.LittleEndian, &mbr); err != nil {
                fileMBR.Close()
                continue
            }

            var particiones []Particion

            // Recorrer primarias y extendida
            for _, part := range mbr.Partitions {
                nombre := Structs.GetName(string(part.Name[:]))
                if nombre != "" {
                    particiones = append(particiones, Particion{
                        Nombre: nombre,
                        Tamano: part.Size,
                        Fit:    strings.TrimSpace(string(part.Fit[:])),
                        Estado: strings.TrimSpace(string(part.Status[:])),
                    })

                    // Si es extendida, buscar lógicas
                    if strings.TrimSpace(string(part.Type[:])) == "E" {
                        var ebr Structs.EBR
                        if err := Herramientas.ReadObject(fileMBR, &ebr, int64(part.Start)); err != nil {
                            break
                        }
                        for {
                            nombreLogica := Structs.GetName(string(ebr.Name[:]))
                            if nombreLogica != "" {
                                particiones = append(particiones, Particion{
                                    Nombre: nombreLogica,
                                    Tamano: ebr.Size,
                                    Fit:    strings.TrimSpace(string(ebr.Fit[:])),
                                    Estado: strings.TrimSpace(string(ebr.Status[:])),
                                })
                            }
                            if ebr.Next == -1 {
                                break
                            }
                            if err := Herramientas.ReadObject(fileMBR, &ebr, int64(ebr.Next)); err != nil {
                                break
                            }
                        }
                    }
                }
            }

            discos = append(discos, Disco{
                Nombre:      archivo.Name(),
                Capacidad:   mbr.MbrSize,
                Fit:         strings.TrimSpace(string(mbr.Fit[:])),
                Particiones: particiones,
            })

            fileMBR.Close() 
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(discos)
}

func explorarParticionHandler(w http.ResponseWriter, r *http.Request) {
	// 💡 Encabezados CORS completos
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 💡 Responder la preflight OPTIONS sin hacer más
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Println("➡️ Se recibio solicitud a /explorar")
	fmt.Println("🆔 Sesión activa:", session.Active)
	fmt.Println("🧩 ID de partición activa:", session.PartitionID)
	

	if !session.Active {
		http.Error(w, "Sesión no activa", http.StatusForbidden)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	var pathDisco string
	for _, m := range Structs.Montadas {
		if m.Id == session.PartitionID {
			pathDisco = m.PathM
			break
		}
	}
	if pathDisco == "" {
		http.Error(w, "No se encontró la partición montada", http.StatusNotFound)
		return
	}

	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		http.Error(w, "Error al abrir el disco", http.StatusInternalServerError)
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		http.Error(w, "Error al leer MBR", http.StatusInternalServerError)
		return
	}

	var super Structs.Superblock
	encontrado := false
	for _, part := range mbr.Partitions {
		if Structs.GetId(string(part.Id[:])) == session.PartitionID {
			if err := Herramientas.ReadObject(disco, &super, int64(part.Start)); err != nil {
				http.Error(w, "Error al leer superbloque", http.StatusInternalServerError)
				return
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		http.Error(w, "No se encontró la partición activa", http.StatusInternalServerError)
		return
	}

	idInodo := permiso.SearchPath(path, disco, super)
	if idInodo == -1 {
		http.Error(w, "No se encontró la ruta "+path, http.StatusNotFound)
		return
	}

	resultado := permiso.ListarContenidoInodo(disco, super, idInodo)
	w.WriteHeader(http.StatusOK)
	fmt.Println("🧾 Explorador contenido final:", resultado)

	json.NewEncoder(w).Encode(resultado)
}


func leerArchivoHandler(w http.ResponseWriter, r *http.Request) {
	//  Encabezados CORS 
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	//  Responder preflight OPTIONS sin hacer más
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	if !session.Active {
		http.Error(w, "Sesión no activa", http.StatusForbidden)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Falta el parámetro 'path'", http.StatusBadRequest)
		return
	}

	contenido, err := MK.Cat([]string{"cat", fmt.Sprintf("-file1=%s", path)})
	if err != nil {
		http.Error(w, "Error al leer archivo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(contenido))
}



func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // 💡 Encabezados CORS completos
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    fmt.Println("➡️ Se recibió solicitud a /logout")
    fmt.Println("🆔 Sesión activa:", session.Active)

    if !session.Active {
        http.Error(w, "No hay sesión activa para cerrar", http.StatusForbidden)
        return
    }

    // Cerrar la sesión
    session.Active = false
    session.CurrentUser = ""
    session.PartitionID = ""

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "Sesión cerrada correctamente.")
}

func obtenerIPPublicaHandler(w http.ResponseWriter, r *http.Request) {
    // 💡 Encabezados CORS completos
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    // Responder la preflight OPTIONS sin hacer más
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    // Leer la IP pública de la instancia EC2
    resp, err := http.Get("http://169.254.169.254/latest/meta-data/public-ipv4")
    if err != nil {
        http.Error(w, "Error al obtener la IP pública", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    ipPublica, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Error al leer la IP pública", http.StatusInternalServerError)
        return
    }

    // Enviar la IP pública al frontend
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"ip": string(ipPublica)})
}
