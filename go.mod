module github.com/tuanle96/agentos-ecosystem

go 1.21

require (
	// Web Framework & API
	github.com/gin-gonic/gin v1.9.1
	github.com/gorilla/websocket v1.5.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/swag v1.16.2

	// Database & ORM
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	github.com/golang-migrate/migrate/v4 v4.16.2

	// Redis & Caching
	github.com/redis/go-redis/v9 v9.3.0
	github.com/hibiken/asynq v0.24.1

	// Authentication & Security
	github.com/golang-jwt/jwt/v5 v5.2.0
	golang.org/x/crypto v0.17.0
	github.com/google/uuid v1.4.0

	// Configuration & Environment
	github.com/spf13/viper v1.17.0
	github.com/joho/godotenv v1.4.0

	// Logging & Monitoring
	github.com/sirupsen/logrus v1.9.3
	github.com/prometheus/client_golang v1.17.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0

	// HTTP Client & Utilities
	github.com/go-resty/resty/v2 v2.10.0
	github.com/tidwall/gjson v1.17.0

	// AI & LLM Integration
	github.com/sashabaranov/go-openai v1.17.9
	github.com/anthropics/anthropic-sdk-go v0.1.0

	// Vector Database Clients
	github.com/pinecone-io/go-pinecone v0.3.0
	github.com/weaviate/weaviate-go-client/v4 v4.13.1

	// Message Queue & Event Streaming
	github.com/segmentio/kafka-go v0.4.47
	github.com/nats-io/nats.go v1.31.0

	// Testing
	github.com/stretchr/testify v1.8.4
	github.com/golang/mock v1.6.0
	github.com/testcontainers/testcontainers-go v0.26.0

	// Development Tools
	github.com/air-verse/air v1.49.0
	github.com/golangci/golangci-lint v1.55.2
)

require (
	// Indirect dependencies will be managed by Go modules
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)