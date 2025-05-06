package permiso

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

// SearchInode busca un inodo según una ruta absoluta, empezando desde el inodo dado
func SearchInode(idInodo int32, path string, superBloque Structs.Superblock, file *os.File) int32 {
	stepsPath := strings.Split(path, "/")
	tmpPath := stepsPath[1:] // de la primera en adelante

	var inodeActual Structs.Inode
	Herramientas.ReadObject(file, &inodeActual, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(Structs.Inode{})))))

	var folderBlock Structs.Folderblock
	for i := 0; i < 12; i++ {
		idBloque := inodeActual.I_block[i]
		if idBloque != -1 {
			Herramientas.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(Structs.Folderblock{})))))
			for j := 0; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				if apuntador != -1 {
					pathActual := Structs.GetB_name(string(folderBlock.B_content[j].B_name[:]))
					if tmpPath[0] == pathActual {
						if len(tmpPath) > 1 {
							subRuta := "/" + strings.Join(tmpPath[1:], "/")
							return SearchInode(apuntador, subRuta, superBloque, file)
						} else {
							fmt.Printf("DEBUG SearchInode → Encontrado: %s con inodo %d\n", pathActual, apuntador)

							return apuntador
						}
					}
				}
			}
		}
	}
	return idInodo // Si no se encuentra se retorna el mismo
}
func BuscarEnCarpeta(nombre string, padre int32, super Structs.Superblock, file *os.File) int32 {
	inodeSize := int64(binary.Size(Structs.Inode{}))
	blockSize := int64(binary.Size(Structs.Folderblock{}))

	var inodo Structs.Inode
	pos := int64(super.S_inode_start) + int64(padre)*inodeSize
	if err := Herramientas.ReadObject(file, &inodo, pos); err != nil {
		return -1
	}

	for i := 0; i < 12; i++ {
		if inodo.I_block[i] == -1 {
			continue
		}

		var folder Structs.Folderblock
		posBloque := int64(super.S_block_start) + int64(inodo.I_block[i])*blockSize
		if err := Herramientas.ReadObject(file, &folder, posBloque); err != nil {
			continue
		}

		for _, entry := range folder.B_content {
			if Structs.GetB_name(string(entry.B_name[:])) == nombre {
				// Leer el inodo hijo y validar si es carpeta
				posInodoHijo := int64(super.S_inode_start) + int64(entry.B_inodo)*inodeSize
				var inodoHijo Structs.Inode
				if err := Herramientas.ReadObject(file, &inodoHijo, posInodoHijo); err != nil {
					return -1
				}
				fmt.Printf("🔎 Inodo tipo: %c (esperado: '0' para carpeta)\n", inodoHijo.I_type[0])
				if inodoHijo.I_type[0] == '0' {
					fmt.Printf("🔍 BuscarEnCarpeta encontró: %s con inodo %d\n", nombre, entry.B_inodo)

					return entry.B_inodo //  es carpeta
				}
				return -2 //  existe pero no es carpeta
			}
		}
	}

	return -1 // no encontrado
}



// CreateFolder crea una carpeta dentro de un inodo padre
func CreateFolder(idPadre int32, nombre string, super *Structs.Superblock, file *os.File, partStart int64) int32 {
	//  Reservar primero el inodo
	fmt.Printf("📂 PREVIO - CreateFolder va a usar inodo %d\n", super.S_first_ino)

	nInodo := super.S_first_ino
	super.S_first_ino++
	super.S_free_inodes_count--
	file.WriteAt([]byte{1}, int64(super.S_bm_inode_start)+int64(nInodo))

	//  Reservar bloque
	nBloque, err := session.ObtenerBloqueLibre(file, super)
	if err != nil {
		fmt.Println("Error al obtener bloque libre:", err)
		return -1
	}
	super.S_first_blo++
	super.S_free_blocks_count--
	file.WriteAt([]byte{1}, int64(super.S_bm_block_start)+int64(nBloque))

	//  Crear y configurar el bloque de carpeta
	var folderBlock Structs.Folderblock
	copy(folderBlock.B_content[0].B_name[:], ".")
	folderBlock.B_content[0].B_inodo = nInodo
	copy(folderBlock.B_content[1].B_name[:], "..")
	folderBlock.B_content[1].B_inodo = idPadre
	for i := 2; i < 4; i++ {
		folderBlock.B_content[i].B_inodo = -1
	}

	blockPos := int64(super.S_block_start) + int64(nBloque)*int64(binary.Size(Structs.Folderblock{}))
	Herramientas.WriteObject(file, folderBlock, blockPos)

	// 📦 Crear y configurar el inodo de carpeta
	var nuevoInodo Structs.Inode
	copy(nuevoInodo.I_type[:], "0") // tipo carpeta
	copy(nuevoInodo.I_perm[:], "664")
	nuevoInodo.I_uid = 1
	nuevoInodo.I_gid = 1
	nuevoInodo.I_size = int32(binary.Size(Structs.Folderblock{}))
	fecha := time.Now().Format("02/01/2006 15:04")
	copy(nuevoInodo.I_ctime[:], fecha)
	copy(nuevoInodo.I_mtime[:], fecha)
	copy(nuevoInodo.I_atime[:], fecha)
	for i := 0; i < 15; i++ {
		nuevoInodo.I_block[i] = -1
	}
	nuevoInodo.I_block[0] = nBloque

	inodePos := int64(super.S_inode_start) + int64(nInodo)*int64(binary.Size(Structs.Inode{}))
	Herramientas.WriteObject(file, nuevoInodo, inodePos)

	//  Insertar entrada en la carpeta padre
	AgregarEntradaACarpeta(file, super, idPadre, nombre, nInodo)

	//  Guardar superbloque actualizado
	if err := Herramientas.WriteObject(file, *super, partStart); err != nil {
		fmt.Println(" Error al guardar superbloque:", err)
  }

	fmt.Printf("📁 CreateFolder creando: %s como carpeta (inodo %d)\n", nombre, nInodo)
	return nInodo
}


// AgregarEntradaACarpeta agrega una nueva carpeta o archivo al folderblock del inodo padre
func AgregarEntradaACarpeta(file *os.File, super *Structs.Superblock, idPadre int32, nombre string, idInodoHijo int32) error {
	inodeSize := int32(binary.Size(Structs.Inode{}))
	blockSize := int32(binary.Size(Structs.Folderblock{}))

	var padreInodo Structs.Inode
	posInodo := int64(super.S_inode_start) + int64(idPadre)*int64(inodeSize)
	if err := Herramientas.ReadObject(file, &padreInodo, posInodo); err != nil {
		return fmt.Errorf("error al leer inodo padre: %v", err)
	}

	for i := 0; i < 12; i++ {
		if padreInodo.I_block[i] == -1 {
			continue
		}
		blockPos := int64(super.S_block_start) + int64(padreInodo.I_block[i])*int64(blockSize)

		var folderBlock Structs.Folderblock
		if err := Herramientas.ReadObject(file, &folderBlock, blockPos); err != nil {
			return fmt.Errorf("error al leer folderblock: %v", err)
		}

		for j := 0; j < 4; j++ {
			if folderBlock.B_content[j].B_inodo == -1 {
				copy(folderBlock.B_content[j].B_name[:], nombre)
				folderBlock.B_content[j].B_inodo = idInodoHijo
				if err := Herramientas.WriteObject(file, folderBlock, blockPos); err != nil {
					return fmt.Errorf("error al escribir folderblock: %v", err)
				}
				// Guardar el inodo padre actualizado también (importante)
				if err := Herramientas.WriteObject(file, padreInodo, posInodo); err != nil {
					return fmt.Errorf("error al guardar inodo padre después de actualizar carpeta: %v", err)
				}
				return nil
				
			}
		}
	}
	// Si no hay espacio, buscar un nuevo bloque libre
	nuevoBloque, err := session.ObtenerBloqueLibre(file, super)
	if err != nil {
		return fmt.Errorf("no se pudo obtener bloque libre: %v", err)
	}

	// Crear nuevo folder block
	var nuevoFolder Structs.Folderblock
	for i := 0; i < 4; i++ {
		nuevoFolder.B_content[i].B_inodo = -1
	}
	copy(nuevoFolder.B_content[0].B_name[:], nombre)
	nuevoFolder.B_content[0].B_inodo = idInodoHijo

	// Buscar el primer apuntador libre en I_block[]
	asignado := false
	for i := 0; i < 12; i++ {
		if padreInodo.I_block[i] == -1 {
			padreInodo.I_block[i] = nuevoBloque
			asignado = true
			break
		}
	}
	if !asignado {
		return fmt.Errorf("no hay espacio en I_block[] para el nuevo bloque")
	}
	
	// Escribir el nuevo bloque
	blockPos := int64(super.S_block_start) + int64(nuevoBloque)*int64(blockSize)
	if err := Herramientas.WriteObject(file, nuevoFolder, blockPos); err != nil {
		return fmt.Errorf("error al escribir nuevo bloque: %v", err)
	}
	

	// Actualizar bitmap
	file.WriteAt([]byte{1}, int64(super.S_bm_block_start)+int64(nuevoBloque))

	// Actualizar superbloque
	super.S_free_blocks_count--
	super.S_first_blo++

	superPos := int64(super.S_inode_start) - int64(binary.Size(Structs.Superblock{}))
	Herramientas.WriteObject(file, *super, superPos)

	// Escribir el inodo actualizado
	if err := Herramientas.WriteObject(file, padreInodo, posInodo); err != nil {
		return fmt.Errorf("error al guardar inodo padre: %v", err)
	}


	
	return nil
}

// VerificarContenidoCarpeta imprime el contenido de un inodo como carpeta(opcional ya que en los reportes se mira mejor)
func VerificarContenidoCarpeta(file *os.File, super Structs.Superblock, idInodo int32) {
	var inodo Structs.Inode
	inodePos := int64(super.S_inode_start) + int64(idInodo)*int64(binary.Size(Structs.Inode{}))
	Herramientas.ReadObject(file, &inodo, inodePos)

	for i := 0; i < 12; i++ {
		block := inodo.I_block[i]
		if block == -1 {
			continue
		}

		var folder Structs.Folderblock
		blockPos := int64(super.S_block_start) + int64(block)*int64(binary.Size(Structs.Folderblock{}))
		Herramientas.ReadObject(file, &folder, blockPos)
		//Esto solo me sirve par ver que se crean las carpetas 
		fmt.Printf("Contenido del bloque %d:\n", i)
		for _, content := range folder.B_content {
			if content.B_inodo != -1 {
				fmt.Printf("  → %s (inodo %d)\n", Structs.GetB_name(string(content.B_name[:])), content.B_inodo)
			}
		}
	}
}




