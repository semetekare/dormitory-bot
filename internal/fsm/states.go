package fsm

const (
	StateUninitialized           = "UNINITIALIZED"
	StateAwaitingContact         = "AWAITING_CONTACT"
	StateAwaitingEISVerification = "AWAITING_EIS_VERIFICATION"
	StateVerified                = "VERIFIED"
	StateInitializingResident    = "INITIALIZING_RESIDENT"
	StateResidentConfirmData     = "RESIDENT_CONFIRM_DATA"
	StateInitializingEmployee    = "INITIALIZING_EMPLOYEE"
	StateAuthorized              = "AUTHORIZED"
	StateAuthorizedEmployee      = "AUTHORIZED_EMPLOYEE"
	StateSelectingDormitory      = "SELECTING_DORMITORY"
	StateRoleVerified            = "ROLE_VERIFIED"

	StateLaundry              = "LAUNDRY"
	StateLaundryViewQueue     = "LAUNDRY_VIEW_QUEUE"
	StateLaundryManageBooking = "LAUNDRY_MANAGE_BOOKING"

	StateRoom          = "ROOM"
	StateRoomViewItems = "ROOM_VIEW_ITEMS"

	StateInfo         = "INFO"
	StateInfoViewList = "INFO_VIEW_LIST"
	StateInfoViewItem = "INFO_VIEW_ITEM"

	StateChatLinks = "CHAT_LINKS"

	StateCleaning = "CLEANING"

	StateModulesManagement = "MODULES_MANAGEMENT"

	StateLaundryManagement = "LAUNDRY_MANAGEMENT"

	StateRoomManagement = "ROOM_MANAGEMENT"
	StateRoomSearch     = "ROOM_SEARCH"

	StateCleaningManagement = "CLEANING_MANAGEMENT"

	StateReferenceManagement     = "REFERENCE_MANAGEMENT"
	StateChatLinkManagement      = "CHAT_LINK_MANAGEMENT"
	StateAwaitingReferenceData   = "AWAITING_REFERENCE_DATA"
	StateAwaitingChatLinkData    = "AWAITING_CHAT_LINK_DATA"
	StateAwaitingControlData     = "AWAITING_CONTROL_DATA"
	StateAwaitingMachineData     = "AWAITING_MACHINE_DATA"
	StateAwaitingLaundrySettings = "AWAITING_LAUNDRY_SETTINGS"
	StateAwaitingAdminBooking    = "AWAITING_ADMIN_BOOKING"

	StateAwaitingPenaltyData   = "AWAITING_PENALTY_DATA"
	StateAwaitingExemptionData = "AWAITING_EXEMPTION_DATA"

	StateModuleRefs    = "MODULE_REFS"
	StateModuleRefView = "MODULE_REF_VIEW"

	StateStaffManagement   = "STAFF_MANAGEMENT"
	StateStaffView         = "STAFF_VIEW"
	StateStaffRoleAssign   = "STAFF_ROLE_ASSIGN"
	StateStaffAdd          = "STAFF_ADD"
	StateAwaitingStaffData = "AWAITING_STAFF_DATA"

	StateRoleManagement   = "ROLE_MANAGEMENT"
	StateRoleView         = "ROLE_VIEW"
	StateRoleCreate       = "ROLE_CREATE"
	StateRoleEdit         = "ROLE_EDIT"
	StateAwaitingRoleData = "AWAITING_ROLE_DATA"

	StateSelectingResidentRole = "SELECTING_RESIDENT_ROLE"
	StateResidentRoleVerified  = "RESIDENT_ROLE_VERIFIED"
	StateResidentProfile       = "RESIDENT_PROFILE"
)
