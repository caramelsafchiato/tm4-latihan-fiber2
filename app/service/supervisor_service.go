package service

import (
	"errors"
	"latihan-fiber2/app/model"
	"latihan-fiber2/app/repository"
	"latihan-fiber2/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id dosen tidak valid")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body JSON tidak valid")
	}
	err := s.tx.WithTx(ctx, func(tx pgx.Tx) error {
		repo := s.supervisors.WithTx(tx)

		// Kunci baris mahasiswa untuk mencegah penetapan ganda secara serentak
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		if _, err := repo.IncrementSupervised(ctx, supervisorID); err != nil {
			return err
		}

		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Created(c, "dosen pembimbing berhasil ditetapkan", nil, "")
}

func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	// Logika atomik ganti pembimbing dengan urutan kunci aman deadlock
	// (Implementasikan mirip dengan konsep Swap pada Course)
	return nil
}

func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id tidak valid")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body tidak valid")
	}
	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return fiber.NewError(fiber.StatusConflict, "kuota dosen pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return fiber.NewError(fiber.StatusConflict, "mahasiswa sudah memiliki dosen pembimbing")
	case errors.Is(err, repository.ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "data sudah diubah pihak lain")
	default:
		return helper.Internal(err)
	}
}