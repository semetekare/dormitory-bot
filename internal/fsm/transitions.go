package fsm

var validTransitions = map[string]map[string]bool{
	StateUninitialized: {
		StateAwaitingContact: true,
	},
	StateAwaitingContact: {
		StateAwaitingEISVerification: true,
		StateAuthorized:              true,
	},
	StateAwaitingEISVerification: {
		StateVerified:        true,
		StateAwaitingContact: true,
	},
	StateVerified: {
		StateInitializingResident: true,
		StateInitializingEmployee: true,
	},
	StateInitializingResident: {
		StateResidentConfirmData: true,
		StateAuthorized:          true,
	},
	StateResidentConfirmData: {
		StateAuthorized:           true,
		StateInitializingResident: true,
	},
	StateInitializingEmployee: {
		StateAuthorizedEmployee: true,
	},
	StateAuthorized: {
		StateLaundry:               true,
		StateRoom:                  true,
		StateInfo:                  true,
		StateChatLinks:             true,
		StateCleaning:              true,
		StateUninitialized:         true,
		StateSelectingResidentRole: true,
	},
	StateAuthorizedEmployee: {
		StateSelectingDormitory: true,
		StateAuthorized:         true,
	},
	StateSelectingDormitory: {
		StateRoleVerified:       true,
		StateAuthorizedEmployee: true,
	},
	StateRoleVerified: {
		StateModulesManagement:     true,
		StateLaundryManagement:     true,
		StateRoomManagement:        true,
		StateCleaningManagement:    true,
		StateReferenceManagement:   true,
		StateChatLinkManagement:    true,
		StateStaffManagement:       true,
		StateRoleManagement:        true,
		StateAwaitingReferenceData: true,
		StateAwaitingChatLinkData:  true,
		StateAuthorizedEmployee:    true,
	},
	StateStaffManagement: {
		StateStaffView:    true,
		StateStaffAdd:     true,
		StateRoleVerified: true,
	},
	StateStaffView: {
		StateStaffRoleAssign: true,
		StateStaffManagement: true,
		StateRoleVerified:    true,
	},
	StateStaffRoleAssign: {
		StateStaffView:    true,
		StateRoleVerified: true,
	},
	StateStaffAdd: {
		StateAwaitingStaffData: true,
		StateStaffManagement:   true,
		StateRoleVerified:      true,
	},
	StateAwaitingStaffData: {
		StateStaffAdd:        true,
		StateStaffManagement: true,
		StateRoleVerified:    true,
	},
	StateRoleManagement: {
		StateRoleView:     true,
		StateRoleCreate:   true,
		StateRoleEdit:     true,
		StateRoleVerified: true,
	},
	StateRoleView: {
		StateRoleEdit:       true,
		StateRoleManagement: true,
		StateRoleVerified:   true,
	},
	StateRoleCreate: {
		StateAwaitingRoleData: true,
		StateRoleManagement:   true,
		StateRoleVerified:     true,
	},
	StateRoleEdit: {
		StateAwaitingRoleData: true,
		StateRoleView:         true,
		StateRoleManagement:   true,
		StateRoleVerified:     true,
	},
	StateAwaitingRoleData: {
		StateRoleCreate:     true,
		StateRoleEdit:       true,
		StateRoleManagement: true,
		StateRoleVerified:   true,
	},
	StateSelectingResidentRole: {
		StateResidentRoleVerified: true,
		StateAuthorized:           true,
	},
	StateResidentRoleVerified: {
		StateLaundryManagement:     true,
		StateCleaningManagement:    true,
		StateChatLinkManagement:    true,
		StateSelectingResidentRole: true,
		StateAuthorized:            true,
	},
	StateResidentProfile: {
		StateRoomManagement: true,
		StateRoleVerified:   true,
	},
	StateAwaitingReferenceData: {
		StateRoleVerified:       true,
		StateAuthorizedEmployee: true,
	},
	StateAwaitingChatLinkData: {
		StateRoleVerified:       true,
		StateAuthorizedEmployee: true,
	},
	StateAwaitingControlData: {
		StateRoleVerified:       true,
		StateAuthorizedEmployee: true,
	},
	StateLaundry: {
		StateLaundryViewQueue:     true,
		StateLaundryManageBooking: true,
		StateModuleRefs:           true,
		StateAuthorized:           true,
	},
	StateLaundryViewQueue: {
		StateLaundry:    true,
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateLaundryManageBooking: {
		StateLaundry:    true,
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateRoom: {
		StateRoomViewItems: true,
		StateModuleRefs:    true,
		StateAuthorized:    true,
	},
	StateRoomViewItems: {
		StateRoom:       true,
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateInfo: {
		StateInfoViewList: true,
		StateAuthorized:   true,
	},
	StateInfoViewList: {
		StateInfoViewItem: true,
		StateInfo:         true,
		StateAuthorized:   true,
	},
	StateInfoViewItem: {
		StateInfoViewList: true,
		StateAuthorized:   true,
	},
	StateChatLinks: {
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateCleaning: {
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateModuleRefs: {
		StateModuleRefView: true,
		StateAuthorized:    true,
	},
	StateModuleRefView: {
		StateModuleRefs: true,
		StateAuthorized: true,
	},
	StateModulesManagement: {
		StateRoleVerified: true,
	},
	StateLaundryManagement: {
		StateAwaitingMachineData:     true,
		StateAwaitingLaundrySettings: true,
		StateAwaitingAdminBooking:    true,
		StateRoleVerified:            true,
	},
	StateAwaitingMachineData: {
		StateLaundryManagement: true,
		StateRoleVerified:      true,
	},
	StateAwaitingLaundrySettings: {
		StateLaundryManagement: true,
		StateRoleVerified:      true,
	},
	StateAwaitingAdminBooking: {
		StateLaundryManagement: true,
		StateRoleVerified:      true,
	},
	StateRoomManagement: {
		StateRoomSearch:          true,
		StateAwaitingControlData: true,
		StateRoleVerified:        true,
		StateResidentProfile:     true,
	},
	StateRoomSearch: {
		StateRoomManagement: true,
		StateRoleVerified:   true,
	},
	StateCleaningManagement: {
		StateAwaitingPenaltyData:   true,
		StateAwaitingExemptionData: true,
		StateRoleVerified:          true,
	},
	StateAwaitingPenaltyData: {
		StateCleaningManagement: true,
		StateRoleVerified:       true,
	},
	StateAwaitingExemptionData: {
		StateCleaningManagement: true,
		StateRoleVerified:       true,
	},
}

func IsValidTransition(fromState, toState string) bool {
	if allowed, ok := validTransitions[fromState]; ok {
		return allowed[toState]
	}
	return false
}
