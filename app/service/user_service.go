package service

import (
	"github.com/gofiber/fiber/v2"
	
	"latihan-fiber2/app/repository"
	"latihan-fiber2/helper"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Handler sementara agar rute tidak error
func (s *UserService) List(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint List User", nil)
}
func (s *UserService) Get(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint Get User", nil)
}
func (s *UserService) Create(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint Create User", nil)
}
func (s *UserService) Replace(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint Replace User", nil)
}
func (s *UserService) Patch(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint Patch User", nil)
}
func (s *UserService) Delete(c *fiber.Ctx) error {
	return helper.Success(c, fiber.StatusOK, "Endpoint Delete User", nil)
}