package Structs

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// NOTA: Recordar que los atributos de los struct deben iniciar con mayuscula
type MBR struct {
	MbrSize    int32        //mbr_tamano
	FechaC     [16]byte     //mbr_fecha_creacion
	Id         int32        //mbr_dsk_signature (random de forma unica)
	Fit        [1]byte      //dsk_fit
	Partitions [4]Partition //mbr_partitions
}

type Partition struct {
	Status      [1]byte  //part_status
	Type        [1]byte  //part_type
	Fit         [1]byte  //part_fit
	Start       int32    //part_start
	Size        int32    //part_s
	Name        [16]byte //part_name
	Correlative int32    //part_correlative
	Id          [4]byte  //part_id
}

// Setear valores de la particion
func (p *Partition) SetInfo(newType string, fit string, newStart int32, newSize int32, name string, correlativo int32) {
	p.Size = newSize
	p.Start = newStart
	p.Correlative = correlativo
	copy(p.Name[:], name)
	copy(p.Fit[:], fit)
	copy(p.Status[:], "I")
	copy(p.Type[:], newType)
}

// Metodos de Partition
func GetName(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)//retorna la posicion del primer byte nulo
	//Si posicionNulo retorna -1 no hay bytes nulos
	if posicionNulo != -1 {//esto nos dice que si la particion tiene nombre y es diferente de bytes nulos
		//guarda la cadena hasta el primer byte nulo (elimina los bytes nulos)
		nombre = nombre[:posicionNulo]//guardamos el nombre de la particion
	}
	return nombre
}
//metodo para obtener el id de la particion
func GetId(nombre string) string {
	//si existe id, no contiene bytes nulos
	posicionNulo := strings.IndexByte(nombre, 0)
	//si posicionNulo  no es -1, no existe id.
	if posicionNulo != -1 {
		nombre = "-"//retorna un guion para indicar que no hay id
	}
	return nombre
}
//metodo para obtener el final de la particion
func (p *Partition) GetEnd() int32 {
	return p.Start + p.Size //retorna el final de la particion por ejemplo si la particion inicia en 0 y tiene un tamaño de 10, retorna 10
}
//metodo para obtener la particion logica
type EBR struct {
	Status [1]byte //part_mount (si esta montada)
	Type   [1]byte
	Fit    [1]byte  //part_fit
	Start  int32    //part_start
	Size   int32    //part_s
	Name   [16]byte //part_name
	Next   int32    //part_next
}
//metodo para setear valores de la particion logica
func (e *EBR) SetInfo(fit string, newStart int32, newSize int32, name string, newNext int32) {
	e.Size = newSize
	e.Start = newStart
	e.Next = newNext
	copy(e.Name[:], name)
	copy(e.Fit[:], fit)
	copy(e.Status[:], "I")
	copy(e.Type[:], "L")
}
//funcion para obtener el final de la particion logica
func (e *EBR) GetEnd() int32 {
	return e.Start + e.Size + int32(binary.Size(e))// aqui se suma el tamaño de la particion logica
}

// Reportes de los Structs
func PrintMBR(data MBR) {
	fmt.Println("\n     Disco")
	fmt.Printf("CreationDate: %s, fit: %s, size: %d, id: %d\n", string(data.FechaC[:]), string(data.Fit[:]), data.MbrSize, data.Id)
	for i := 0; i < 4; i++ {
		fmt.Printf("Partition %d: %s, %s, %d, %d, %s, %d\n", i, string(data.Partitions[i].Name[:]), string(data.Partitions[i].Type[:]), data.Partitions[i].Start, data.Partitions[i].Size, string(data.Partitions[i].Fit[:]), data.Partitions[i].Correlative)
	}
}

func PrintEbr(data EBR) {
	fmt.Println("part_status ", string(data.Status[:]))
	fmt.Println("part_type ", string(data.Type[:]))
	fmt.Println("part_fit: ", string(data.Fit[:]))
	fmt.Println("part_start: ", data.Start)
	fmt.Println("part_s ", data.Size)
	fmt.Println("part_name: ", string(data.Name[:]))
	fmt.Println("next_part: ", data.Next)
}