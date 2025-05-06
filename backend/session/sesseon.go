package session

// Active indica si hay una sesión iniciada (true) o no (false)
var Active bool = false

// CurrentUser almacena el nombre del usuario que inició sesión (si Active es true)
var CurrentUser string = ""//current user no da el nombre del usuario actual

// PartitionID almacena el ID de la partición de la sesión activa (si Active es true)
var PartitionID string = ""
