package database

import (
	"fmt"
	"log"

	"github.com/dormitory-bot/internal/domain"
	"gorm.io/gorm"
)

var Models = []interface{}{
	&domain.Role{},
	&domain.User{},
	&domain.Employee{},
	&domain.EmployeeDormitoryRole{},
	&domain.Dormitory{},
	&domain.Floor{},
	&domain.Wing{},
	&domain.Room{},
	&domain.Resident{},
	&domain.WashingMachine{},
	&domain.LaundryBooking{},
	&domain.LaundrySlot{},
	&domain.LaundrySettings{},
	&domain.MaterialResponsibility{},
	&domain.SanitaryControl{},
	&domain.DisciplineControl{},
	&domain.CohabitationControl{},
	&domain.ResidentFinancialDebt{},
	&domain.ReferenceMaterial{},
	&domain.ChatLink{},
	&domain.CleaningDuty{},
	&domain.PenaltyCleaning{},
	&domain.CleaningExemption{},
	&domain.CleaningSettings{},
	&domain.BotModule{},
	&domain.UserDormitoryRole{},
}

func init() {
	domain.Models = Models
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	log.Println("Step 1: Creating types...")
	if err := createTypes(db); err != nil {
		log.Printf("Warning: Failed to create types: %v", err)
	}

	log.Println("Step 2: Creating/updating all tables...")
	if err := db.AutoMigrate(Models...); err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Step 2.5: Post-migration fixes...")
	if err := preMigrateFixes(db); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Println("Step 3: Dropping old FK constraints...")
	if err := dropOldConstraints(db); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Println("Step 4: Creating foreign key constraints...")
	if err := createConstraints(db); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Println("Step 5: Creating indexes...")
	if err := createIndexes(db); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Println("Migration completed successfully")
	return nil
}

func createTypes(db *gorm.DB) error {
	types := []string{
		"DO $$ BEGIN CREATE TYPE platform_type AS ENUM ('telegram', 'max', 'vk'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE person_type AS ENUM ('student', 'employee'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE employee_role AS ENUM ('director', 'commandant', 'duty_officer', 'chairman', 'starosta'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE booking_status AS ENUM ('active', 'cancelled', 'completed'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE cancel_source AS ENUM ('resident', 'system', 'employee'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE control_status AS ENUM ('active', 'completed', 'cancelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE chat_link_type AS ENUM ('dormitory', 'floor', 'wing', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE chat_platform AS ENUM ('telegram', 'whatsapp', 'viber', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE cleaning_type AS ENUM ('regular', 'penalty', 'kitchen', 'garbage'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE duty_status AS ENUM ('scheduled', 'completed', 'missed', 'cancelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		"DO $$ BEGIN CREATE TYPE bot_status AS ENUM ('running', 'stopped', 'error'); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
	}
	for _, t := range types {
		if err := db.Exec(t).Error; err != nil {
			log.Printf("Warning creating type: %v", err)
		}
	}
	return nil
}

func dropOldConstraints(db *gorm.DB) error {
	return nil
}

type constraintInfo struct {
	table      string
	constraint string
	sql        string
}

func createConstraints(db *gorm.DB) error {
	constraints := []constraintInfo{
		{`"resident"`, "fk_resident_user", `ALTER TABLE "resident" ADD CONSTRAINT "fk_resident_user" FOREIGN KEY ("user_id") REFERENCES "user"("id")`},
		{`"resident"`, "fk_resident_dormitory", `ALTER TABLE "resident" ADD CONSTRAINT "fk_resident_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"resident"`, "fk_resident_room", `ALTER TABLE "resident" ADD CONSTRAINT "fk_resident_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"employee"`, "fk_employee_user", `ALTER TABLE "employee" ADD CONSTRAINT "fk_employee_user" FOREIGN KEY ("user_id") REFERENCES "user"("id")`},
		{`"employee"`, "fk_employee_primary_dormitory", `ALTER TABLE "employee" ADD CONSTRAINT "fk_employee_primary_dormitory" FOREIGN KEY ("primary_dormitory_id") REFERENCES "dormitory"("id")`},
		{`"employee_dormitory_role"`, "fk_employee_dormitory_role_employee", `ALTER TABLE "employee_dormitory_role" ADD CONSTRAINT "fk_employee_dormitory_role_employee" FOREIGN KEY ("employee_id") REFERENCES "employee"("id")`},
		{`"employee_dormitory_role"`, "fk_employee_dormitory_role_dormitory", `ALTER TABLE "employee_dormitory_role" ADD CONSTRAINT "fk_employee_dormitory_role_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"floor"`, "fk_floor_dormitory", `ALTER TABLE "floor" ADD CONSTRAINT "fk_floor_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"wing"`, "fk_wing_floor", `ALTER TABLE "wing" ADD CONSTRAINT "fk_wing_floor" FOREIGN KEY ("floor_id") REFERENCES "floor"("id")`},
		{`"room"`, "fk_room_dormitory", `ALTER TABLE "room" ADD CONSTRAINT "fk_room_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"room"`, "fk_room_floor", `ALTER TABLE "room" ADD CONSTRAINT "fk_room_floor" FOREIGN KEY ("floor_id") REFERENCES "floor"("id")`},
		{`"room"`, "fk_room_wing", `ALTER TABLE "room" ADD CONSTRAINT "fk_room_wing" FOREIGN KEY ("wing_id") REFERENCES "wing"("id")`},
		{`"washing_machine"`, "fk_washing_machine_dormitory", `ALTER TABLE "washing_machine" ADD CONSTRAINT "fk_washing_machine_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"washing_machine"`, "fk_washing_machine_floor", `ALTER TABLE "washing_machine" ADD CONSTRAINT "fk_washing_machine_floor" FOREIGN KEY ("floor_id") REFERENCES "floor"("id")`},
		{`"laundry_booking"`, "fk_laundry_booking_machine", `ALTER TABLE "laundry_booking" ADD CONSTRAINT "fk_laundry_booking_machine" FOREIGN KEY ("machine_id") REFERENCES "washing_machine"("id")`},
		{`"laundry_booking"`, "fk_laundry_booking_user", `ALTER TABLE "laundry_booking" ADD CONSTRAINT "fk_laundry_booking_user" FOREIGN KEY ("user_id") REFERENCES "user"("id")`},
		{`"laundry_settings"`, "fk_laundry_settings_dormitory", `ALTER TABLE "laundry_settings" ADD CONSTRAINT "fk_laundry_settings_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"material_responsibility"`, "fk_material_responsibility_room", `ALTER TABLE "material_responsibility" ADD CONSTRAINT "fk_material_responsibility_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"material_responsibility"`, "fk_material_responsibility_accepted_by_resident", `ALTER TABLE "material_responsibility" ADD CONSTRAINT "fk_material_responsibility_accepted_by_resident" FOREIGN KEY ("accepted_by_resident_id") REFERENCES "resident"("id")`},
		{`"sanitary_control"`, "fk_sanitary_control_room", `ALTER TABLE "sanitary_control" ADD CONSTRAINT "fk_sanitary_control_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"sanitary_control"`, "fk_sanitary_control_inspector", `ALTER TABLE "sanitary_control" ADD CONSTRAINT "fk_sanitary_control_inspector" FOREIGN KEY ("inspector_id") REFERENCES "employee"("id")`},
		{`"discipline_control"`, "fk_discipline_control_room", `ALTER TABLE "discipline_control" ADD CONSTRAINT "fk_discipline_control_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"discipline_control"`, "fk_discipline_control_resident", `ALTER TABLE "discipline_control" ADD CONSTRAINT "fk_discipline_control_resident" FOREIGN KEY ("resident_id") REFERENCES "resident"("id")`},
		{`"discipline_control"`, "fk_discipline_control_responsible", `ALTER TABLE "discipline_control" ADD CONSTRAINT "fk_discipline_control_responsible" FOREIGN KEY ("responsible_id") REFERENCES "employee"("id")`},
		{`"cohabitation_control"`, "fk_cohabitation_control_room", `ALTER TABLE "cohabitation_control" ADD CONSTRAINT "fk_cohabitation_control_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"resident_financial_debt"`, "fk_resident_financial_debt_resident", `ALTER TABLE "resident_financial_debt" ADD CONSTRAINT "fk_resident_financial_debt_resident" FOREIGN KEY ("resident_id") REFERENCES "resident"("id")`},
		{`"reference_material"`, "fk_reference_material_dormitory", `ALTER TABLE "reference_material" ADD CONSTRAINT "fk_reference_material_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"reference_material"`, "fk_reference_material_created_by", `ALTER TABLE "reference_material" ADD CONSTRAINT "fk_reference_material_created_by" FOREIGN KEY ("created_by") REFERENCES "employee"("id")`},
		{`"chat_link"`, "fk_chat_link_dormitory", `ALTER TABLE "chat_link" ADD CONSTRAINT "fk_chat_link_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"chat_link"`, "fk_chat_link_floor", `ALTER TABLE "chat_link" ADD CONSTRAINT "fk_chat_link_floor" FOREIGN KEY ("floor_id") REFERENCES "floor"("id")`},
		{`"chat_link"`, "fk_chat_link_wing", `ALTER TABLE "chat_link" ADD CONSTRAINT "fk_chat_link_wing" FOREIGN KEY ("wing_id") REFERENCES "wing"("id")`},
		{`"chat_link"`, "fk_chat_link_created_by", `ALTER TABLE "chat_link" ADD CONSTRAINT "fk_chat_link_created_by" FOREIGN KEY ("created_by") REFERENCES "employee"("id")`},
		{`"cleaning_duty"`, "fk_cleaning_duty_dormitory", `ALTER TABLE "cleaning_duty" ADD CONSTRAINT "fk_cleaning_duty_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"cleaning_duty"`, "fk_cleaning_duty_room", `ALTER TABLE "cleaning_duty" ADD CONSTRAINT "fk_cleaning_duty_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"cleaning_duty"`, "fk_cleaning_duty_floor", `ALTER TABLE "cleaning_duty" ADD CONSTRAINT "fk_cleaning_duty_floor" FOREIGN KEY ("floor_id") REFERENCES "floor"("id")`},
		{`"cleaning_duty"`, "fk_cleaning_duty_completed_by", `ALTER TABLE "cleaning_duty" ADD CONSTRAINT "fk_cleaning_duty_completed_by" FOREIGN KEY ("completed_by_id") REFERENCES "resident"("id")`},
		{`"cleaning_duty"`, "fk_cleaning_duty_penalty", `ALTER TABLE "cleaning_duty" ADD CONSTRAINT "fk_cleaning_duty_penalty" FOREIGN KEY ("penalty_cleaning_id") REFERENCES "penalty_cleaning"("id")`},
		{`"penalty_cleaning"`, "fk_penalty_cleaning_room", `ALTER TABLE "penalty_cleaning" ADD CONSTRAINT "fk_penalty_cleaning_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"penalty_cleaning"`, "fk_penalty_cleaning_resident", `ALTER TABLE "penalty_cleaning" ADD CONSTRAINT "fk_penalty_cleaning_resident" FOREIGN KEY ("resident_id") REFERENCES "resident"("id")`},
		{`"penalty_cleaning"`, "fk_penalty_cleaning_issued_by", `ALTER TABLE "penalty_cleaning" ADD CONSTRAINT "fk_penalty_cleaning_issued_by" FOREIGN KEY ("issued_by_id") REFERENCES "employee"("id")`},
		{`"cleaning_exemption"`, "fk_cleaning_exemption_room", `ALTER TABLE "cleaning_exemption" ADD CONSTRAINT "fk_cleaning_exemption_room" FOREIGN KEY ("room_id") REFERENCES "room"("id")`},
		{`"cleaning_exemption"`, "fk_cleaning_exemption_created_by", `ALTER TABLE "cleaning_exemption" ADD CONSTRAINT "fk_cleaning_exemption_created_by" FOREIGN KEY ("created_by_id") REFERENCES "employee"("id")`},
		{`"bot_module"`, "fk_bot_module_dormitory", `ALTER TABLE "bot_module" ADD CONSTRAINT "fk_bot_module_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"user_dormitory_role"`, "fk_udr_user", `ALTER TABLE "user_dormitory_role" ADD CONSTRAINT "fk_udr_user" FOREIGN KEY ("user_id") REFERENCES "user"("id")`},
		{`"user_dormitory_role"`, "fk_udr_dormitory", `ALTER TABLE "user_dormitory_role" ADD CONSTRAINT "fk_udr_dormitory" FOREIGN KEY ("dormitory_id") REFERENCES "dormitory"("id")`},
		{`"user_dormitory_role"`, "fk_udr_assigned_by", `ALTER TABLE "user_dormitory_role" ADD CONSTRAINT "fk_udr_assigned_by" FOREIGN KEY ("assigned_by") REFERENCES "employee"("id")`},
	}

	for _, c := range constraints {
		sql := fmt.Sprintf("DO $$ BEGIN %s; EXCEPTION WHEN duplicate_object THEN NULL; END $$", c.sql)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Warning: Failed to create FK constraint %s on %s: %v", c.constraint, c.table, err)
		}
	}
	return nil
}

func createIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_user_phone ON \"user\"(phone)",
		"CREATE INDEX IF NOT EXISTS idx_user_platform ON \"user\"(platform, platform_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_eis_person_id ON \"user\"(eis_person_id)",
		"CREATE INDEX IF NOT EXISTS idx_resident_user_id ON resident(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_resident_dormitory_id ON resident(dormitory_id)",
		"CREATE INDEX IF NOT EXISTS idx_resident_room_id ON resident(room_id)",
		"CREATE INDEX IF NOT EXISTS idx_employee_user_id ON employee(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_employee_role ON employee(role)",
		"CREATE INDEX IF NOT EXISTS idx_edr_employee_dormitory ON employee_dormitory_role(employee_id, dormitory_id)",
		"CREATE INDEX IF NOT EXISTS idx_dormitory_eis_code ON dormitory(eis_dormitory_code)",
		"CREATE INDEX IF NOT EXISTS idx_floor_dormitory_number ON floor(dormitory_id, floor_number)",
		"CREATE INDEX IF NOT EXISTS idx_wing_floor_name ON wing(floor_id, name)",
		"CREATE INDEX IF NOT EXISTS idx_room_dorm_floor_number ON room(dormitory_id, floor_id, room_number)",
		"CREATE INDEX IF NOT EXISTS idx_wm_dorm_floor_number ON washing_machine(dormitory_id, floor_id, machine_number)",
		"CREATE INDEX IF NOT EXISTS idx_lb_machine_slot ON laundry_booking(machine_id, slot_date, slot_start)",
		"CREATE INDEX IF NOT EXISTS idx_lb_user_date ON laundry_booking(user_id, slot_date)",
		"CREATE INDEX IF NOT EXISTS idx_cleaning_duty_dormitory_date ON cleaning_duty(dormitory_id, duty_date)",
		"CREATE INDEX IF NOT EXISTS idx_cleaning_duty_floor_room_date ON cleaning_duty(floor_id, room_id, duty_date)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("Warning creating index: %v", err)
		}
	}

	db.Exec(`UPDATE reference_material SET module_key = 'references' WHERE module_key IS NULL`)

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_udr_user_dormitory ON user_dormitory_role(user_id, dormitory_id)").Error; err != nil {
		log.Printf("Warning creating index idx_udr_user_dormitory: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_udr_dormitory ON user_dormitory_role(dormitory_id)").Error; err != nil {
		log.Printf("Warning creating index idx_udr_dormitory: %v", err)
	}
	return nil
}

func preMigrateFixes(db *gorm.DB) error {
	db.Exec("UPDATE cleaning_duty SET floor_id = (SELECT floor_id FROM room WHERE room.id = cleaning_duty.room_id) WHERE floor_id IS NULL")

	var cnt int64
	db.Raw("SELECT COUNT(*) FROM cleaning_duty WHERE floor_id IS NULL").Scan(&cnt)
	if cnt > 0 {
		log.Printf("Cleaning duties without floor_id: %d, deleting them", cnt)
		db.Exec("DELETE FROM cleaning_duty WHERE floor_id IS NULL")
	}

	db.Exec(`ALTER TABLE "cleaning_duty" ADD COLUMN IF NOT EXISTS "penalty_cleaning_id" uuid`)

	// #05 Migration: rename resident_id → user_id + add booker_type on laundry_booking.
	// Run only if resident_id still exists.
	var colExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name='laundry_booking' AND column_name='resident_id'
	)`).Scan(&colExists)
	if colExists {
		// Drop old FK before renaming column.
		db.Exec(`ALTER TABLE "laundry_booking" DROP CONSTRAINT IF EXISTS "fk_laundry_booking_resident"`)
		db.Exec(`ALTER TABLE "laundry_booking" DROP CONSTRAINT IF EXISTS "laundry_booking_resident_id_fkey"`)
		db.Exec(`ALTER TABLE "laundry_booking" RENAME COLUMN "resident_id" TO "user_id"`)
	}
	db.Exec(`ALTER TABLE "laundry_booking" ADD COLUMN IF NOT EXISTS "booker_type" varchar(10) NOT NULL DEFAULT 'resident'`)

	// #06 Migration: rename booking_same_day_only → advance_booking_enabled + add advance_booking_max_days.
	db.Exec(`ALTER TABLE "laundry_settings" ADD COLUMN IF NOT EXISTS "advance_booking_enabled" boolean NOT NULL DEFAULT false`)
	db.Exec(`ALTER TABLE "laundry_settings" ADD COLUMN IF NOT EXISTS "advance_booking_max_days" integer NOT NULL DEFAULT 1`)
	// Drop the old column if it still exists.
	var bsdColExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name='laundry_settings' AND column_name='booking_same_day_only'
	)`).Scan(&bsdColExists)
	if bsdColExists {
		// Migrate: if booking_same_day_only was false, enable advance booking.
		db.Exec(`UPDATE "laundry_settings" SET "advance_booking_enabled" = true WHERE "booking_same_day_only" = false`)
		db.Exec(`ALTER TABLE "laundry_settings" DROP COLUMN "booking_same_day_only"`)
	}

	return nil
}
