package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber2/app/model"
	"latihan-fiber2/app/repository"
	"latihan-fiber2/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	// Endpoint ini hak aksesnya murni dijaga oleh Middleware di route.go
	return helper.Success(c, fiber.StatusOK, "Endpoint List User", nil)
}

func (s *UserService) Create(c *fiber.Ctx) error {
	// Endpoint ini hak aksesnya murni dijaga oleh Middleware di route.go
	return helper.Success(c, fiber.StatusOK, "Endpoint Create User", nil)
}

func (s *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

func (s *UserService) Replace(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}

	return helper.Success(c, fiber.StatusOK, "Endpoint Replace User (lolos authz)", nil)
}

func (s *UserService) Patch(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}

	return helper.Success(c, fiber.StatusOK, "Endpoint Patch User (lolos authz)", nil)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengubah role user")
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", result)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if current.UserID == id {
		return helper.Fail(c, fiber.StatusForbidden,
			"tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus user")
	}

	return helper.Success(c, fiber.StatusNoContent, "user berhasil dihapus", nil)
}