package service

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"latihan-fiber2/app/model"
	"latihan-fiber2/app/repository"
	"latihan-fiber2/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(portfolios repository.PortfolioRepository, students repository.StudentRepository) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// GET /students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa tidak valid")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	portfolio, err := s.portfolios.GetByStudentID(ctx, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio belum dibuat")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio ditemukan", portfolio)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa tidak valid")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	var req model.CreateUpdatePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	p := model.Portfolio{
		Headline:     req.Headline,
		Bio:          req.Bio,
		Projects:     req.Projects,
		Certificates: req.Certificates,
	}

	updated, err := s.portfolios.UpsertPortfolio(ctx, studentID, p)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diperbarui", updated)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa tidak valid")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	proj := model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	}

	updated, err := s.portfolios.AddProject(ctx, studentID, proj)
	if err != nil {
		if errors.Is(err, repository.ErrProjectLimitExceeded) {
			return fiber.NewError(fiber.StatusConflict, "jumlah proyek sudah mencapai batas maksimal 20")
		}
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio belum dibuat, silakan buat portofolio terlebih dahulu")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", updated)
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa tidak valid")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("id proyek tidak boleh kosong")
	}

	err := s.portfolios.DeleteProject(ctx, studentID, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio atau proyek tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus", nil)
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) GetTopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.GetTopTechnologies(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "10 teknologi paling banyak digunakan berhasil diambil", stats)
}