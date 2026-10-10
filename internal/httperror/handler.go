package httperror

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Page struct {
	Status         int
	Code           string
	Eyebrow        string
	Title          string
	Description    string
	Icon           string
	ActionURL      string
	ActionLabel    string
	SecondaryURL   string
	SecondaryLabel string
}

func page(code string) Page {
	switch code {
	case "403":
		return Page{Status: http.StatusForbidden, Code: "403", Eyebrow: "Akses ditolak", Title: "Anda tidak memiliki akses", Description: "Akun Anda belum memiliki permission untuk membuka halaman ini. Hubungi administrator jika akses tersebut diperlukan.", Icon: "fa-shield-halved", ActionURL: "/", ActionLabel: "Kembali ke dashboard", SecondaryURL: "/profile", SecondaryLabel: "Lihat profile"}
	case "419":
		return Page{Status: 419, Code: "419", Eyebrow: "Sesi berakhir", Title: "Form sudah kedaluwarsa", Description: "Token keamanan form tidak lagi berlaku. Muat ulang halaman asal, lalu kirim kembali perubahan Anda.", Icon: "fa-clock-rotate-left", ActionURL: "/", ActionLabel: "Kembali ke dashboard", SecondaryURL: "/login", SecondaryLabel: "Login kembali"}
	case "500":
		return Page{Status: http.StatusInternalServerError, Code: "500", Eyebrow: "Gangguan server", Title: "Terjadi kesalahan pada sistem", Description: "Permintaan belum dapat diproses. Tim pengembang dapat memeriksa log server jika masalah ini terus terjadi.", Icon: "fa-triangle-exclamation", ActionURL: "/", ActionLabel: "Coba ke dashboard"}
	case "maintenance":
		return Page{Status: http.StatusServiceUnavailable, Code: "503", Eyebrow: "Pemeliharaan sistem", Title: "Kami sedang melakukan perawatan", Description: "Aplikasi sedang ditingkatkan agar tetap aman dan stabil. Silakan coba kembali beberapa saat lagi.", Icon: "fa-screwdriver-wrench", ActionURL: "/", ActionLabel: "Periksa kembali"}
	default:
		return Page{Status: http.StatusNotFound, Code: "404", Eyebrow: "Halaman tidak ditemukan", Title: "Sepertinya Anda tersesat", Description: "Alamat yang dibuka tidak tersedia, sudah dipindahkan, atau mungkin terdapat kesalahan pada tautan.", Icon: "fa-map-location-dot", ActionURL: "/", ActionLabel: "Kembali ke dashboard"}
	}
}

func Render(c *gin.Context, appName, code string) {
	item := page(code)
	c.HTML(item.Status, "errors/index.html", gin.H{"Title": item.Code, "AppName": appName, "IsAuthPage": true, "Error": item})
}

func Handler(appName, code string) gin.HandlerFunc {
	return func(c *gin.Context) { Render(c, appName, code) }
}

func Recovery(appName string) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic recovered", "error", recovered, "path", c.Request.URL.Path)
		Render(c, appName, "500")
	})
}

func Maintenance(appName, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" || path == "/errors/maintenance" || len(path) >= 8 && path[:8] == "/static/" || len(path) >= 8 && path[:8] == "/vendor/" {
			c.Next()
			return
		}
		item := page("maintenance")
		if message != "" {
			item.Description = message
		}
		c.HTML(item.Status, "errors/index.html", gin.H{"Title": "Maintenance", "AppName": appName, "IsAuthPage": true, "Error": item})
		c.Abort()
	}
}
