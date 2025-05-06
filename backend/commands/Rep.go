package commands

import (
	"backend/Reportes"
	"backend/Structs"
	"fmt"
	"strings"
)

func Rep(entrada []string) (string, error){
	var salida strings.Builder
	salida.WriteString("========REP========")
	var name, path, id,pathFileLS string

	for _, arg := range entrada[1:] {
		tmp := strings.TrimSpace(arg)
		param := strings.SplitN(tmp, "=", 2)
		if len(param) != 2 {
			salida.WriteString(fmt.Sprintf("REP Error: Parámetro incorrecto:", tmp))
			return salida.String(), nil
		}
		key := strings.ToLower(param[0])
		value := strings.ReplaceAll(param[1], "\"", "")

		switch key {
		case "name":
			name = strings.ToLower(value)
		case "path":
			path = value
		case "id":
			id = value
		case "path_file_ls":
			pathFileLS = value
		default:
			salida.WriteString(fmt.Sprintf("REP Error: Parámetro no reconocido:", key))
			return salida.String(), nil
		}
	}

	if name == "" || path == "" || id == "" {
		salida.WriteString(fmt.Sprintf("REP Error: Parámetros obligatorios faltantes."))
		return salida.String(), nil
	}

	switch name {
	case "mbr":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteMBR(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte MBR:", err))
				}
				return salida.String(), err
			}
		}
	case "disk":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteDISK(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte DISK:", err))
				}
				return salida.String(), err
			}
		}
	case "inode":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteInode(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte SB:", err))
				}
				return salida.String(), err
			}
		}
	case "block":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteBlock(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte Block:", err))
				}
				return salida.String(), err
			}
		}
	case "bm_inode":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteBMInode(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte BM Inode:", err))
				}
				return salida.String(), err
			}
		}
	case "bm_block":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteBmBlock(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte BM Block:", err))
				}
				return salida.String(), err
			}
		}
	case "tree":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				err := reportes.GenerarReporteTree(montada.PathM, path)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte Tree:", err))
				}
				return 	salida.String(), err
			}
		}
	case "sb":
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.GenerarReporteSB(montada.PathM, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte SB:", err))
				}
				return salida.String(), err
			}
		}
	case "file":
		// if pathFileLS == "" {
		// 	salida.WriteString(fmt.Sprintf("REP Error: El parámetro -path_file_ls es obligatorio para el reporte file.")
		// 	return
		// }
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.RepFile(montada.PathM, pathFileLS, path)
				salida.WriteString(msg)

				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte File:", err))
				}
				return salida.String(), err
			}
		}
	case "ls":
		if pathFileLS == "" {
			salida.WriteString(fmt.Sprintf("REP Error: El parámetro -path_file_ls es obligatorio para el reporte ls."))
			return salida.String(), nil 
		}
		for _, montada := range Structs.Montadas {
			if montada.Id == id {
				msg,err := reportes.RepLS(montada.PathM, pathFileLS, path)
				salida.WriteString(msg)
				if err != nil {
					salida.WriteString(fmt.Sprintf("REP Error al generar reporte LS:", err))
				}
				return salida.String(), err
			}
		}
	default:
		salida.WriteString(fmt.Sprintf("REP Error: Tipo de reporte no válido:", name))
	}
	return salida.String(), nil
}
