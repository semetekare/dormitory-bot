package domain

type ResourceModule struct {
	Key         string
	DisplayName string
	TargetType  string
	Operations  []string
}

const (
	TargetTypeEmployee = "employee"
	TargetTypeResident = "resident"
	TargetTypeBoth     = "both"
)

var ResourceRegistryOrder = []string{
	"laundry", "room", "cleaning", "chat_link", "reference", "module", "staff", "role",
}

var ResourceRegistry = map[string]ResourceModule{
	"laundry": {
		Key:         "laundry",
		DisplayName: "Стирка",
		TargetType:  TargetTypeBoth,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"room": {
		Key:         "room",
		DisplayName: "Комнаты и жильцы",
		TargetType:  TargetTypeEmployee,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"cleaning": {
		Key:         "cleaning",
		DisplayName: "Дежурства",
		TargetType:  TargetTypeBoth,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"chat_link": {
		Key:         "chat_link",
		DisplayName: "Ссылки на чаты",
		TargetType:  TargetTypeBoth,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"reference": {
		Key:         "reference",
		DisplayName: "Справки",
		TargetType:  TargetTypeEmployee,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"module": {
		Key:         "module",
		DisplayName: "Боты",
		TargetType:  TargetTypeEmployee,
		Operations:  []string{"read", "update"},
	},
	"staff": {
		Key:         "staff",
		DisplayName: "Сотрудники",
		TargetType:  TargetTypeEmployee,
		Operations:  []string{"read", "create", "update", "delete"},
	},
	"role": {
		Key:         "role",
		DisplayName: "Роли",
		TargetType:  TargetTypeEmployee,
		Operations:  []string{"read", "create", "update", "delete"},
	},
}

func (r ResourceModule) ResourceKey() string {
	return "resource." + r.Key
}

func GetResourceModule(key string) (ResourceModule, bool) {
	m, ok := ResourceRegistry[key]
	return m, ok
}

func GetModuleDisplayNames() map[string]string {
	result := make(map[string]string, len(ResourceRegistry))
	for _, m := range ResourceRegistry {
		result[m.Key] = m.DisplayName
	}
	return result
}

func GetModulesByTargetType(targetType string) []ResourceModule {
	var result []ResourceModule
	for _, m := range ResourceRegistry {
		if m.TargetType == targetType || m.TargetType == TargetTypeBoth {
			result = append(result, m)
		}
	}
	return result
}

func (r ResourceModule) HasOperation(op string) bool {
	for _, o := range r.Operations {
		if o == op {
			return true
		}
	}
	return false
}

func GenerateFullAccessJSON() JSONAccessMap {
	access := make(JSONAccessMap)
	for key := range ResourceRegistry {
		access["resource."+key] = ResourcePermissions{"*": true}
	}
	return access
}

func GenerateReadOnlyJSON() JSONAccessMap {
	access := make(JSONAccessMap)
	for key := range ResourceRegistry {
		access["resource."+key] = ResourcePermissions{"read": true}
	}
	return access
}

func ValidateJSONAccess(access JSONAccessMap) []string {
	var errors []string
	for resourceKey, perms := range access {
		moduleKey := resourceKey
		if len(resourceKey) > 9 && resourceKey[:9] == "resource." {
			moduleKey = resourceKey[9:]
		}
		if _, ok := ResourceRegistry[moduleKey]; !ok {
			errors = append(errors, "неизвестный модуль: "+resourceKey)
			continue
		}
		for op := range perms {
			if op == "*" {
				continue
			}
			if !ResourceRegistry[moduleKey].HasOperation(op) {
				errors = append(errors, "неизвестная операция "+op+" для модуля "+moduleKey)
			}
		}
	}
	return errors
}
