package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"

	"github.com/yuwanadev/mdm-backend/internal/config"
	"github.com/yuwanadev/mdm-backend/internal/database"
	"github.com/yuwanadev/mdm-backend/internal/handler"
	"github.com/yuwanadev/mdm-backend/internal/middleware"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
	"github.com/yuwanadev/mdm-backend/internal/service"
	ws "github.com/yuwanadev/mdm-backend/internal/websocket"
)

func main() {
	// ── Load config ───────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ── Database ──────────────────────────────────────────────────
	pool, err := database.NewPool(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to PostgreSQL")

	// ── Repositories ──────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	deviceRepo := repository.NewDeviceRepo(pool)
	statusRepo := repository.NewStatusRepo(pool)
	commandRepo := repository.NewCommandRepo(pool)
	groupRepo := repository.NewGroupRepo(pool)
	apkRepo := repository.NewAPKRepo(pool)

	// ── Services ──────────────────────────────────────────────────
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	screenshotService := service.NewScreenshotService("storage/screenshots")
	deviceService := service.NewDeviceService(deviceRepo, statusRepo)
	commandService := service.NewCommandService(commandRepo, deviceRepo, screenshotService)
	groupService := service.NewGroupService(groupRepo)

	// ── Seed admin user ───────────────────────────────────────────
	if cfg.AdminPass != "" {
		if err := authService.SeedAdmin(context.Background(), cfg.AdminUser, cfg.AdminPass); err != nil {
			log.Fatalf("Failed to seed admin: %v", err)
		}
		log.Printf("✓ Admin user ready (%s)", cfg.AdminUser)
	}

	// ── WebSocket Hub ─────────────────────────────────────────────
	hub := ws.NewHub()
	hub.SetMessageHandler(func(deviceID uuid.UUID, msg *ws.WSMessage) {
		handleDeviceMessage(deviceID, msg, statusRepo, deviceRepo, commandService, hub)
	})
	hub.StartPingLoop(30 * time.Second)
	log.Println("✓ WebSocket hub started")

	// ── Handlers ──────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(deviceService, screenshotService)
	commandHandler := handler.NewCommandHandler(commandService, hub)
	statusHandler := handler.NewStatusHandler(statusRepo)
	groupHandler := handler.NewGroupHandler(groupService)
	apkHandler := handler.NewAPKHandler(apkRepo)
	wsHandler := ws.NewHandler(hub, deviceService, authService)

	// ── Fiber App ─────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "YuwanaDev MDM",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// ── Middleware ─────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS(cfg.CORSOrigins))

const backendVersion = "1.0.2"

	// ── Health check ──────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":            "ok",
			"version":           backendVersion,
			"connected_devices": hub.ConnectedDeviceCount(),
		})
	})

	// ── Public Assets ─────────────────────────────────────────────
	app.Get("/screenshots/:id", deviceHandler.Screenshot)

	// ── Auth routes (public) ──────────────────────────────────────
	auth := app.Group("/api/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// ── Protected API routes ──────────────────────────────────────
	api := app.Group("/api", middleware.JWTAuth(authService))

	// Devices
	api.Get("/devices", deviceHandler.List)
	api.Post("/devices", deviceHandler.Create)
	api.Get("/devices/:id", deviceHandler.Get)
	api.Put("/devices/:id", deviceHandler.Update)
	api.Delete("/devices/:id", deviceHandler.Delete)

	// Groups
	api.Get("/groups", groupHandler.List)
	api.Post("/groups", groupHandler.Create)
	api.Delete("/groups/:id", groupHandler.Delete)
	api.Post("/groups/:id/commands", commandHandler.BulkSend)

	// Device Status
	api.Get("/devices/:id/status", statusHandler.Get)

	// Commands
	api.Post("/devices/:id/commands", commandHandler.Send)
	api.Get("/devices/:id/commands", commandHandler.History)

	// APKs
	api.Get("/apks", apkHandler.List)
	api.Post("/apks", apkHandler.Upload)
	app.Get("/api/apks/:id", apkHandler.Download)

	// ── WebSocket routes ──────────────────────────────────────────
	app.Use("/ws/device", wsHandler.DeviceUpgradeCheck())
	app.Get("/ws/device", wsHandler.DeviceUpgrade())

	app.Use("/ws/dashboard", wsHandler.DashboardUpgradeCheck())
	app.Get("/ws/dashboard", wsHandler.DashboardUpgrade())

	// ── Graceful shutdown ─────────────────────────────────────────
	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		log.Printf("✓ Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ── Background Workers ────────────────────────────────────────
	startBackgroundWorkers(deviceService, commandService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	_ = app.Shutdown()
	pool.Close()
	log.Println("✓ Shutdown complete")
}

// handleDeviceMessage processes incoming messages from devices.
func handleDeviceMessage(
	deviceID uuid.UUID,
	msg *ws.WSMessage,
	statusRepo *repository.StatusRepo,
	deviceRepo *repository.DeviceRepo,
	commandService *service.CommandService,
	hub *ws.Hub,
) {
	ctx := context.Background()

	switch msg.Type {
	case ws.MsgHeartbeat:
		var payload ws.HeartbeatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("[MSG] Invalid heartbeat from %s: %v", deviceID, err)
			return
		}

		log.Printf("[HEARTBEAT] Device %s — battery=%v%%, app=%v", deviceID, payload.Battery, payload.ForegroundApp)

		// Update device status in DB
		status := &models.DeviceStatus{
			DeviceID:          deviceID,
			Battery:           payload.Battery,
			Temperature:       payload.Temperature,
			BatteryHealth:     payload.BatteryHealth,
			BatteryStatus:     payload.BatteryStatus,
			BatteryTechnology: payload.BatteryTechnology,
			BatteryVoltage:    payload.BatteryVoltage,
			RAMUsage:          payload.RAMUsage,
			StorageTotal:      payload.StorageTotal,
			StorageUsed:       payload.StorageUsed,
			AppVersion:        payload.AppVersion,
			NetworkInfo:       payload.NetworkInfo,
			ForegroundApp:     payload.ForegroundApp,
			NetworkStrength:   payload.NetworkStrength,
			Location:          payload.Location,
		}
		if err := statusRepo.Upsert(ctx, status); err != nil {
			log.Printf("[MSG] Failed to upsert status for %s: %v", deviceID, err)
		}

		// Update last_seen
		_ = deviceRepo.SetOnline(ctx, deviceID, true)

		// Forward to dashboards
		msg.DeviceID = deviceID.String()
		dashMsg, _ := ws.NewMessage(ws.MsgStatusUpdate, map[string]interface{}{
			"device_id":          deviceID.String(),
			"battery":            payload.Battery,
			"temperature":        payload.Temperature,
			"battery_health":     payload.BatteryHealth,
			"battery_status":     payload.BatteryStatus,
			"battery_technology": payload.BatteryTechnology,
			"battery_voltage":    payload.BatteryVoltage,
			"ram_usage":          payload.RAMUsage,
			"storage_used":       payload.StorageUsed,
			"app_version":        payload.AppVersion,
			"network_info":       payload.NetworkInfo,
			"foreground_app":     payload.ForegroundApp,
			"network_strength":   payload.NetworkStrength,
			"location":           payload.Location,
		})
		hub.BroadcastToDashboards(dashMsg)

	case ws.MsgCommandResult:
		var payload ws.CommandResultPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("[MSG] Invalid command result from %s: %v", deviceID, err)
			return
		}

		log.Printf("[CMD_RESULT] Device %s — cmd=%s success=%v msg=%s", deviceID, payload.CommandID, payload.Success, payload.Message)

		cmdID, err := uuid.Parse(payload.CommandID)
		if err != nil {
			log.Printf("[MSG] Invalid command ID from %s: %v", deviceID, err)
			return
		}

		_ = commandService.HandleResult(ctx, cmdID, payload.Success, payload.Message, payload.Data)

		// Forward result to dashboards
		dashMsg, _ := ws.NewMessage(ws.MsgCommandResult, map[string]interface{}{
			"device_id":  deviceID.String(),
			"command_id": payload.CommandID,
			"success":    payload.Success,
			"message":    payload.Message,
			"data":       payload.Data,
		})
		hub.BroadcastToDashboards(dashMsg)

	case ws.MsgDeviceInfo:
		var payload ws.DeviceInfoPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("[MSG] Invalid device info from %s: %v", deviceID, err)
			return
		}

		log.Printf("[DEVICE_INFO] Device %s — model=%s android=%s", deviceID, payload.Model, payload.AndroidVersion)

		_ = deviceRepo.UpdateDeviceInfo(ctx, deviceID, payload.Model, payload.AndroidVersion, payload.AppVersion)

		// Forward to dashboards
		dashMsg, _ := ws.NewMessage(ws.MsgDeviceInfo, map[string]interface{}{
			"device_id":       deviceID.String(),
			"model":           payload.Model,
			"manufacturer":    payload.Manufacturer,
			"android_version": payload.AndroidVersion,
			"agent_version":   payload.AppVersion,
		})
		hub.BroadcastToDashboards(dashMsg)

	case ws.MsgMirrorFrame:
		// Device is sending a screen frame — forward raw bytes to mirroring dashboard
		var frameData struct {
			Data string `json:"data"` // base64 encoded JPEG
		}
		if err := json.Unmarshal(msg.Payload, &frameData); err != nil {
			log.Printf("[MIRROR] ✗ Invalid mirror frame from %s: %v", deviceID, err)
			return
		}

		// Decode base64 to raw JPEG bytes
		rawBytes, err := base64Decode(frameData.Data)
		if err != nil {
			log.Printf("[MIRROR] ✗ Failed to decode mirror frame from %s: %v", deviceID, err)
			return
		}

		log.Printf("[MIRROR] ← Frame from device %s: %d bytes (b64: %d chars) → relaying to dashboard", deviceID, len(rawBytes), len(frameData.Data))

		// Send raw binary to the dashboard mirroring this device
		hub.SendMirrorFrame(deviceID, rawBytes)

	case ws.MsgWebRTCSignal:
		log.Printf("[WEBRTC] ← Signal from device %s → relaying to dashboards", deviceID)
		
		// Add device_id to payload so dashboard knows which device sent the signal
		var rawPayload map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &rawPayload); err == nil {
			dashMsg, _ := ws.NewMessage(ws.MsgWebRTCSignal, map[string]interface{}{
				"device_id": deviceID.String(),
				"signal":    rawPayload,
			})
			hub.BroadcastToDashboards(dashMsg)
		} else {
			log.Printf("[WEBRTC] ✗ Invalid signal payload from %s: %v", deviceID, err)
		}

	default:
		log.Printf("[MSG] Unknown message type '%s' from %s", msg.Type, deviceID)
	}
}

func startBackgroundWorkers(ds *service.DeviceService, cs *service.CommandService) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()

			// 1. Offline Detection (haven't seen heartbeat in 2 mins)
			if affected, err := ds.MarkOfflineInactive(ctx, 2*time.Minute); err == nil && affected > 0 {
				log.Printf("[Worker] Marked %d devices as offline", affected)
			}

			// 2. Command Timeout (sent more than 60s ago)
			if affected, err := cs.MarkTimeouts(ctx, 60*time.Second); err == nil && affected > 0 {
				log.Printf("[Worker] Timed out %d commands", affected)
			}
		}
	}()
	log.Println("✓ Background workers started")
}

func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
