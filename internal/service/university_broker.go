package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/dormitory-bot/internal/infrastructure/rabbitmq"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

const (
	// Separate Redis dedup namespaces per sync type to prevent cross-type
	// key collisions (e.g. residents sync occupying the employees dedup key).
	syncIDPrefixResidents   = "sync:id:res:"
	syncIDPrefixEmployees   = "sync:id:emp:"
	syncIDPrefixDormitories = "sync:id:dorm:"
	syncIDTTL               = 24 * time.Hour
)

// UniversityBroker is the bot-side bridge to the EIS via RabbitMQ.
type UniversityBroker struct {
	rmq            *rabbitmq.EISClient
	redis          *cache.RedisClient
	userRepo       *repository.BaseRepository[domain.User]
	residentRepo   *repository.BaseRepository[domain.Resident]
	employeeRepo   *repository.BaseRepository[domain.Employee]
	dormitoryRepo  *repository.BaseRepository[domain.Dormitory]
	roomRepo       *repository.BaseRepository[domain.Room]
	floorRepo      *repository.BaseRepository[domain.Floor]
	employeeDRRepo *repository.EmployeeDormitoryRoleRepository
}

// NewUniversityBroker creates a new broker. rmq may be nil (RMQ disabled).
func NewUniversityBroker(
	rmq *rabbitmq.EISClient,
	redis *cache.RedisClient,
	userRepo *repository.BaseRepository[domain.User],
	residentRepo *repository.BaseRepository[domain.Resident],
	employeeRepo *repository.BaseRepository[domain.Employee],
	dormitoryRepo *repository.BaseRepository[domain.Dormitory],
	roomRepo *repository.BaseRepository[domain.Room],
	floorRepo *repository.BaseRepository[domain.Floor],
	employeeDRRepo *repository.EmployeeDormitoryRoleRepository,
) *UniversityBroker {
	return &UniversityBroker{
		rmq:            rmq,
		redis:          redis,
		userRepo:       userRepo,
		residentRepo:   residentRepo,
		employeeRepo:   employeeRepo,
		dormitoryRepo:  dormitoryRepo,
		roomRepo:       roomRepo,
		floorRepo:      floorRepo,
		employeeDRRepo: employeeDRRepo,
	}
}

// IsRMQConnected returns true if the underlying RMQ connection is alive.
func (b *UniversityBroker) IsRMQConnected() bool {
	if b.rmq == nil {
		return false
	}
	return b.rmq.IsConnected()
}

// ── Verify (async) ──────────────────────────────────

// VerifyUserAsync publishes a verification request and stores pending state.
// Returns nil on success. The caller should show "⏳ Checking..." to user.
// VerifyUserRPC performs a synchronous EIS lookup via RabbitMQ RPC.
// Returns (person, employee, student, contract) or error on timeout/connect failure.
func (b *UniversityBroker) VerifyUserRPC(ctx context.Context, phone string) (*rabbitmq.PersonInfo, *rabbitmq.EmployeeInfo, *rabbitmq.StudentInfo, *rabbitmq.ContractInfo, error) {
	if b.rmq == nil || !b.IsRMQConnected() {
		return nil, nil, nil, nil, fmt.Errorf("RabbitMQ not available")
	}

	msg := rabbitmq.VerifyRequest{
		Version:       rabbitmq.ContractVersion,
		CorrelationID: uuid.New().String(),
		Phone:         phone,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	body, err := b.rmq.PublishRPC("irgups.verify.request", msg, 10*time.Second)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("RPC verify failed: %w", err)
	}

	var resp rabbitmq.VerifyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("decode verify response: %w", err)
	}

	if !resp.Found || resp.Person == nil {
		return nil, nil, nil, nil, nil
	}

	return resp.Person, resp.Employee, resp.Student, resp.Contract, nil
}

// ── Sync: Residents ──────────────────────────────────

// ConsumeResidentSync consumes resident sync pushes from the adapter.
func (b *UniversityBroker) ConsumeResidentSync(ctx context.Context) {
	if b.rmq == nil {
		return
	}
	handler := func(ctx context.Context, td rabbitmq.TypedDelivery[rabbitmq.SyncResidentsPush]) {
		if b.applyResidentSync(ctx, &td.Msg) {
			td.Nack(true)
			return
		}
		td.Ack()
	}
	rabbitmq.ConsumeWithReconnect(b.rmq, ctx,
		"eis.sync.residents.bot", "irgups.sync.residents", "dormitory.eis.dlx",
		handler,
	)
}

func (b *UniversityBroker) applyResidentSync(ctx context.Context, msg *rabbitmq.SyncResidentsPush) (needsRetry bool) {
	if !msg.ValidateVersion() {
		log.Printf("[Broker] SyncResidents version mismatch: %d", msg.Version)
		return false
	}

	// Guard: empty payload → don't set dedup, don't block retries.
	if len(msg.Residents) == 0 {
		log.Printf("[Broker] Resident sync %s: empty payload, skipping", msg.SyncID)
		return false
	}

	// Idempotency — check BEFORE processing, set AFTER success.
	syncKey := syncIDPrefixResidents + msg.SyncID
	if b.redis != nil {
		exists, _ := b.redis.Exists(ctx, syncKey)
		if exists {
			log.Printf("[Broker] Duplicate resident sync %s, skipping", msg.SyncID)
			return false
		}
	}

	timestamp, _ := time.Parse(time.RFC3339, msg.Timestamp)

	dorm, _ := b.dormitoryRepo.FindByField("eis_dormitory_code", msg.DormitoryEISCode)
	if dorm == nil {
		log.Printf("[Broker] Dormitory not found for code %s, will retry in 5s", msg.DormitoryEISCode)
		time.Sleep(5 * time.Second)
		return true
	}

	processed := 0
	eisPhones := make(map[string]bool) // for bot-side departed detection
	for _, r := range msg.Residents {
		normalizedPhone := NormalizePhone(r.Phone)
		if normalizedPhone != "" {
			eisPhones[normalizedPhone] = true
		}

		var user *domain.User
		platformUserID := fmt.Sprintf("eis_%d", r.PersonID)
		eisPersonID := fmt.Sprintf("%d", r.PersonID)

		// 1. Основной поиск: по телефону
		if normalizedPhone != "" {
			u, err := b.userRepo.FindByField("phone", normalizedPhone)
			if err == nil && u != nil {
				user = u
			}
		}

		// 2. Защитный fallback: тот же PersonID мог быть создан
		//    другим сообщением с другим телефоном.
		if user == nil {
			existingByPID, pidErr := b.userRepo.FindByField("platform_user_id", platformUserID)
			if pidErr == nil && existingByPID != nil {
				user = existingByPID
				if user.Phone != normalizedPhone && normalizedPhone != "" {
					user.Phone = normalizedPhone
					b.userRepo.Update(user)
				}
			}
		}

		// 2.5 EISPersonID fallback: user мог быть создан через verify
		//     (где PlatformUserID = MAX ID), а sync ищет по "eis_<PersonID>".
		//     EISPersonID — общий ключ между verify и sync.
		if user == nil && eisPersonID != "" {
			u, eisErr := b.userRepo.FindByField("eis_person_id", eisPersonID)
			if eisErr == nil && u != nil {
				user = u
				// Backfill PlatformUserID so future syncs find this user.
				if user.PlatformUserID != platformUserID {
					user.PlatformUserID = platformUserID
					b.userRepo.Update(user)
				}
				if user.Phone != normalizedPhone && normalizedPhone != "" {
					user.Phone = normalizedPhone
					b.userRepo.Update(user)
				}
			}
		}

		// 3. Не найдено — создаём нового
		if user == nil {
			user = &domain.User{
				FirstName:      r.FirstName,
				LastName:       r.LastName,
				MiddleName:     r.MiddleName,
				Phone:          normalizedPhone,
				Platform:       "max",
				PlatformUserID: platformUserID,
				EISVerified:    true,
				EISPersonID:    fmt.Sprintf("%d", r.PersonID),
				PersonType:     "student",
			}
			if err := b.userRepo.Create(user); err != nil {
				// 4. Обработка гонки: duplicate key при параллельной обработке
				if strings.Contains(err.Error(), "duplicate key") ||
					strings.Contains(err.Error(), "23505") {
					existing, findErr := b.userRepo.FindByField("platform_user_id", platformUserID)
					if findErr == nil && existing != nil {
						user = existing
					} else {
						log.Printf("[Broker] Failed to create user and couldn't recover: %v", err)
						continue
					}
				} else {
					log.Printf("[Broker] Failed to create user from sync: %v", err)
					continue
				}
			}
		} else if user.UpdatedAt.After(timestamp) {
			continue
		}

		room, _ := b.roomRepo.FindByField("eis_room_code", r.RoomEISCode)
		if room == nil {
			continue
		}

		resident, _ := b.residentRepo.FindByField("user_id", user.ID)
		if resident == nil {
			resident = &domain.Resident{
				UserID:      user.ID,
				DormitoryID: dorm.ID,
				RoomID:      room.ID,
				IsActive:    r.IsActive,
			}
			if r.Contract != nil {
				resident.ContractNumber = &r.Contract.Number
				if r.Contract.DateIn != "" {
					v := truncateDate(r.Contract.DateIn)
					resident.ContractStartDate = &v
				}
				if r.Contract.DateEnd != "" {
					v := truncateDate(r.Contract.DateEnd)
					resident.ContractEndDate = &v
				}
			}
			if err := b.residentRepo.Create(resident); err != nil {
				log.Printf("[Broker] Failed to create resident from sync: %v", err)
				continue
			}
		} else if resident.UpdatedAt.Before(timestamp) {
			resident.IsActive = r.IsActive
			resident.RoomID = room.ID
			if r.Contract != nil {
				resident.ContractNumber = &r.Contract.Number
				if r.Contract.DateIn != "" {
					v := truncateDate(r.Contract.DateIn)
					resident.ContractStartDate = &v
				}
				if r.Contract.DateEnd != "" {
					v := truncateDate(r.Contract.DateEnd)
					resident.ContractEndDate = &v
				}
			}
			if err := b.residentRepo.Update(resident); err != nil {
				log.Printf("[Broker] Failed to update resident from sync: %v", err)
				continue
			}
		}
		processed++
	}

	// Deactivate residents present in DB but absent from EIS (bot-side diff).
	// Also honors DeactivatedPhones from adapter (when provided).
	deactivated := make(map[string]bool)
	for _, phone := range msg.DeactivatedPhones {
		deactivated[NormalizePhone(phone)] = true
	}
	var allRes []domain.Resident
	b.residentRepo.DB().Where("dormitory_id = ? AND is_active = true", dorm.ID).Find(&allRes)
	for _, r := range allRes {
		var u domain.User
		if err := b.userRepo.DB().Where("id = ?", r.UserID).First(&u).Error; err == nil {
			phone := NormalizePhone(u.Phone)
			if (deactivated[phone] || !eisPhones[phone]) && r.IsActive && r.UpdatedAt.Before(timestamp) {
				r.IsActive = false
				b.residentRepo.Update(&r)
				processed++
			}
		}
	}

	// Set dedup key ONLY after successful processing.
	if processed > 0 && b.redis != nil {
		b.redis.Set(ctx, syncKey, "1", syncIDTTL)
	}
	log.Printf("[Broker] Applied resident sync %s: %d residents", msg.SyncID, processed)
	return false
}

// ── Sync: Employees ──────────────────────────────────

// ConsumeEmployeeSync consumes employee sync pushes from the adapter.
func (b *UniversityBroker) ConsumeEmployeeSync(ctx context.Context) {
	if b.rmq == nil {
		return
	}
	handler := func(ctx context.Context, td rabbitmq.TypedDelivery[rabbitmq.SyncEmployeesPush]) {
		if b.applyEmployeeSync(ctx, &td.Msg) {
			td.Nack(true)
			return
		}
		td.Ack()
	}
	rabbitmq.ConsumeWithReconnect(b.rmq, ctx,
		"eis.sync.employees.bot", "irgups.sync.employees", "dormitory.eis.dlx",
		handler,
	)
}

func (b *UniversityBroker) applyEmployeeSync(ctx context.Context, msg *rabbitmq.SyncEmployeesPush) (needsRetry bool) {
	if !msg.ValidateVersion() {
		return
	}

	// Guard: empty payload → don't set dedup, don't block retries.
	if len(msg.Employees) == 0 {
		log.Printf("[Broker] Employee sync %s: empty payload, skipping", msg.SyncID)
		return
	}

	// Idempotency — check BEFORE processing, set AFTER success.
	syncKey := syncIDPrefixEmployees + msg.SyncID
	if b.redis != nil {
		exists, _ := b.redis.Exists(ctx, syncKey)
		if exists {
			log.Printf("[Broker] Duplicate employee sync %s, skipping", msg.SyncID)
			return
		}
	}

	timestamp, _ := time.Parse(time.RFC3339, msg.Timestamp)
	processed := 0
	skippedNoPhone := 0
	skippedNewer := 0
	skippedHasEmployee := 0
	createdUsers := 0
	createdEmployees := 0
	assignedDormRoles := 0

	// Resolve dormitory for role assignment
	var dorm *domain.Dormitory
	if msg.DormitoryEISCode != "" {
		dorm, _ = b.dormitoryRepo.FindByField("eis_dormitory_code", msg.DormitoryEISCode)
	}
	// If dorm not found yet (race with dormitory sync), requeue with delay.
	if dorm == nil && msg.DormitoryEISCode != "" {
		log.Printf("[Broker] Dormitory not found for employee sync code %s, will retry in 5s", msg.DormitoryEISCode)
		time.Sleep(5 * time.Second)
		return true
	}

	for _, e := range msg.Employees {
		normalizedPhone := NormalizePhone(e.Phone)
		platformUserID := fmt.Sprintf("eis_emp_%d", e.PersonID)
		eisPersonID := fmt.Sprintf("%d", e.PersonID)

		// 1. Основной поиск: по телефону
		user, err := b.userRepo.FindByField("phone", normalizedPhone)
		if err != nil || user == nil {
			// 2. Защитный fallback: тот же PersonID мог быть создан другим
			//    сообщением синхронизации (другой корпус) с другим телефоном.
			if normalizedPhone != "" {
				existingByPID, pidErr := b.userRepo.FindByField("platform_user_id", platformUserID)
				if pidErr == nil && existingByPID != nil {
					user = existingByPID
					// Обновить телефон на актуальный из этого сообщения
					if user.Phone != normalizedPhone {
						log.Printf("[Broker] Phone changed for %s: %s → %s",
							platformUserID, user.Phone, normalizedPhone)
						user.Phone = normalizedPhone
						b.userRepo.Update(user)
					}
				}
			}

			// 2.5 EISPersonID fallback: user мог быть создан через verify
			//     (PlatformUserID = MAX ID), а sync ищет по "eis_emp_<PersonID>".
			if user == nil && eisPersonID != "" {
				u, eisErr := b.userRepo.FindByField("eis_person_id", eisPersonID)
				if eisErr == nil && u != nil {
					user = u
					if user.PlatformUserID != platformUserID {
						user.PlatformUserID = platformUserID
					}
					if user.Phone != normalizedPhone && normalizedPhone != "" {
						user.Phone = normalizedPhone
					}
					b.userRepo.Update(user)
				}
			}
		}

		if user == nil {
			// 3. Не найдено нигде — создаём нового
			if normalizedPhone == "" {
				skippedNoPhone++
				continue
			}
			user = &domain.User{
				FirstName:      e.FirstName,
				LastName:       e.LastName,
				MiddleName:     e.MiddleName,
				Phone:          normalizedPhone,
				Platform:       "max",
				PlatformUserID: platformUserID,
				EISVerified:    true,
				EISPersonID:    fmt.Sprintf("%d", e.PersonID),
				PersonType:     "employee",
			}
			if err := b.userRepo.Create(user); err != nil {
				// 4. Если Create упал с duplicate key — гонка между
				//    параллельными обработчиками. Найти существующего.
				if strings.Contains(err.Error(), "duplicate key") ||
					strings.Contains(err.Error(), "23505") {
					existing, findErr := b.userRepo.FindByField("platform_user_id", platformUserID)
					if findErr == nil && existing != nil {
						log.Printf("[Broker] Race condition resolved for %s: using existing user", platformUserID)
						user = existing
					} else {
						log.Printf("[Broker] Failed to create user and couldn't recover: %v", err)
						continue
					}
				} else {
					log.Printf("[Broker] Failed to create user from employee sync: %v", err)
					continue
				}
			} else {
				createdUsers++
			}
		} else if user.UpdatedAt.After(timestamp) {
			skippedNewer++
			continue
		}

		emp, _ := b.employeeRepo.FindByField("user_id", user.ID)
		if emp == nil {
			emp = &domain.Employee{
				UserID:     user.ID,
				Position:   e.PositionName,
				Department: e.Department,
				Role:       e.Role,
			}
			if err := b.employeeRepo.Create(emp); err != nil {
				log.Printf("[Broker] Failed to create employee from sync: %v", err)
				continue
			}
			createdEmployees++
		} else {
			skippedHasEmployee++
		}

		// Ensure EmployeeDormitoryRole — required for VerifyUser to find employee's dormitory.
		if dorm != nil {
			created, err := b.employeeDRRepo.CreateOrIgnore(&domain.EmployeeDormitoryRole{
				EmployeeID:  emp.ID,
				DormitoryID: dorm.ID,
				Role:        e.Role,
			})
			if err != nil {
				log.Printf("[Broker] Failed to create dormitory role: %v", err)
			} else if created {
				assignedDormRoles++
			}
		}

		processed++
	}

	// Set dedup key ONLY after successful processing.
	if processed > 0 && b.redis != nil {
		b.redis.Set(ctx, syncKey, "1", syncIDTTL)
	}
	log.Printf("[Broker] Employee sync %s: total=%d, processed=%d, created(users=%d, emp=%d, dormRoles=%d), skipped(noPhone=%d, newer=%d, hasEmp=%d)",
		msg.SyncID, len(msg.Employees), processed, createdUsers, createdEmployees, assignedDormRoles, skippedNoPhone, skippedNewer, skippedHasEmployee)
	return false
}

// ── Sync: Dormitories ────────────────────────────────

// ConsumeDormitorySync consumes dormitory sync pushes from the adapter.
func (b *UniversityBroker) ConsumeDormitorySync(ctx context.Context) {
	if b.rmq == nil {
		return
	}
	handler := func(ctx context.Context, td rabbitmq.TypedDelivery[rabbitmq.SyncDormitoriesPush]) {
		b.applyDormitorySync(ctx, &td.Msg)
		td.Ack()
	}
	rabbitmq.ConsumeWithReconnect(b.rmq, ctx,
		"eis.sync.dormitories.bot", "irgups.sync.dormitories", "dormitory.eis.dlx",
		handler,
	)
}

func (b *UniversityBroker) applyDormitorySync(ctx context.Context, msg *rabbitmq.SyncDormitoriesPush) {
	if !msg.ValidateVersion() {
		return
	}

	// Guard: empty payload → don't set dedup, don't block retries.
	if len(msg.Dormitories) == 0 {
		log.Printf("[Broker] Dormitory sync %s: empty payload, skipping", msg.SyncID)
		return
	}

	// Idempotency — check BEFORE processing, set AFTER success.
	syncKey := syncIDPrefixDormitories + msg.SyncID
	if b.redis != nil {
		exists, _ := b.redis.Exists(ctx, syncKey)
		if exists {
			log.Printf("[Broker] Duplicate dormitory sync %s, skipping", msg.SyncID)
			return
		}
	}

	processed := 0
	for _, d := range msg.Dormitories {
		dorm, _ := b.dormitoryRepo.FindByField("eis_dormitory_code", d.EISCode)
		if dorm == nil {
			normalizedName := NormalizeDormitoryName(d.Name)
			// Also try match by normalized name (for pre-existing dormitories from verify path)
			dorm, _ = b.dormitoryRepo.FindByField("name", normalizedName)
		}
		if dorm == nil {
			normalizedName := NormalizeDormitoryName(d.Name)
			dorm = &domain.Dormitory{
				Name:             normalizedName,
				EISDormitoryCode: d.EISCode,
				IsActive:         true,
			}
			b.dormitoryRepo.Create(dorm)
			log.Printf("[Broker] Created dormitory from sync: %s (from %s)", normalizedName, d.Name)
		} else if dorm.EISDormitoryCode == "" {
			// Backfill code for dormitories created via verify (before sync ran).
			dorm.EISDormitoryCode = d.EISCode
			b.dormitoryRepo.Update(dorm)
			log.Printf("[Broker] Backfilled EISDormitoryCode for %s: %s", dorm.Name, d.EISCode)
		}

		// Ensure floors
		for _, floorNum := range d.Floors {
			b.ensureFloor(ctx, dorm.ID, floorNum)
		}

		// Ensure rooms
		for _, r := range d.Rooms {
			existing, _ := b.roomRepo.FindByField("eis_room_code", r.EISCode)
			if existing == nil {
				fid := b.ensureFloor(ctx, dorm.ID, r.Floor)
				room := &domain.Room{
					DormitoryID: dorm.ID,
					FloorID:     fid,
					RoomNumber:  r.Name,
					Capacity:    int(r.Capacity),
					EISRoomCode: r.EISCode,
					IsActive:    true,
				}
				b.roomRepo.Create(room)
			}
			processed++
		}
	}

	// Set dedup key ONLY after successful processing.
	if processed > 0 && b.redis != nil {
		b.redis.Set(ctx, syncKey, "1", syncIDTTL)
	}
	log.Printf("[Broker] Applied dormitory sync %s: %d dormitories", msg.SyncID, len(msg.Dormitories))
}

func (b *UniversityBroker) ensureFloor(ctx context.Context, dormitoryID uuid.UUID, floorNumber int) uuid.UUID {
	floors, _ := b.floorRepo.FindAllByField("dormitory_id", dormitoryID)
	for _, f := range floors {
		if f.FloorNumber == floorNumber {
			return f.ID
		}
	}
	newFloor := &domain.Floor{
		DormitoryID: dormitoryID,
		FloorNumber: floorNumber,
	}
	b.floorRepo.Create(newFloor)
	return newFloor.ID
}

// ── Helpers shared with db_service.go ─────────────────
// These are duplicates/copies of functions in db_service.go to avoid circular imports.
// When the old EIS code is removed, these become the canonical versions.

func resolveEmployeeRole(position string, department string) string {
	pos := strings.ToLower(position)
	dep := strings.ToLower(department)

	if strings.Contains(pos, "заведующ") {
		return "director"
	}
	if strings.Contains(pos, "комендант") || strings.Contains(pos, "директор") {
		return "director"
	}
	if strings.Contains(pos, "воспитател") || strings.Contains(dep, "воспитател") {
		return "ovr"
	}
	if strings.Contains(pos, "паспортист") || strings.Contains(pos, "специалист") {
		return "manager"
	}
	if strings.Contains(pos, "дежурн") || strings.Contains(pos, "вахт") {
		return "duty"
	}
	return "manager"
}

// NormalizePhone standardizes a Russian phone number to E.164 format.
func NormalizePhone(phone string) string {
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(phone, "")
	if strings.HasPrefix(digits, "8") && len(digits) == 11 {
		return "7" + digits[1:]
	}
	if len(digits) == 10 {
		return "7" + digits
	}
	return digits
}

// NormalizeDormitoryName extracts the canonical name "Общежитие №N" from
// various representations like "АОУ Студгородок общежитие 8", "Общежитие №8", etc.
// Shared between verify (ensureDormitoryFromName) and sync push (applyDormitorySync).
func NormalizeDormitoryName(name string) string {
	lower := strings.ToLower(name)
	re := regexp.MustCompile(`общежит[иеё]+[\s]*(?:n|№|n)?[\s#]*(\d+)`)
	m := re.FindStringSubmatch(lower)
	if len(m) >= 2 {
		return "Общежитие №" + m[1]
	}
	return name
}


