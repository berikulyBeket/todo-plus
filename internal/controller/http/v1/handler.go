package v1

import (
	"net/http"

	"github.com/berikulyBeket/todo-plus/internal/middleware"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	appauth "github.com/berikulyBeket/todo-plus/pkg/app_auth"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler is responsible for handling HTTP requests
type Handler struct {
	Usecases *usecase.UseCase
	AppAuth  appauth.Interface
	Logger   logger.Interface
	Metrics  metrics.Interface
}

// NewHandler creates a new instance of Handler
func NewHandler(usecases *usecase.UseCase, appAuth appauth.Interface, l logger.Interface, m metrics.Interface) *Handler {
	return &Handler{
		Usecases: usecases,
		AppAuth:  appAuth,
		Logger:   l,
		Metrics:  m,
	}
}

// RegisterRoutes -.
// Swagger spec:
// @title       Todo Apps
// @description Advanced todo app
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (h *Handler) RegisterRoutes(router *gin.Engine) *gin.Engine {
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.AppAuth(appauth.PublicAccess, h.AppAuth))
		{
			auth.POST("/sign-up", h.SignUp)
			auth.POST("/sign-in", h.SignIn)
		}

		api := v1.Group("/api")
		api.Use(middleware.AppAuth(appauth.PublicAccess, h.AppAuth))
		api.Use(middleware.Authentication(h.Usecases.Auth, h.Logger))
		{
			lists := api.Group("/lists")
			{
				lists.POST("/", h.CreateList)
				lists.GET("/", h.GetAllLists)
				lists.GET("/:id", h.GetListById)
				lists.PUT("/:id", h.UpdateList)
				lists.DELETE("/:id", h.DeleteList)
				lists.GET("/search", h.SearchLists)

				items := lists.Group(":id/items")
				{
					items.POST("/", h.CreateItem)
					items.GET("/", h.GetAllItems)
				}
			}

			items := api.Group("/items")
			{
				items.GET("/:id", h.GetItemById)
				items.PUT("/:id", h.UpdateItem)
				items.DELETE("/:id", h.DeleteItem)
				items.GET("/search", h.SearchItems)
			}
		}

		private := v1.Group("/private")
		private.Use(middleware.AppAuth(appauth.PrivateAccess, h.AppAuth))
		{
			api := private.Group("/api")
			{
				api.DELETE("/users/:id", h.DeleteUserByAdmin)
				api.DELETE("/lists/:id", h.DeleteListByAdmin)
				api.DELETE("/items/:id", h.DeleteItemByAdmin)
			}
		}
	}

	return router
}
