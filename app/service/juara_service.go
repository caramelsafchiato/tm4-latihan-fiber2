package service

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber2/app/repository"
	"latihan-fiber2/helper"
)

type JuaraService struct {
	repo repository.JuaraRepository
}

func NewJuaraService(repo repository.JuaraRepository) *JuaraService {
	return &JuaraService{repo: repo}
}

func (s *JuaraService) GetByStudent(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "ID mahasiswa harus berupa angka")
	}

	juaras, err := s.repo.FindByStudentID(ctx, studentID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data prestasi")
	}

	if len(juaras) == 0 {
		return helper.Fail(c, fiber.StatusNotFound, "data prestasi tidak ditemukan untuk mahasiswa ini")
	}

	return helper.Success(c, fiber.StatusOK, "daftar prestasi berhasil diambil", juaras)
}


