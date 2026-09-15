package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/redis/go-redis/v9"

	"github.com/dormitory-bot/config"
	"github.com/dormitory-bot/internal/api"
	"github.com/dormitory-bot/internal/auth"
	"github.com/dormitory-bot/internal/bot"
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/dormitory-bot/internal/infrastructure/database"
	"github.com/dormitory-bot/internal/infrastructure/rabbitmq"
	"github.com/dormitory-bot/internal/mock"
	"github.com/dormitory-bot/internal/repository"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/timeutil"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrations(cfg); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	pgDB, err := database.ConnectPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := database.AutoMigrate(pgDB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	database.SeedRoles(pgDB)

	var mockProvider mock.EISDataProvider
	seedMode := os.Getenv("SEED_MODE")

	if cfg.MockMode {
		log.Println("=== MOCK MODE ENABLED ===")
		log.Println("  - RabbitMQ: SKIPPED")
		log.Println("  - EIS adapter: NOT REQUIRED")
		log.Println("  - EIS data: mock provider loaded")
		log.Println("  - Sync: mock data loaded from internal/mock/")

		mockProvider = mock.NewMockProvider()
		mock.LoadMockSyncData(pgDB)
	} else {
		if seedMode == "seed" || seedMode == "dev" || seedMode == "" {
			if seedMode == "" {
				seedMode = "dev"
			}
			log.Printf("Seeding test data with mode: %s", seedMode)
			// SeedDormitories FIRST (creates rooms needed by SeedTestUser residents),
			// then SeedTestUser (creates employee needed by SeedContent FK),
			// then SeedContent (chat_links, ref_materials needing employee FK)
			database.SeedDormitories(pgDB)
			database.SeedTestUser(pgDB, "79025643215", 143554557, "max")
			database.SeedContent(pgDB)
		}
	}

	// RabbitMQ — EIS gateway (async verify + sync via adapter pushers)
	// SKIPPED in mock mode
	var rmqClient *rabbitmq.EISClient
	var rmqErr error
	if !cfg.MockMode {
		rmqClient, rmqErr = rabbitmq.NewEISClient(cfg.RabbitMQ.URL, cfg.RabbitMQ.ExchangeName)
		if rmqErr != nil {
			log.Printf("Warning: RabbitMQ not available: %v", rmqErr)
		}
		if rmqClient != nil {
			log.Println("RabbitMQ connected successfully")
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Println("Redis connected successfully")

	redisCacheClient := cache.NewRedisClient(cfg.Redis)

	deps := &service.Dependencies{
		DB:          pgDB,
		RmqClient:   rmqClient,
		Redis:       redisCacheClient,
		EISProvider: mockProvider,
	}

	svc := service.New(deps)

	// Start RabbitMQ EIS consumers
	if svc.UniBroker != nil && rmqClient != nil {
		if err := rmqClient.SetupInfrastructure(); err != nil {
			log.Printf("Warning: RMQ infrastructure setup failed: %v", err)
		}
		rmqClient.StartAutoReconnect(ctx)

		go svc.UniBroker.ConsumeResidentSync(ctx)
		go svc.UniBroker.ConsumeEmployeeSync(ctx)
		go svc.UniBroker.ConsumeDormitorySync(ctx)

		// Trigger initial sync (async, with periodic resend for adapter recovery).
		go startSyncTriggerLoop(ctx, rmqClient)
	}

	timeutil.SetLocation(cfg.Timezone)
	service.SetReminderTimezone(cfg.Timezone)

	// Seed dormitory ID for cleaning schedule and mock launch token
	seedDormID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	if cfg.MockMode {
		seedDormID = mock.MockDormID8()
	}
	if cfg.MockMode || seedMode == "dev" || seedMode == "seed" {
		now := time.Now()
		var dutyCount int64
		pgDB.Model(&domain.CleaningDuty{}).Where("dormitory_id = ?", seedDormID).Count(&dutyCount)
		if dutyCount == 0 {
			if err := svc.CleaningService.GenerateSchedule(seedDormID, int(now.Month()), now.Year()); err != nil {
				log.Printf("Seed cleaning schedule error: %v", err)
			} else {
				log.Printf("Seed cleaning: generated schedule for %d.%d", now.Month(), now.Year())
			}
		}
	}

	// In mock mode, auto-generate a launch token so the frontend can log in
	// without needing the bot to send /start. Token is valid for 5 minutes,
	// but we log it so the dev can regenerate via /api/internal/dev/launch-token
	if cfg.MockMode {
		// Resolve the commandant user (created by mock.LoadMockSyncData)
		var cmdUser domain.User
		if err := pgDB.Where("phone = ?", "79025643215").First(&cmdUser).Error; err != nil {
			log.Printf("[mock] WARNING: commandant user 79025643215 not found in DB")
		} else {
			var cmdEmp domain.Employee
			pgDB.Where("user_id = ?", cmdUser.ID).First(&cmdEmp)
			token, err := svc.GenerateLaunchToken(cmdUser.ID.String(), cmdUser.PersonType,
				seedDormID.String(), cmdEmp.ID.String())
			if err != nil {
				log.Printf("[mock] WARNING: failed to generate launch token: %v", err)
			} else {
				log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				log.Printf("[mock] Launch token (valid 5 min): %s", token)
				log.Printf("[mock] Mini App URL: http://localhost:3000?launch_token=%s", token)
				log.Printf("[mock] To regenerate: curl -X POST http://localhost:8080/api/internal/dev/launch-token -H 'Content-Type: application/json' -d '{\"phone\":\"79025643215\"}'")
				log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			}
		}
	}

	fsmMgr := fsm.NewFSMManager(redisClient)

	maxAPI, err := maxbot.New(cfg.MAXBotToken)
	if err != nil {
		log.Fatalf("Failed to initialize MAX bot API: %v", err)
	}

	botInfo, err := maxAPI.Bots.GetBot(ctx)
	if err != nil {
		log.Printf("Warning: Could not get bot info: %v", err)
	} else {
		log.Printf("Bot connected as: %s (@%s)", botInfo.Name, botInfo.Username)
	}

	botInstance := bot.NewBot(fsmMgr, svc, maxAPI, cfg.MAXBotToken, "https://api.max.ru")
	log.Println("Bot instance created")

	// JWT service for Mini App API auth
	jwtSvc := auth.NewJWTService(cfg.JWTSecretOrEncryptionKey())
	log.Println("JWT service initialized")

	go func() {
		if err := botInstance.PollingLoop(ctx); err != nil {
			log.Printf("Polling loop error: %v", err)
		}
	}()

	app := api.NewRouter(svc, jwtSvc)
	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("Starting API server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("API server error: %v", err)
		}
	}()

	go startHealthServer()
	go startLaundryCleanup(ctx, svc)
	go startReminderPoller(ctx, botInstance)

	log.Println("Dormitory Bot running. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
	redisClient.Close()
	if rmqClient != nil {
		rmqClient.Close()
	}
}

func startHealthServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","service":"dormitory-bot"}`)
	})
	port := "8081"
	log.Printf("Starting health server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Health server error: %v", err)
	}
}

func startLaundryCleanup(ctx context.Context, svc *service.DBService) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.LaundryService.CleanupExpiredBookings(); err != nil {
				log.Printf("Laundry cleanup error: %v", err)
			}
		}
	}
}

func startReminderPoller(ctx context.Context, b *bot.Bot) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			members, err := b.GetPendingReminders(ctx)
			if err != nil {
				log.Printf("Reminder poll error: %v", err)
				continue
			}
			for _, member := range members {
				parts := splitReminder(member)
				if len(parts) < 3 {
					continue
				}
				reminderID, userIDStr, text := parts[0], parts[1], parts[2]
				_ = reminderID
				userID, parseErr := strconv.ParseInt(userIDStr, 10, 64)
				if parseErr != nil {
					continue
				}
				if err := b.SendTextToUser(ctx, userID, text); err != nil {
					log.Printf("Failed to send reminder to %d: %v", userID, err)
				}
			}
		}
	}
}

func splitReminder(member string) []string {
	parts := []string{}
	start := 0
	for i, c := range member {
		if c == '|' && len(parts) < 2 {
			parts = append(parts, member[start:i])
			start = i + 1
		}
	}
	parts = append(parts, member[start:])
	return parts
}

func runMigrations(cfg *config.Config) error {
	pgDB, err := database.ConnectPostgres(cfg.Database)
	if err != nil {
		return err
	}
	return database.AutoMigrate(pgDB)
}

// startSyncTriggerLoop sends MySQL credentials via RabbitMQ to the eis-adapter
// and resends periodically. Interval is controlled by SYNC_TRIGGER_INTERVAL env
// var (Go duration format, e.g. "4h", "30m", "1h30m"). Default: 4h.
func startSyncTriggerLoop(ctx context.Context, rmqClient *rabbitmq.EISClient) {
	interval := 4 * time.Hour
	if v := os.Getenv("SYNC_TRIGGER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
			log.Printf("[sync] Using SYNC_TRIGGER_INTERVAL=%s", v)
		} else {
			log.Printf("[sync] Invalid SYNC_TRIGGER_INTERVAL=%q: %v, using default 4h", v, err)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Immediate first attempt
	publishSyncTrigger(rmqClient)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			publishSyncTrigger(rmqClient)
		}
	}
}

// publishSyncTrigger sends MySQL credentials to the adapter.
// Retries up to 10 times with 5-second backoff — the adapter may not be running yet.
func publishSyncTrigger(rmqClient *rabbitmq.EISClient) {
	host := os.Getenv("UNIVERSITY_DB_HOST")
	if host == "" {
		log.Println("[sync] UNIVERSITY_DB_HOST not set, skipping sync trigger")
		return
	}
	port := 3306
	if p := os.Getenv("UNIVERSITY_DB_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	password := os.Getenv("UNIVERSITY_DB_PASSWORD")
	if password == "" {
		log.Println("[sync] UNIVERSITY_DB_PASSWORD not set, skipping sync trigger")
		return
	}

	trigger := rabbitmq.SyncTrigger{
		Version:       rabbitmq.ContractVersion,
		TriggerID:     uuid.New().String(),
		MySQLHost:     host,
		MySQLPort:     port,
		MySQLUser:     os.Getenv("UNIVERSITY_DB_USER"),
		MySQLPassword: password,
		MySQLDBName:   os.Getenv("UNIVERSITY_DB_DBNAME"),
	}

	if trigger.MySQLDBName == "" {
		trigger.MySQLDBName = "Irgups"
	}

	log.Printf("[sync] Sending sync trigger to adapter: %s:%d/%s", host, port, trigger.MySQLDBName)

	// Retry up to 10 attempts with 5s backoff — adapter may still be starting.
	for attempt := 1; attempt <= 10; attempt++ {
		if err := rmqClient.PublishJSON("irgups.sync.trigger", trigger); err != nil {
			log.Printf("[sync] Trigger attempt %d/10 failed: %v", attempt, err)
			if attempt < 10 {
				time.Sleep(5 * time.Second)
			}
			continue
		}
		log.Printf("[sync] Sync trigger sent successfully on attempt %d", attempt)
		return
	}
}

func init() {
	_ = domain.Models
	_ = repository.NewBaseRepository[*domain.User]
}
