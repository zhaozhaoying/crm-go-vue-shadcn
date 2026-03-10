package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                  string
	AppPort                 string
	FrontendOrigin          string
	DBDriver                string
	DBPath                  string
	GormLogLevel            string
	GormSlowThresholdMS     int
	MySQLDSN                string
	MySQLHost               string
	MySQLPort               string
	MySQLUser               string
	MySQLPassword           string
	MySQLDB                 string
	MySQLCharset            string
	MySQLParseTime          string
	MySQLLoc                string
	JWTSecret               string
	JWTExpiryHours          int
	RefreshTokenExpiryHours int
	BaiduMapAK              string
	BaiduMapBaseURL         string
	OSSEndpoint             string
	OSSAccessKeyID          string
	OSSAccessKeySecret      string
	OSSBucketName           string
	OSSBasePath             string
	AlibabaSearchBaseURL    string
	MadeInChinaBaseURL      string
	GoogleAPIKey            string
	GoogleCX                string
	GoogleSearchNum         int
	GoogleProxyURL          string
	SearchWorkerCount       int
	SearchPollIntervalMS    int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:                  getEnv("APP_ENV", "local"),
		AppPort:                 getEnv("APP_PORT", "8080"),
		FrontendOrigin:          getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		DBDriver:                strings.ToLower(strings.TrimSpace(getEnv("DB_DRIVER", "sqlite"))),
		DBPath:                  getEnv("DB_PATH", "data.db"),
		GormLogLevel:            strings.ToLower(strings.TrimSpace(getEnv("GORM_LOG_LEVEL", "error"))),
		GormSlowThresholdMS:     getEnvInt("GORM_SLOW_THRESHOLD_MS", 200),
		MySQLDSN:                getEnv("MYSQL_DSN", ""),
		MySQLHost:               getEnv("MYSQL_HOST", "127.0.0.1"),
		MySQLPort:               getEnv("MYSQL_PORT", "3306"),
		MySQLUser:               getEnv("MYSQL_USER", ""),
		MySQLPassword:           getEnv("MYSQL_PASSWORD", ""),
		MySQLDB:                 getEnv("MYSQL_DB", ""),
		MySQLCharset:            getEnv("MYSQL_CHARSET", "utf8mb4"),
		MySQLParseTime:          getEnv("MYSQL_PARSE_TIME", "true"),
		MySQLLoc:                getEnv("MYSQL_LOC", "Local"),
		JWTSecret:               getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHours:          getEnvInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenExpiryHours: getEnvInt("REFRESH_TOKEN_EXPIRY_HOURS", 168),
		BaiduMapAK:              getEnv("BAIDU_MAP_AK", ""),
		BaiduMapBaseURL:         getEnv("BAIDU_MAP_BASE_URL", "https://api.map.baidu.com"),
		OSSEndpoint:             getEnv("OSS_ENDPOINT", ""),
		OSSAccessKeyID:          getEnv("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret:      getEnv("OSS_ACCESS_KEY_SECRET", ""),
		OSSBucketName:           getEnv("OSS_BUCKET_NAME", ""),
		OSSBasePath:             getEnv("OSS_BASE_PATH", "avatars/"),
		AlibabaSearchBaseURL:    getEnv("ALIBABA_SEARCH_BASE_URL", "https://www.alibaba.com/search/api/supplierTextSearch"),
		MadeInChinaBaseURL:      getEnv("MADE_IN_CHINA_BASE_URL", "https://www.made-in-china.com"),
		GoogleAPIKey:            getEnv("GOOGLE_API_KEY", ""),
		GoogleCX:                getEnv("GOOGLE_CX", ""),
		GoogleSearchNum:         getEnvInt("GOOGLE_SEARCH_NUM", 10),
		GoogleProxyURL:          strings.TrimSpace(getEnv("GOOGLE_PROXY_URL", "")),
		SearchWorkerCount:       getEnvInt("EXTERNAL_COMPANY_SEARCH_WORKER_COUNT", 2),
		SearchPollIntervalMS:    getEnvInt("EXTERNAL_COMPANY_SEARCH_POLL_INTERVAL_MS", 1000),
	}
}

func (c Config) EffectiveMySQLDSN() string {
	if strings.TrimSpace(c.MySQLDSN) != "" {
		return strings.TrimSpace(c.MySQLDSN)
	}

	if strings.TrimSpace(c.MySQLHost) == "" || strings.TrimSpace(c.MySQLPort) == "" ||
		strings.TrimSpace(c.MySQLUser) == "" || strings.TrimSpace(c.MySQLDB) == "" {
		return ""
	}

	charset := strings.TrimSpace(c.MySQLCharset)
	if charset == "" {
		charset = "utf8mb4"
	}
	parseTime := strings.TrimSpace(c.MySQLParseTime)
	if parseTime == "" {
		parseTime = "true"
	}
	loc := strings.TrimSpace(c.MySQLLoc)
	if loc == "" {
		loc = "Local"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
		c.MySQLUser,
		c.MySQLPassword,
		c.MySQLHost,
		c.MySQLPort,
		c.MySQLDB,
		charset,
		parseTime,
		url.QueryEscape(loc),
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func (c Config) Validate() error {
	driver := strings.ToLower(strings.TrimSpace(c.DBDriver))
	switch driver {
	case "", "sqlite":
	case "mysql":
		if strings.TrimSpace(c.EffectiveMySQLDSN()) == "" {
			return fmt.Errorf("MYSQL_DSN or MYSQL_HOST/MYSQL_PORT/MYSQL_USER/MYSQL_DB must be set when DB_DRIVER=mysql")
		}
	default:
		return fmt.Errorf("unsupported DB_DRIVER: %s", c.DBDriver)
	}

	switch strings.ToLower(strings.TrimSpace(c.GormLogLevel)) {
	case "", "silent", "error", "warn", "info":
	default:
		return fmt.Errorf("unsupported GORM_LOG_LEVEL: %s (allowed: silent,error,warn,info)", c.GormLogLevel)
	}

	appEnv := strings.ToLower(strings.TrimSpace(c.AppEnv))
	if appEnv == "production" || appEnv == "prod" {
		if strings.TrimSpace(c.JWTSecret) == "" || c.JWTSecret == "change-me-in-production" {
			return fmt.Errorf("JWT_SECRET must be explicitly set in production")
		}
	}
	if c.JWTExpiryHours <= 0 {
		return fmt.Errorf("JWT_EXPIRY_HOURS must be greater than 0")
	}
	if c.RefreshTokenExpiryHours <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_EXPIRY_HOURS must be greater than 0")
	}
	if c.SearchWorkerCount <= 0 {
		return fmt.Errorf("EXTERNAL_COMPANY_SEARCH_WORKER_COUNT must be greater than 0")
	}
	if c.SearchPollIntervalMS <= 0 {
		return fmt.Errorf("EXTERNAL_COMPANY_SEARCH_POLL_INTERVAL_MS must be greater than 0")
	}
	if c.GoogleProxyURL != "" {
		parsed, err := url.Parse(c.GoogleProxyURL)
		if err != nil || strings.TrimSpace(parsed.Scheme) == "" || strings.TrimSpace(parsed.Host) == "" {
			return fmt.Errorf("GOOGLE_PROXY_URL must be a valid proxy URL")
		}
	}
	return nil
}
