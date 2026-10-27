package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

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
	CourseService  *service.CourseService 
	Mongo *mongo.Client
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

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

	// Tambahkan RequireAuth agar semua request ke /students wajib pakai token
	students := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)

	// Hak dapat diputuskan tanpa melihat data -> pasang di middleware
	students.Get("/",
		middleware.RequirePermission(perms, "student:list"),
		deps.StudentService.List)

	students.Post("/",
		middleware.RequirePermission(perms, "student:create"),
		deps.StudentService.Create)

	students.Delete("/:id",
		middleware.RequirePermission(perms, "student:delete"),
		deps.StudentService.Delete)

	// Hak bergantung pada kepemilikan data (owner_id) -> diperiksa di service
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)

	// --- juara (rute dari pertemuan sebelumnya) ---
	api.Get("/students/:id/juaras", deps.JuaraService.GetByStudent)

	// --- kelas dan KRS (Tambahan Modul 9) ---
	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai urutan pendaftaran
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	enrollments.Post("/swap", deps.CourseService.Swap)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
        defer cancel()

        if err := pool.Ping(ctx); err != nil {
            return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
        }

        if err := mongoClient.Ping(ctx, nil); err != nil {
            return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
        }

        return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
    }
}