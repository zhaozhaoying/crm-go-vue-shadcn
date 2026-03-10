package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"

	_ "backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(
	cfg config.Config,
	healthHandler *handler.HealthHandler,
	dashboardHandler *handler.DashboardHandler,
	customerHandler *handler.CustomerHandler,
	externalCompanySearchHandler *handler.ExternalCompanySearchHandler,
	crontabHandler *handler.CrontabHandler,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	systemSettingHandler *handler.SystemSettingHandler,
	followRecordHandler *handler.FollowRecordHandler,
	resourcePoolHandler *handler.ResourcePoolHandler,
	uploadHandler *handler.UploadHandler,
	contractHandler *handler.ContractHandler,
	notificationHandler *handler.NotificationHandler,
	tokenChecker middleware.TokenBlacklistChecker,
) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.FrontendOrigin))

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api")
	{
		api.GET("/health", healthHandler.GetHealth)
	}

	v1 := api.Group("/v1")
	{
		v1.POST("/tasks/customer-drop/run", crontabHandler.RunAutoDropTask)

		auth := v1.Group("/auth")
		{
			auth.GET("/captcha", authHandler.Captcha)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret, tokenChecker))
		{
			protected.GET("/dashboard/overview", dashboardHandler.GetOverview)

			protected.GET("/notifications/activity-logs", notificationHandler.ListActivityLogs)
			protected.GET("/notifications/read-keys", notificationHandler.ListReadKeys)
			protected.GET("/notifications/unread-count", notificationHandler.UnreadCount)
			protected.POST("/notifications/mark-read", notificationHandler.MarkAsRead)
			protected.GET("/auth/me", authHandler.Me)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/customers", customerHandler.List)
			protected.GET("/customers/my", customerHandler.ListMy)
			protected.GET("/customers/pool", customerHandler.ListPool)
			protected.GET("/customers/potential", customerHandler.ListPotential)
			protected.GET("/customers/partner", customerHandler.ListPartner)
			protected.GET("/customers/search", customerHandler.ListSearch)
			protected.POST("/customers", customerHandler.Create)
			protected.POST("/customers/import/csv", customerHandler.ImportCSV)
			protected.PUT("/customers/:id", customerHandler.Update)
			protected.POST("/customers/validate-unique", customerHandler.CheckUnique)
			protected.POST("/customers/:id/claim", customerHandler.Claim)
			protected.POST("/customers/:id/release", customerHandler.Release)
			protected.POST("/customers/:id/transfer", customerHandler.Transfer)

			// Phone management
			protected.POST("/customers/:id/phones", customerHandler.AddPhone)
			protected.GET("/customers/:id/phones", customerHandler.ListPhones)
			protected.PUT("/customers/:id/phones/:phoneId", customerHandler.UpdatePhone)
			protected.DELETE("/customers/:id/phones/:phoneId", customerHandler.DeletePhone)

			// Status logs
			protected.GET("/customers/:id/status-logs", customerHandler.ListStatusLogs)
			protected.POST("/customers/:id/status-logs", customerHandler.CreateStatusLog)

			// 合同管理
			protected.GET("/contracts/check-number", contractHandler.CheckNumber)
			protected.GET("/contracts", contractHandler.List)
			protected.GET("/contracts/:id", contractHandler.GetByID)
			protected.POST("/contracts", contractHandler.Create)
			protected.POST("/contracts/:id/audit", contractHandler.Audit)
			protected.PUT("/contracts/:id", contractHandler.Update)
			protected.DELETE("/contracts/:id", contractHandler.Delete)

			users := protected.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/search", userHandler.Search)
				users.GET("/:id", userHandler.GetByID)
				users.POST("", userHandler.Create)
				users.PUT("/batch/disable", userHandler.BatchDisable)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
				users.POST("/avatar/upload", uploadHandler.UploadAvatar)
			}

			roles := protected.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.POST("", roleHandler.Create)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
			}

			settings := protected.Group("/settings")
			{
				settings.GET("", systemSettingHandler.GetSettings)
				settings.PUT("", systemSettingHandler.UpdateSettings)
				settings.GET("/customer-levels", systemSettingHandler.GetCustomerLevels)
				settings.POST("/customer-levels", systemSettingHandler.CreateCustomerLevel)
				settings.PUT("/customer-levels/:id", systemSettingHandler.UpdateCustomerLevel)
				settings.DELETE("/customer-levels/:id", systemSettingHandler.DeleteCustomerLevel)
				settings.GET("/customer-sources", systemSettingHandler.GetCustomerSources)
				settings.POST("/customer-sources", systemSettingHandler.CreateCustomerSource)
				settings.PUT("/customer-sources/:id", systemSettingHandler.UpdateCustomerSource)
				settings.DELETE("/customer-sources/:id", systemSettingHandler.DeleteCustomerSource)
			}

			// 跟进方式管理
			protected.GET("/follow-methods", followRecordHandler.ListFollowMethods)
			protected.POST("/follow-methods", followRecordHandler.CreateFollowMethod)
			protected.PUT("/follow-methods/:id", followRecordHandler.UpdateFollowMethod)
			protected.DELETE("/follow-methods/:id", followRecordHandler.DeleteFollowMethod)

			// 运营跟进记录
			protected.GET("/operation-follow-records", followRecordHandler.ListOperationFollowRecords)
			protected.GET("/operation-follow-records/all", followRecordHandler.ListAllOperationFollowRecords)
			protected.POST("/operation-follow-records", followRecordHandler.CreateOperationFollowRecord)

			// 销售跟进记录
			protected.GET("/sales-follow-records", followRecordHandler.ListSalesFollowRecords)
			protected.GET("/sales-follow-records/all", followRecordHandler.ListAllSalesFollowRecords)
			protected.POST("/sales-follow-records", followRecordHandler.CreateSalesFollowRecord)

			// 资源池
			protected.GET("/resource-pool", resourcePoolHandler.List)
			protected.POST("/resource-pool/search", resourcePoolHandler.SearchAndStore)
			protected.POST("/resource-pool/convert/batch", resourcePoolHandler.ConvertBatchToCustomer)
			protected.POST("/resource-pool/:id/convert", resourcePoolHandler.ConvertToCustomer)

			protected.GET("/external-company-search/tasks", externalCompanySearchHandler.ListTasks)
			protected.POST("/external-company-search/tasks", externalCompanySearchHandler.CreateTasks)
			protected.GET("/external-company-search/results", externalCompanySearchHandler.ListAllResults)
			protected.GET("/external-company-search/tasks/:id", externalCompanySearchHandler.GetTask)
			protected.POST("/external-company-search/tasks/:id/cancel", externalCompanySearchHandler.CancelTask)
			protected.GET("/external-company-search/tasks/:id/results", externalCompanySearchHandler.ListResults)
			protected.GET("/external-company-search/tasks/:id/events", externalCompanySearchHandler.ListEvents)
			protected.GET("/external-company-search/tasks/:id/stream", externalCompanySearchHandler.StreamTask)
		}
	}

	return engine
}
