package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber2/app/service"
	"latihan-fiber2/helper"
	"latihan-fiber2/middleware"
)

// Dependencies menyimpan semua service yang dibutuhkan oleh rute
type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	UserService    *service.UserService
	AuthService    *service.AuthService
	StudentService *service.StudentService
	JuaraService   *service.JuaraService
	NilaiService   *service.NilaiService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
	auth.Get("/check-role", middleware.RequireAuth(deps.JWT), deps.AuthService.CheckRole)

	// --- wajib login, hak akses diperiksa per endpoint ---
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat data -> middleware.
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List)

	users.Post("/",
		middleware.RequirePermission(perms, "user:update:any"),
		deps.UserService.Create)

	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete)

	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole)

	// Hak bergantung pada kepemilikan data -> diperiksa di service.
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// --- students (rute dari pertemuan sebelumnya) ---
	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", deps.StudentService.Create)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
	students.Delete("/:id", deps.StudentService.Delete)

	// --- juara (rute dari pertemuan sebelumnya) ---
	api.Get("/students/:id/juaras", deps.JuaraService.GetByStudent)

	// --- nilai (rute baru) ---
	api.Get("/students/:id/nilais", middleware.RequireAuth(deps.JWT), deps.NilaiService.GetByStudent)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}