package domain

type PermissionContext string

const (
	ContextOwn        PermissionContext = "own"
	ContextAll        PermissionContext = "all"
	ContextDepartment PermissionContext = "department"
)

func ParsePermissionValue(raw interface{}) ([]PermissionContext, bool) {
	switch v := raw.(type) {
	case bool:
		if v {
			return []PermissionContext{ContextAll}, true
		}
		return nil, false
	case string:
		if v == "*" {
			return []PermissionContext{ContextOwn, ContextAll, ContextDepartment}, true
		}
		return []PermissionContext{PermissionContext(v)}, true
	case []interface{}:
		contexts := make([]PermissionContext, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				contexts = append(contexts, PermissionContext(str))
			}
		}
		return contexts, len(contexts) > 0
	default:
		return nil, false
	}
}

func HasContext(contexts []PermissionContext, target PermissionContext) bool {
	for _, ctx := range contexts {
		if ctx == target {
			return true
		}
	}
	return false
}
