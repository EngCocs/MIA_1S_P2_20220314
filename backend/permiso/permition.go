package permiso



func TienePermiso(perm string, tipo string, currentUID int32, currentGID int32, inodeUID int32, inodeGID int32) bool {
	if currentUID == 0 { // root UID = 0
		return true
	}

	if len(perm) != 3 {
		return false
	}

	var cat int
	if currentUID == inodeUID {
		cat = 0
	} else if currentGID == inodeGID {
		cat = 1
	} else {
		cat = 2
	}

	digit := perm[cat] - '0'
	switch tipo {
	case "r":
		return digit&4 != 0
	case "w":
		return digit&2 != 0
	case "x":
		return digit&1 != 0
	default:
		return false
	}
}
