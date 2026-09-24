package service

import (
	"latihan-fiber2/app/model"
	"latihan-fiber2/helper"
)

// CanAccessStudent memutuskan apakah user berhak mengakses suatu data student[cite: 7]
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// 1. Pemilik data SELALU diizinkan mengakses datanya sendiri
	if current.UserID == ownerID {
		return true
	}

	// 2. Kalau bukan pemilik data, periksa apakah rolenya punya permission :any[cite: 8]
	// (misalnya 'student:read:any' atau 'student:update:any')
	return perms.Can(current.Role, anyPermission)
}