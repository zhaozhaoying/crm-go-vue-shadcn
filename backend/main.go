package main

import (
	"backend/internal/authctx"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/external/companysearch"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"
	"context"
	"log"
	"time"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

// @title           Backend API
// @version         1.0
// @description     后端服务API文档
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Bearer token, 格式: "Bearer {token}"
func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}
	scheduleLocation := cfg.ScheduleLocation()
	config.ApplyProcessTimezone(scheduleLocation)

	db := database.Open(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db from gorm: %v", err)
	}
	defer sqlDB.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	customerRepo := repository.NewGormCustomerRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	roleRepo := repository.NewGormRoleRepository(db)
	authTokenRepo := repository.NewGormAuthTokenRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)
	followRecordRepo := repository.NewGormFollowRecordRepository(db)
	resourcePoolRepo := repository.NewGormResourcePoolRepository(db)
	contractRepo := repository.NewGormContractRepository(db)
	dashboardRepo := repository.NewGormDashboardRepository(db)
	externalCompanySearchRepo := repository.NewGormExternalCompanySearchRepository(db)
	hanghangCRMDailyUserCallStatRepo := repository.NewGormHanghangCRMDailyUserCallStatRepository(db)
	salesDailyScoreRepo := repository.NewGormSalesDailyScoreRepository(db)
	callRecordingRepo := repository.NewCallRecordingRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	notificationRepo := repository.NewGormNotificationRepository(db)
	customerVisitRepo := repository.NewCustomerVisitRepository(db)

	customerService := service.NewCustomerService(customerRepo, systemSettingRepo, activityLogRepo)
	customerImportService := service.NewCustomerImportService(db, activityLogRepo)
	authService := service.NewAuthService(
		userRepo,
		roleRepo,
		authTokenRepo,
		cfg.JWTSecret,
		time.Duration(cfg.JWTExpiryHours)*time.Hour,
		time.Duration(cfg.RefreshTokenExpiryHours)*time.Hour,
	)
	userService := service.NewUserService(userRepo, roleRepo)
	roleService := service.NewRoleService(roleRepo)
	systemSettingService := service.NewSystemSettingService(systemSettingRepo)
	followRecordService := service.NewFollowRecordService(followRecordRepo, activityLogRepo)
	resourcePoolService := service.NewResourcePoolService(
		resourcePoolRepo,
		customerService,
		customerRepo,
		cfg.BaiduMapAK,
		cfg.BaiduMapBaseURL,
	)
	contractService := service.NewContractService(contractRepo, systemSettingRepo, activityLogRepo)
	dashboardService := service.NewDashboardService(dashboardRepo, scheduleLocation)
	hanghangCRMDailyUserCallStatService := service.NewHanghangCRMDailyUserCallStatService(hanghangCRMDailyUserCallStatRepo, cfg.HanghangCRMCloudToken)
	salesDailyScoreService := service.NewSalesDailyScoreService(salesDailyScoreRepo)
	callRecordingService := service.NewCallRecordingService(callRecordingRepo)
	callRecordingSyncService := service.NewCallRecordingSyncService(callRecordingService, cfg.FeigeCallRecordingCookie)
	hanghangCRMDailyScoreRuntime := service.NewHanghangCRMDailyScoreRuntime(
		hanghangCRMDailyUserCallStatService,
		salesDailyScoreService,
		time.Minute,
		scheduleLocation,
	)
	customerAutoDropService := service.NewCustomerAutoDropService(db, systemSettingRepo, activityLogRepo)
	customerAutoDropRuntime := service.NewCustomerAutoDropRuntime(customerAutoDropService, time.Minute)
	captchaService := service.NewCaptchaService(2*time.Minute, 5)
	externalCompanySearchHub := service.NewExternalCompanySearchHub(64)
	externalCompanySearchHTTPClient := companysearch.NewDefaultHTTPClient()
	madeInChinaHTTPClient := companysearch.NewHTTPClient(companysearch.HTTPClientConfig{
		Timeout:               30 * time.Second,
		ConnectTimeout:        15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		RetryCount:            2,
		RetryWait:             2 * time.Second,
		ProxyURL:              cfg.MadeInChinaProxyURL,
		DisableHTTP2:          true,
	})
	googleSearchHTTPClient := companysearch.NewHTTPClient(companysearch.HTTPClientConfig{
		ProxyURL: cfg.GoogleProxyURL,
	})
	externalCompanySearchRuntime := service.NewExternalCompanySearchRuntime(
		externalCompanySearchRepo,
		externalCompanySearchHub,
		cfg.SearchWorkerCount,
		time.Duration(cfg.SearchPollIntervalMS)*time.Millisecond,
		companysearch.NewAlibabaProvider(externalCompanySearchHTTPClient, cfg.AlibabaSearchBaseURL),
		companysearch.NewMadeInChinaProvider(madeInChinaHTTPClient, cfg.MadeInChinaBaseURL),
		companysearch.NewGoogleProvider(googleSearchHTTPClient, cfg.GoogleAPIKey, cfg.GoogleCX, cfg.GoogleSearchNum),
	)
	externalCompanySearchService := service.NewExternalCompanySearchService(
		externalCompanySearchRepo,
		externalCompanySearchRuntime,
		externalCompanySearchRuntime,
	)
	alibabaEnricher := companysearch.NewAlibabaEnricher(externalCompanySearchHTTPClient)
	madeInChinaEnricher := companysearch.NewMadeInChinaEnricher(madeInChinaHTTPClient)
	websiteContactExtractor := companysearch.NewWebsiteContactExtractor(externalCompanySearchHTTPClient)
	externalCompanyEnrichService := service.NewExternalCompanyEnrichService(
		externalCompanySearchRepo,
		alibabaEnricher,
		madeInChinaEnricher,
		websiteContactExtractor,
	)
	uploadService, err := service.NewUploadService(cfg)
	if err != nil {
		log.Printf("upload service init failed: %v", err)
	}
	customerVisitLocationResolver := service.NewNominatimReverseGeocoder(
		cfg.ReverseGeocodeBaseURL,
		cfg.ReverseGeocodeUserAgent,
	)
	customerVisitService := service.NewCustomerVisitService(customerVisitRepo, customerVisitLocationResolver)

	authContextProvider := authctx.NewProvider(userRepo, roleRepo)

	healthHandler := handler.NewHealthHandler("backend", version, gitCommit, buildTime)
	customerHandler := handler.NewCustomerHandler(customerService, customerImportService)
	crontabHandler := handler.NewCrontabHandler(
		customerAutoDropService,
		hanghangCRMDailyUserCallStatService,
		salesDailyScoreService,
		scheduleLocation,
	)
	authHandler := handler.NewAuthHandler(authService, authContextProvider, captchaService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	systemSettingHandler := handler.NewSystemSettingHandler(systemSettingService)
	followRecordHandler := handler.NewFollowRecordHandler(followRecordService)
	resourcePoolHandler := handler.NewResourcePoolHandler(resourcePoolService)
	uploadHandler := handler.NewUploadHandler(uploadService)
	contractHandler := handler.NewContractHandler(contractService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	notificationHandler := handler.NewNotificationHandler(activityLogRepo, notificationRepo)
	customerVisitHandler := handler.NewCustomerVisitHandler(customerVisitService)
	salesDailyScoreHandler := handler.NewSalesDailyScoreHandler(salesDailyScoreService)
	callRecordingHandler := handler.NewCallRecordingHandler(callRecordingService, callRecordingSyncService, authContextProvider)
	externalCompanySearchHandler := handler.NewExternalCompanySearchHandler(
		externalCompanySearchService,
		externalCompanyEnrichService,
		externalCompanySearchHub,
		cfg.FrontendOrigin,
	)

	appCtx := context.Background()
	customerAutoDropRuntime.Start(appCtx)
	hanghangCRMDailyScoreRuntime.Start(appCtx)
	externalCompanySearchRuntime.Start(appCtx)

	engine := router.New(
		cfg,
		healthHandler,
		dashboardHandler,
		customerHandler,
		externalCompanySearchHandler,
		crontabHandler,
		authHandler,
		userHandler,
		roleHandler,
		systemSettingHandler,
		followRecordHandler,
		resourcePoolHandler,
		uploadHandler,
		contractHandler,
		notificationHandler,
		customerVisitHandler,
		salesDailyScoreHandler,
		callRecordingHandler,
		authTokenRepo,
	)
	if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
	}
	addr := ":" + cfg.AppPort

	log.Printf(
		"starting server on %s (%s) version=%s commit=%s buildTime=%s",
		addr,
		cfg.AppEnv,
		version,
		gitCommit,
		buildTime,
	)
	log.Printf("swagger docs: http://localhost%s/swagger/index.html", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
