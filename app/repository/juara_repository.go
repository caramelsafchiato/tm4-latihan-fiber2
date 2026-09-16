package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber2/app/model"
)

type JuaraRepository interface {
	FindByStudentID(ctx context.Context, studentID int) ([]model.Juara, error)
}

type JuaraRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewJuaraRepository(pool *pgxpool.Pool) JuaraRepository {
	return &JuaraRepositoryImpl{pool: pool} 
}

func (r *JuaraRepositoryImpl) FindByStudentID(ctx context.Context, studentID int) ([]model.Juara, error) {
    // REVISI DI BARIS INI: Ganti id_juara menjadi id_prestasi, dan $2 menjadi $1
	query := `SELECT id_prestasi, nama_prestasi, juara, id_student FROM prestasi WHERE id_student = $1`
    rows, err := r.pool.Query(ctx, query, studentID)
    if err != nil {
        return nil, err
    }       
        
    defer rows.Close()

    var juaras []model.Juara
    for rows.Next() {
        var j model.Juara
        if err := rows.Scan(&j.ID, &j.NamaPrestasi, &j.Juara, &j.StudentID); err != nil {
            return nil, err
        }
        juaras = append(juaras, j)
    }

    return juaras, nil
}



