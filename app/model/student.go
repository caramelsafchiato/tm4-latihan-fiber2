package model

import "time"

// User (atau Student) merepresentasikan entitas data mahasiswa di basis data
type Student struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest menampung data saat pendaftaran baru
type CreateStudentRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ReplaceUserRequest menampung data untuk penggantian menyeluruh (PUT)
type ReplaceStudentRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest menampung data untuk pembaruan sebagian (PATCH)
type PatchStudentRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	IsActive *bool   `json:"is_active"`
}

// ListQuery menampung parameter query string untuk pagination, search, & sort
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
// Perhitungan ini pindah ke sini karena kini dipakai langsung oleh SQL.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

// WebResponse adalah format standar respons JSON sukses/gagal
type WebResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	Meta     *Meta       `json:"meta,omitempty"`
}

// Meta menampung informasi ringkasan paginasi data
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}