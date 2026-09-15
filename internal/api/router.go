package api

import (
	"time"

	"github.com/dormitory-bot/internal/api/middleware"
	"github.com/dormitory-bot/internal/auth"
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"

	_ "github.com/dormitory-bot/docs"
)

func timeNow() time.Time { return time.Now() }

func NewRouter(svc *service.DBService, jwtSvc *auth.JWTService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: customErrorHandler})
	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Method() == "OPTIONS" { return c.SendStatus(204) }
		return c.Next()
	})

	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status":"ok","service":"dormitory-bot"}) })
	app.Get("/health/detailed", func(c *fiber.Ctx) error {
		rmqOK, redisOK := false, false
		if svc.UniBroker != nil { rmqOK = svc.UniBroker.IsRMQConnected(); redisOK = true }
		status := "ok"; if !rmqOK { status = "degraded" }
		return c.JSON(fiber.Map{"status":status,"rmq":rmqOK,"redis":redisOK})
	})

	api := app.Group("/api/v1")

	// ── Swagger meta ──
	// @title          Dormitory Bot API
	// @version        1.0
	// @description    REST API for dormitory management mini-app (MAX Messenger).
	// @host           localhost:8080
	// @BasePath       /api/v1
	// @securityDefinitions.apikey BearerAuth
	// @in             header
	// @name           Authorization

	// ═══════════════════════════════════════
	// Auth
	// ═══════════════════════════════════════
	auth := api.Group("/auth")

	// @Summary      Verify phone, get launch token
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Success      200  {object}  AuthVerifyResponse
	// @Failure      400  {object}  ErrorResponse
	// @Failure      404  {object}  ErrorResponse
	// @Router       /auth/verify [post]
	auth.Post("/verify", func(c *fiber.Ctx) error {
		type req struct{ Phone string `json:"phone"` }
		var r req
		if err := c.BodyParser(&r); err != nil || r.Phone == "" { return c.Status(400).JSON(fiber.Map{"error":"phone required"}) }
		result, err := svc.VerifyUser(r.Phone, 0, "miniapp")
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		if !result.Found { return c.Status(404).JSON(fiber.Map{"error":"user not found"}) }
		eid, did := "", result.DormitoryID
		if result.Employee != nil { eid = result.Employee.EmployeeID }
		token, err := svc.GenerateLaunchToken(result.UserID, result.PersonType, did, eid)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"launch_token":token,"user":fiber.Map{"user_id":result.UserID,"person_type":result.PersonType,"first_name":result.FirstName,"phone":r.Phone}})
	})

	// @Summary      Mini-app login (exchange launch token for JWT)
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Success      200  {object}  AuthMiniappResponse
	// @Failure      401  {object}  ErrorResponse
	// @Router       /auth/miniapp [post]
	auth.Post("/miniapp", func(c *fiber.Ctx) error {
		type req struct{ LaunchToken string `json:"launch_token"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		token, profile, err := svc.AuthorizeMiniApp(r.LaunchToken, jwtSvc)
		if err != nil { return c.Status(401).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"token":token,"user":profile})
	})

	// @Summary      Current user from JWT
	// @Tags         Auth
	// @Security     BearerAuth
	// @Produce      json
	// @Success      200  {object}  AuthMeResponse
	// @Router       /auth/me [get]
	auth.Get("/me", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		uid, _ := uuid.Parse(c.Locals("user_id").(string))
		user, err := svc.GetUserByUserID(uid)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}

		resp := fiber.Map{
			"user_id":     uid,
			"person_type": c.Locals("person_type").(string),
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"middle_name": user.MiddleName,
		}

		// Если у пользователя есть привязка к общежитию (житель), добавляем
		// информацию о договоре и общежитии.
		resident, err := svc.GetResidentByUserID(uid)
		if err == nil && resident != nil {
			resp["contract_number"]     = resident.ContractNumber
			resp["contract_start_date"] = resident.ContractStartDate
			resp["contract_end_date"]   = resident.ContractEndDate
			resp["dormitory"]           = &resident.Dormitory.Name
			resp["room_number"]         = &resident.Room.RoomNumber
		}

		return c.JSON(resp)
	})

	// ═══════════════════════════════════════
	// Dormitories
	// ═══════════════════════════════════════
	dormitories := api.Group("/dormitories")

	// @Summary      List all active dormitories
	// @Tags         Dormitories
	// @Produce      json
	// @Success      200  {object}  DormitoryListResponse
	// @Router       /dormitories [get]
	dormitories.Get("/", func(c *fiber.Ctx) error {
		dorms, err := svc.GetDormitoryRepo().FindAllByField("is_active", true)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		if dorms == nil { dorms = []domain.Dormitory{} }
		return c.JSON(fiber.Map{"data":dorms})
	})

	// @Summary      Get dormitory by UUID
	// @Tags         Dormitories
	// @Produce      json
	// @Param        id  path  string  true  "Dormitory UUID"
	// @Success      200 {object} DormitoryItem
	// @Failure      404 {object} ErrorResponse
	// @Router       /dormitories/{id} [get]
	dormitories.Get("/:id", func(c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid id"}) }
		dorm, err := svc.GetDormitoryRepo().GetByID(id)
		if err != nil { return c.Status(404).JSON(fiber.Map{"error":"dormitory not found"}) }
		return c.JSON(dorm)
	})

	// @Summary      Floors of a dormitory
	// @Tags         Dormitories
	// @Produce      json
	// @Param        id  path  string  true  "Dormitory UUID"
	// @Success      200 {object} FloorListResponse
	// @Router       /dormitories/{id}/floors [get]
	dormitories.Get("/:id/floors", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		floors, _ := svc.ResidentService.GetFloorsByDormitory(id)
		return c.JSON(floors)
	})

	// @Summary      Staff of a dormitory
	// @Tags         Dormitories
	// @Produce      json
	// @Param        id  path  string  true  "Dormitory UUID"
	// @Success      200 {object} StaffListResponse
	// @Router       /dormitories/{id}/staff [get]
	dormitories.Get("/:id/staff", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		staff, err := svc.GetStaffByDormitory(id)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":staff})
	})

	// ═══════════════════════════════════════
	// Floors
	// ═══════════════════════════════════════
	floors := api.Group("/floors")

	// @Summary      Rooms on a floor
	// @Tags         Rooms
	// @Produce      json
	// @Param        id  path  string  true  "Floor UUID"
	// @Success      200 {object} RoomListResponse
	// @Router       /floors/{id}/rooms [get]
	floors.Get("/:id/rooms", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		rooms, _ := svc.ResidentService.GetRoomsByFloor(id)
		return c.JSON(rooms)
	})

	// ═══════════════════════════════════════
	// Rooms
	// ═══════════════════════════════════════
	rooms := api.Group("/rooms")

	// @Summary      Search rooms by number
	// @Tags         Rooms
	// @Produce      json
	// @Param        q  query  string  true  "Room number substring"
	// @Success      200 {object} RoomListResponse
	// @Failure      400 {object} ErrorResponse
	// @Router       /rooms/search [get]
	rooms.Get("/search", func(c *fiber.Ctx) error {
		q := c.Query("q")
		if q == "" { return c.Status(400).JSON(fiber.Map{"error":"query parameter 'q' required"}) }
		results, err := svc.RoomService.SearchRooms(q)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":results})
	})

	// @Summary      Get room by UUID
	// @Tags         Rooms
	// @Produce      json
	// @Param        id  path  string  true  "Room UUID"
	// @Success      200 {object} RoomItem
	// @Router       /rooms/{id} [get]
	rooms.Get("/:id", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		room, _ := svc.RoomService.GetRoomInfo(id)
		return c.JSON(room)
	})

	// @Summary      Room material items
	// @Tags         Rooms
	// @Produce      json
	// @Param        id  path  string  true  "Room UUID"
	// @Success      200 {object} MaterialItemListResponse
	// @Router       /rooms/{id}/items [get]
	rooms.Get("/:id/items", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		items, _ := svc.RoomService.GetRoomItems(id)
		return c.JSON(items)
	})

	// @Summary      Add material item to room
	// @Tags         Rooms
	// @Security     BearerAuth
	// @Accept       json
	// @Produce      json
	// @Param        id    path   string  true  "Room UUID"
	// @Param        body  body   object{name=string,inventory_number=string,quantity=int,condition=string}  true  "Item data"
	// @Success      200   {object} MaterialItemResponse
	// @Router       /rooms/{id}/items [post]
	rooms.Post("/:id/items", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		type req struct{ ItemName, InvNumber, Condition string; Quantity int `json:"quantity"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(c.Params("id"))
		item, err := svc.RoomService.AddRoomItem(rid, r.ItemName, r.InvNumber, r.Quantity, r.Condition)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":item})
	})

	// @Summary      Residents of a room
	// @Tags         Rooms
	// @Produce      json
	// @Param        id  path  string  true  "Room UUID"
	// @Success      200 {object} RoomResidentsResponse
	// @Router       /rooms/{id}/residents [get]
	rooms.Get("/:id/residents", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		residents, _ := svc.ResidentService.GetRoomWithResidents(id)
		return c.JSON(residents)
	})

	// ═══════════════════════════════════════
	// Roles
	// ═══════════════════════════════════════
	roles := api.Group("/roles")

	// @Summary      All system and custom roles
	// @Tags         Roles
	// @Produce      json
	// @Success      200 {object} RoleListResponse
	// @Router       /roles [get]
	roles.Get("/", func(c *fiber.Ctx) error {
		all, err := svc.ListRoles()
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":all})
	})

	// ═══════════════════════════════════════
	// Residents
	// ═══════════════════════════════════════
	residents := api.Group("/residents")

	// @Summary      Resident profile
	// @Tags         Residents
	// @Security     BearerAuth
	// @Produce      json
	// @Param        id  path  string  true  "Resident UUID"
	// @Success      200 {object} ResidentProfileResponse
	// @Router       /residents/{id} [get]
	residents.Get("/:id", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		profile, _ := svc.ResidentService.GetFullProfile(id)
		return c.JSON(profile)
	})

	// @Summary      Resident financial debt
	// @Tags         Residents
	// @Produce      json
	// @Param        id  path  string  true  "Resident UUID"
	// @Success      200 {object} ResidentDebtResponse
	// @Router       /residents/{id}/debt [get]
	residents.Get("/:id/debt", func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		debts, _ := svc.ResidentService.GetDebt(id)
		return c.JSON(debts)
	})

	// ═══════════════════════════════════════
	// Employees
	// ═══════════════════════════════════════
	employees := api.Group("/employees")

	// @Summary      Employee profile
	// @Tags         Employees
	// @Security     BearerAuth
	// @Produce      json
	// @Param        id  path  string  true  "Employee UUID"
	// @Success      200 {object} EmployeeInfoResponse
	// @Router       /employees/{id} [get]
	employees.Get("/:id", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		empInfo, err := svc.InitEmployee(c.Params("id"))
		if err != nil { return c.Status(404).JSON(fiber.Map{"error":"employee not found"}) }
		return c.JSON(fiber.Map{"data":empInfo})
	})

	// @Summary      Employee's assigned dormitories
	// @Tags         Employees
	// @Security     BearerAuth
	// @Produce      json
	// @Param        id  path  string  true  "Employee UUID"
	// @Success      200 {object} EmployeeDormitoriesResponse
	// @Router       /employees/{id}/dormitories [get]
	employees.Get("/:id/dormitories", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		dorms, err := svc.GetEmployeeDormitories(c.Params("id"))
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":dorms})
	})

	// ═══════════════════════════════════════
	// Laundry
	// ═══════════════════════════════════════
	laundry := api.Group("/laundry")

	// @Summary      Available laundry time slots
	// @Tags         Laundry
	// @Produce      json
	// @Param        dormitory_id  query  string  true   "Dormitory UUID"
	// @Param        date          query  string  false  "Date YYYY-MM-DD"
	// @Success      200 {object} LaundrySlotListResponse
	// @Router       /laundry/slots [get]
	laundry.Get("/slots", func(c *fiber.Ctx) error {
		did := c.Query("dormitory_id")
		date := c.Query("date", timeNow().Format("2006-01-02"))
		if did == "" { return c.Status(400).JSON(fiber.Map{"error":"dormitory_id required"}) }
		id, err := uuid.Parse(did)
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid dormitory_id"}) }
		slots, err := svc.LaundryService.GetAvailableSlots(id, date)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":slots})
	})

	// @Summary      Book a laundry slot
	// @Tags         Laundry
	// @Security     BearerAuth
	// @Accept       json
	// @Produce      json
	// @Param        body  body  object{machine_id=string,date=string,slot_start=string,slot_end=string}  true  "Booking"
	// @Success      200 {object} LaundryBookingResponse
	// @Failure      404 {object} ErrorResponse
	// @Failure      409 {object} ErrorResponse
	// @Router       /laundry/bookings [post]
	laundry.Post("/bookings", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		type req struct {
			MachineID string `json:"machine_id"`
			Date      string `json:"date"`
			SlotStart string `json:"slot_start"`
			SlotEnd   string `json:"slot_end"`
		}
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		uid, _ := uuid.Parse(c.Locals("user_id").(string))
		mid, _ := uuid.Parse(r.MachineID)
		result, err := svc.LaundryService.CreateBooking(uid, "resident", mid, r.Date, r.SlotStart, r.SlotEnd)
		if err != nil { return c.Status(409).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":result})
	})

	// @Summary      Cancel laundry booking
	// @Tags         Laundry
	// @Security     BearerAuth
	// @Produce      json
	// @Param        id  path  string  true  "Booking UUID"
	// @Success      200 {object} MessageResponse
	// @Router       /laundry/bookings/{id} [delete]
	laundry.Delete("/bookings/:id", middleware.JWTAuthMiddleware(jwtSvc), func(c *fiber.Ctx) error {
		bid, _ := uuid.Parse(c.Params("id"))
		uid, _ := uuid.Parse(c.Locals("user_id").(string))
		if err := svc.LaundryService.CancelBooking(bid, uid); err != nil { return c.Status(409).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"message":"booking cancelled"})
	})

	// @Summary      Washing machines in a dormitory
	// @Tags         Laundry
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} WashingMachineListResponse
	// @Router       /laundry/machines [get]
	laundry.Get("/machines", func(c *fiber.Ctx) error {
		did := c.Query("dormitory_id")
		if did == "" { return c.Status(400).JSON(fiber.Map{"error":"dormitory_id required"}) }
		id, err := uuid.Parse(did)
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid dormitory_id"}) }
		machines, err := svc.LaundryService.GetWashingMachines(id)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":machines})
	})

	laundry.Post("/machines", middleware.RequirePermissionJWT(svc, "laundry", "create"), func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message":"create machine"}) })

	// @Summary      Laundry settings for a dormitory
	// @Tags         Laundry
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} LaundrySettingsResponse
	// @Router       /laundry/settings [get]
	laundry.Get("/settings", func(c *fiber.Ctx) error {
		did := c.Query("dormitory_id")
		if did == "" { return c.Status(400).JSON(fiber.Map{"error":"dormitory_id required"}) }
		id, err := uuid.Parse(did)
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid dormitory_id"}) }
		settings, err := svc.LaundryService.GetLaundrySettings(id)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":settings})
	})

	// ═══════════════════════════════════════
	// Cleaning
	// ═══════════════════════════════════════
	cleaning := api.Group("/cleaning")

	// @Summary      Cleaning schedule
	// @Tags         Cleaning
	// @Produce      json
	// @Param        dormitory_id  query  string  true   "Dormitory UUID"
	// @Param        month         query  int     false  "Month 1-12"
	// @Param        year          query  int     false  "Year"
	// @Success      200 {object} CleaningScheduleResponse
	// @Router       /cleaning/schedule [get]
	cleaning.Get("/schedule", func(c *fiber.Ctx) error {
		did, _ := uuid.Parse(c.Query("dormitory_id"))
		month := c.QueryInt("month", int(timeNow().Month()))
		year := c.QueryInt("year", timeNow().Year())
		duties, err := svc.CleaningService.GetDormitorySchedule(did, month, year)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":duties})
	})

	cleaning.Post("/generate", middleware.RequirePermissionJWT(svc, "cleaning", "create"), func(c *fiber.Ctx) error {
		type req struct {
			DormitoryID string `json:"dormitory_id"`
			Month       int    `json:"month"`
			Year        int    `json:"year"`
		}
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		did, _ := uuid.Parse(r.DormitoryID)
		if err := svc.CleaningService.GenerateSchedule(did, r.Month, r.Year); err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"message":"schedule generated"})
	})

	// @Summary      Penalty cleaning duties
	// @Tags         Cleaning
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} CleaningPenaltyResponse
	// @Router       /cleaning/penalties [get]
	cleaning.Get("/penalties", func(c *fiber.Ctx) error {
		did, _ := uuid.Parse(c.Query("dormitory_id"))
		penalties, err := svc.CleaningService.GetPenaltiesByDormitory(did)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":penalties})
	})

	cleaning.Post("/penalties", middleware.RequirePermissionJWT(svc, "cleaning", "create"), func(c *fiber.Ctx) error {
		type req struct{ RoomID, Reason, IssuedBy string; Count int `json:"count"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(r.RoomID); issuedBy, _ := uuid.Parse(r.IssuedBy)
		penalty, err := svc.CleaningService.CreatePenalty(rid, nil, r.Count, r.Reason, issuedBy)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":penalty})
	})

	cleaning.Delete("/penalties/:id", middleware.RequirePermissionJWT(svc, "cleaning", "create"), func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		if err := svc.CleaningService.DeletePenalty(id); err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"message":"penalty deleted"})
	})

	// @Summary      Cleaning duty exemptions
	// @Tags         Cleaning
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} CleaningExemptionResponse
	// @Router       /cleaning/exemptions [get]
	cleaning.Get("/exemptions", func(c *fiber.Ctx) error {
		did, _ := uuid.Parse(c.Query("dormitory_id"))
		exemptions, err := svc.CleaningService.GetExemptionsByDormitory(did)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":exemptions})
	})

	cleaning.Post("/exemptions", middleware.RequirePermissionJWT(svc, "cleaning", "create"), func(c *fiber.Ctx) error {
		type req struct {
			RoomID    string `json:"room_id"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
			Reason    string `json:"reason"`
			CreatedBy string `json:"created_by"`
		}
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(r.RoomID); createdBy, _ := uuid.Parse(r.CreatedBy)
		exemption, err := svc.CleaningService.CreateExemption(rid, r.StartDate, r.EndDate, r.Reason, createdBy)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":exemption})
	})

	// ═══════════════════════════════════════
	// Reference materials
	// ═══════════════════════════════════════
	reference := api.Group("/reference-materials")

	// @Summary      Reference materials for a dormitory
	// @Tags         References
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} ReferenceMaterialListResponse
	// @Router       /reference-materials [get]
	reference.Get("/", func(c *fiber.Ctx) error {
		did := c.Query("dormitory_id")
		if did == "" { return c.Status(400).JSON(fiber.Map{"error":"dormitory_id required"}) }
		id, err := uuid.Parse(did)
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid dormitory_id"}) }
		materials, err := svc.ReferenceService.GetByDormitory(id)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		if materials == nil { materials = []domain.ReferenceMaterial{} }
		return c.JSON(fiber.Map{"data":materials})
	})

	reference.Post("/", middleware.RequirePermissionJWT(svc, "reference", "create"), func(c *fiber.Ctx) error {
		type req struct{ DormitoryID, Name, Description, Category string; Ordinal int `json:"ordinal"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		did, _ := uuid.Parse(r.DormitoryID)
		createdBy, _ := c.Locals("employee_id").(string)
		cbid, _ := uuid.Parse(createdBy)
		material := &domain.ReferenceMaterial{DormitoryID:did, Name:r.Name, Description:r.Description, Category:r.Category, Ordinal:r.Ordinal, CreatedBy:cbid}
		if err := svc.ReferenceService.Create(material); err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":material})
	})

	// ═══════════════════════════════════════
	// Chat links
	// ═══════════════════════════════════════
	chatLinks := api.Group("/chat-links")

	// @Summary      Chat links for a dormitory
	// @Tags         ChatLinks
	// @Produce      json
	// @Param        dormitory_id  query  string  true  "Dormitory UUID"
	// @Success      200 {object} ChatLinkListResponse
	// @Router       /chat-links [get]
	chatLinks.Get("/", func(c *fiber.Ctx) error {
		did := c.Query("dormitory_id")
		if did == "" { return c.Status(400).JSON(fiber.Map{"error":"dormitory_id required"}) }
		id, err := uuid.Parse(did)
		if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid dormitory_id"}) }
		links, err := svc.ChatLinkService.List(id)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		if links == nil { links = []domain.ChatLink{} }
		return c.JSON(fiber.Map{"data":links})
	})

	chatLinks.Post("/", middleware.RequirePermissionJWT(svc, "chat_link", "create"), func(c *fiber.Ctx) error {
		type req struct{ DormitoryID, LinkType, Platform, URL, Title string; FloorID *string `json:"floor_id"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		did, _ := uuid.Parse(r.DormitoryID)
		createdBy, _ := c.Locals("employee_id").(string)
		cbid, _ := uuid.Parse(createdBy)
		link := &domain.ChatLink{DormitoryID:did, LinkType:r.LinkType, Platform:r.Platform, URL:r.URL, Title:r.Title, IsActive:true, CreatedBy:cbid}
		if r.FloorID != nil { fid, _ := uuid.Parse(*r.FloorID); link.FloorID = &fid }
		if err := svc.ChatLinkService.Create(link); err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":link})
	})

	chatLinks.Delete("/:id", middleware.RequirePermissionJWT(svc, "chat_link", "delete"), func(c *fiber.Ctx) error {
		id, _ := uuid.Parse(c.Params("id"))
		if err := svc.ChatLinkService.Delete(id); err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"message":"chat link deleted"})
	})

	// ═══════════════════════════════════════
	// Controls
	// ═══════════════════════════════════════
	controls := api.Group("/controls")
	controls.Post("/sanitary", middleware.RequirePermissionJWT(svc, "room", "create"), func(c *fiber.Ctx) error {
		type req struct{ RoomID, StartDate, EndDate, Reason string }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(r.RoomID); inspStr, _ := c.Locals("employee_id").(string); insp, _ := uuid.Parse(inspStr)
		ctrl, err := svc.ControlService.SetSanitaryControl(rid, r.StartDate, r.EndDate, r.Reason, insp)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":ctrl})
	})
	controls.Post("/discipline", middleware.RequirePermissionJWT(svc, "room", "create"), func(c *fiber.Ctx) error {
		type req struct{ RoomID, StartDate, EndDate, Violation string; ResidentID *string `json:"resident_id"` }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(r.RoomID); respStr, _ := c.Locals("employee_id").(string); resp, _ := uuid.Parse(respStr)
		var residentID *uuid.UUID
		if r.ResidentID != nil { uid, _ := uuid.Parse(*r.ResidentID); residentID = &uid }
		ctrl, err := svc.ControlService.SetDisciplineControl(rid, residentID, r.StartDate, r.EndDate, r.Violation, resp)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":ctrl})
	})
	controls.Post("/cohabitation", middleware.RequirePermissionJWT(svc, "room", "create"), func(c *fiber.Ctx) error {
		type req struct{ RoomID, StartDate, EndDate, Reason string }
		var r req
		if err := c.BodyParser(&r); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }
		rid, _ := uuid.Parse(r.RoomID)
		ctrl, err := svc.ControlService.SetCohabitationControl(rid, r.StartDate, r.EndDate, r.Reason)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"data":ctrl})
	})

	// ═══════════════════════════════════════
	// Modules / Internal (admin only, not for miniapp)
	// ═══════════════════════════════════════
	modules := api.Group("/modules")
	modules.Get("/", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message":"bot modules"}) })
	modules.Put("/:id/token", middleware.RequirePermissionJWT(svc, "module", "update"), func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message":"update token"}) })
	modules.Post("/:id/control", middleware.RequirePermissionJWT(svc, "module", "create"), func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message":"control bot"}) })

	internal := app.Group("/api/internal")
	internal.Post("/eis/verify", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message":"eis verify"}) })

	internal.Post("/dev/launch-token", func(c *fiber.Ctx) error {
		type req struct{ Phone string `json:"phone"` }
		var r req
		if err := c.BodyParser(&r); err != nil || r.Phone == "" { return c.Status(400).JSON(fiber.Map{"error":"phone required"}) }
		result, err := svc.VerifyUser(r.Phone, 0, "max")
		if err != nil || !result.Found { return c.Status(404).JSON(fiber.Map{"error":"user not found"}) }
		eid, did := "", result.DormitoryID
		if result.Employee != nil { eid = result.Employee.EmployeeID }
		token, err := svc.GenerateLaunchToken(result.UserID, result.PersonType, did, eid)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"launch_token":token,"mini_app_url":"http://localhost:3000?launch_token="+token,"user":fiber.Map{"user_id":result.UserID,"person_type":result.PersonType,"first_name":result.FirstName,"phone":r.Phone}})
	})

	internal.Get("/dev/auth", func(c *fiber.Ctx) error {
		role, phone := c.Query("role"), c.Query("phone")
		if phone == "" { phone = devRolePhone(role) }
		if phone == "" { return c.Status(400).JSON(fiber.Map{"error":"role or phone required"}) }
		if role == "" { role = devRoleDetect(phone) }
		result, err := svc.VerifyUser(phone, 0, "max")
		if err != nil || !result.Found { return c.Status(404).JSON(fiber.Map{"error":"test user not found in DB","hint":"run with MOCK_MODE=true"}) }
		eid := ""
		if result.Employee != nil { eid = result.Employee.EmployeeID }
		token, err := jwtSvc.IssueTokenWithTTL(result.UserID, result.PersonType, result.DormitoryID, eid, 30*24*time.Hour)
		if err != nil { return c.Status(500).JSON(fiber.Map{"error":err.Error()}) }
		return c.JSON(fiber.Map{"token":token,"user":fiber.Map{"user_id":result.UserID,"person_type":result.PersonType,"first_name":result.FirstName,"role":role,"dormitory_id":result.DormitoryID,"employee_id":eid}})
	})

	internal.Get("/dev/users", func(c *fiber.Ctx) error {
		users := buildDevUserList()
		if len(users.Employ) == 0 && len(users.Resident) == 0 { return c.Status(404).JSON(fiber.Map{"error":"no mock users available"}) }
		return c.JSON(users)
	})

	return app
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok { code = e.Code }
	return c.Status(code).JSON(fiber.Map{"success":false,"error":err.Error(),"code":code})
}

func devRolePhone(role string) string {
	switch role {
	case "employ": return "79025643215"
	case "resident": return "79001234501"
	case "stud_council": return "79140001144"
	default: return ""
	}
}

func devRoleDetect(phone string) string {
	switch phone {
	case "79025643215","79140001122","79140001133": return "employ"
	case "79140001144","79140001155": return "stud_council"
	default: return "resident"
	}
}

type devUserList struct {
	Employ      []devUserCard `json:"employ"`
	StudCouncil []devUserCard `json:"stud_council"`
	Resident    []devUserCard `json:"resident"`
}
type devUserCard struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
	Debt  string `json:"debt,omitempty"`
	Role  string `json:"role"`
}

func buildDevUserList() devUserList {
	return devUserList{
		Employ: []devUserCard{
			{Phone:"79025643215",Name:"Данил Высоких",Title:"Заведующий (общ. №8)",Role:"employ"},
			{Phone:"79140001122",Name:"Елена Кузнецова",Title:"Дежурный (общ. №8)",Role:"employ"},
			{Phone:"79140001133",Name:"Игорь Морозов",Title:"Заведующий (общ. №5)",Role:"employ"},
		},
		StudCouncil: []devUserCard{
			{Phone:"79140001144",Name:"Анна Фёдорова",Title:"Председатель студсовета",Role:"stud_council"},
			{Phone:"79140001155",Name:"Павел Григорьев",Title:"Староста",Role:"stud_council"},
		},
		Resident: []devUserCard{
			{Phone:"79001234501",Name:"Артём Иванов",Debt:"1500 ₽",Role:"resident"},
			{Phone:"79001234502",Name:"Максим Петров",Debt:"—",Role:"resident"},
			{Phone:"79001234503",Name:"Анна Смирнова",Debt:"3200 ₽",Role:"resident"},
			{Phone:"79001234504",Name:"Дмитрий Кузнецов",Debt:"—",Role:"resident"},
			{Phone:"79001234505",Name:"Алексей Попов",Debt:"—",Role:"resident"},
			{Phone:"79001234506",Name:"Никита Соколов",Debt:"5000 ₽",Role:"resident"},
			{Phone:"79001234507",Name:"Ольга Васильева",Debt:"—",Role:"resident"},
			{Phone:"79001234508",Name:"Сергей Белов",Debt:"—",Role:"resident"},
			{Phone:"79001234509",Name:"Ирина Морозова",Debt:"—",Role:"resident"},
			{Phone:"79001234510",Name:"Владимир Ершов",Debt:"—",Role:"resident"},
		},
	}
}
