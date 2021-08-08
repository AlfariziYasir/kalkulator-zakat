package controllers

import (
	"fmt"
	"zakat/api/middleware"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

func (s *Server) InitializeRoutes() {
	adapter, err := gormadapter.NewAdapterByDB(s.DB)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize casbin adapter : %v", err))
	}

	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		panic(fmt.Sprintf("failed to create casbin enforcer: %v", err))
	}

	if hasPolicy := enforcer.HasPolicy("admin", "report", "read"); !hasPolicy {
		enforcer.AddPolicy("admin", "report", "read")
	}
	if hasPolicy := enforcer.HasPolicy("admin", "report", "write"); !hasPolicy {
		enforcer.AddPolicy("admin", "report", "write")
	}
	if hasPolicy := enforcer.HasPolicy("user", "report", "read"); !hasPolicy {
		enforcer.AddPolicy("user", "report", "read")
	}

	v1 := s.Router.Group("/api")
	{
		v1.POST("/login", s.Login)
		v1.POST("/register", s.CreateUser(enforcer))
	}

	v2 := v1.Group("/users", middleware.TokenMiddleware())
	{
		v2.GET("/", middleware.Authorize("report", "write", enforcer), s.GetUsers)
		v2.GET("/:uid", middleware.Authorize("report", "read", enforcer), s.GetUser)
		v2.PUT("/:uid", middleware.Authorize("report", "read", enforcer), s.UpdateUser)
		v2.DELETE("/:uid", middleware.Authorize("report", "read", enforcer), s.DeleteUser)
	}

	v3 := v1.Group("/muzakki", middleware.TokenMiddleware())
	{
		v3.POST("/", middleware.Authorize("report", "read", enforcer), s.CreateMuzakki)
		v3.GET("/", middleware.Authorize("report", "write", enforcer), s.GetMuzakkis)
		v3.GET("/:uid", middleware.Authorize("report", "read", enforcer), s.GetMuzakki)
		v3.PUT("/:uid", middleware.Authorize("report", "read", enforcer), s.UpdateMuzakki)
		v3.DELETE("/:uid", middleware.Authorize("report", "read", enforcer), s.DeleteMuzakki)
	}

	v4 := v1.Group("/zakat-fitrah", middleware.TokenMiddleware())
	{
		v4.POST("/check", s.CheckZakatFitrah)
		v4.POST("/", middleware.Authorize("report", "read", enforcer), s.CreateZakatFitrah)
		v4.GET("/", middleware.Authorize("report", "write", enforcer), s.GetZakatFitrahs)
		v4.GET("/:uid", middleware.Authorize("report", "read", enforcer), s.GetZakatFitrah)
		v4.PUT("/:uid", middleware.Authorize("report", "read", enforcer), s.UpdateZakatFitrah)
		v4.DELETE("/:uid", middleware.Authorize("report", "read", enforcer), s.DeleteZakatFitrah)
	}

	v5 := v1.Group("/zakat-mal", middleware.TokenMiddleware())
	{
		v5.POST("/check", s.CheckZakatMal)
		v5.GET("/update-price/:metal", s.UpdatePriceIDR)
		v5.POST("/", middleware.Authorize("report", "read", enforcer), s.CreateZakatMal)
		v5.GET("/", middleware.Authorize("report", "write", enforcer), s.GetZakatMals)
		v5.GET("/:uid", middleware.Authorize("report", "read", enforcer), s.GetZakatMalByID)
		v5.GET("/:uid/:type", middleware.Authorize("report", "read", enforcer), s.GetZakatMalByType)
		v5.PUT("/:uid/:type", middleware.Authorize("report", "read", enforcer), s.UpdateZakatMal)
		v5.DELETE("/:uid", middleware.Authorize("report", "read", enforcer), s.DeleteZakatMalByID)
		v5.DELETE("/:uid/:type", middleware.Authorize("report", "read", enforcer), s.DeleteZakatMalByType)
	}

}
