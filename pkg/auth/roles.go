package auth

import "strings"

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleFinance    = "finance"
	RoleOps        = "ops"
	RoleUser       = "user"
	RoleCustomer   = "customer"
	RoleDriver     = "driver"
	RoleMerchant   = "merchant"
)

func NormalizeRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "super admin", "super-admin":
		return RoleSuperAdmin
	case "ops.admin", "ops_admin":
		return RoleAdmin
	case RoleCustomer:
		return RoleUser
	default:
		return strings.TrimSpace(strings.ToLower(role))
	}
}

func RoleAllowed(actual string, required ...string) bool {
	actual = NormalizeRole(actual)
	if actual == "" {
		return false
	}

	for _, candidate := range required {
		normalized := NormalizeRole(candidate)
		if actual == normalized {
			return true
		}
		if actual == RoleSuperAdmin {
			return true
		}
		if actual == RoleAdmin && (normalized == RoleAdmin || normalized == RoleOps || normalized == RoleFinance || normalized == RoleUser) {
			return true
		}
		if actual == RoleFinance && normalized == RoleUser {
			return true
		}
	}
	return false
}
