package casbin

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"ac/bootstrap/database"

	casbin "github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
)

var (
	// enforcer holds the singleton Casbin TransactionalEnforcer instance.
	enforcer *casbin.TransactionalEnforcer
	// once ensures the Initialize function runs only a single time.
	once sync.Once
	// initErr stores any error encountered during initialization.
	initErr error
)

const (
	// GTypeUserRole is the grouping policy type for linking users to roles.
	GTypeUserRole = "g"
	// GTypeObjectGroup is the grouping policy type for linking objects to object groups.
	GTypeObjectGroup = "g2"

	// Type codes for internal subject/object identification.
	UserCode = "user"
	RoleCode = "role"
	MenuCode = "menu" // Used as an object type in policies

	// Prefixes are used to differentiate subject/object types in the Casbin policy store.
	// This design choice is necessary because Casbin treats all entities in a column
	// uniformly (e.g., users and roles are both subjects 'sub'), but our application
	// logic requires distinguishing them clearly.
	PrefixUser = "u:"
	PrefixRole = "r:"
	PrefixMenu = "o:"
)

var (
	// Standardized errors for clarity and easy checking by callers.
	ErrEnforcerNotInit     = fmt.Errorf("casbin enforcer not initialized")
	ErrPolicyExists        = fmt.Errorf("policy already exists")
	ErrPolicyNotExist      = fmt.Errorf("policy does not exist")
	ErrGroupingExists      = fmt.Errorf("grouping policy already exists")
	ErrGroupingNotExist    = fmt.Errorf("grouping policy does not exist")
	ErrRoleAlreadyAssigned = fmt.Errorf("role already assigned to user")
	ErrRoleNotAssigned     = fmt.Errorf("role not assigned to user")
	ErrUserAlreadyAssigned = fmt.Errorf("user already assigned to role")
	ErrUserNotAssigned     = fmt.Errorf("user not assigned to role")
)

// Initialize sets up the Casbin enforcer as a singleton.
// It loads the model from a string and configures the Gorm adapter.
// The model includes time-based enforcement logic.
// It uses `sync.Once` to ensure thread-safe, single-time initialization.
func Initialize() error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: started\n")

		// Prevent auto-migration to manage schema evolution outside of this package.
		gormadapter.TurnOffAutoMigrate(database.DB)

		adapter, err := gormadapter.NewTransactionalAdapterByDB(database.DB)
		if err != nil {
			initErr = fmt.Errorf("failed to create Gorm Adapter: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create adapter failed: %v\n", err)
			return
		}

		// Defines the Casbin model.
		// Key design: Includes 'time' in the request and 'begin_time', 'end_time' in the policy
		// to enable time-restricted access control.
		m, err := casbinModel.NewModelFromString(`
[request_definition]
r = sub, obj, act, time

[policy_definition]
p = sub, obj, act, begin_time, end_time

[role_definition]
g = _, _ // User-Role grouping
g2 = _, _ // Object-Group grouping

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g2(r.obj, p.obj) && r.act == p.act && r.time >= p.begin_time && r.time <= p.end_time
`)
		if err != nil {
			initErr = fmt.Errorf("failed to load Casbin model: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: load model failed: %v\n", err)
			return
		}

		// Use TransactionalEnforcer for consistency in policy changes.
		enforcer, err = casbin.NewTransactionalEnforcer(m, adapter)
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin enforcer: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create enforcer failed: %v\n", err)
			return
		}

		// NOTE: In a production environment, consider adding `enforcer.EnableAutoNotify(true)`
		// if other instances might modify the policy store, to keep local enforcers synchronized.

		fmt.Fprintf(os.Stdout, "INFO: casbin: init: succeeded, model=memory, policy_table=casbin_rule\n")
	})
	return initErr
}

// LoadPolicy reloads all policies from the underlying persistence layer into the enforcer's cache.
// This is necessary if policies are modified directly in the database or by other services.
//
// Deprecation Hint: Casbin has built-in mechanisms (like watchers) that might be a better choice
// for real-time synchronization in a distributed environment than manual reloads.
func LoadPolicy(ctx *gin.Context) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return err
	}
	return nil
}

// Enforce checks if a subject (user/role) has permission to perform an action on an object at the given time.
// The time parameter is crucial for the time-based access control model (`r.time >= p.begin_time && r.time <= p.end_time`).
func Enforce(ctx *gin.Context, sub, obj, act string, t time.Time) (bool, error) {
	if enforcer == nil {
		return false, ErrEnforcerNotInit
	}
	// Use UTC time for consistency if the input time 't' is not already defined as local/UTC.
	ok, err := enforcer.Enforce(sub, obj, act, formatTime(t))
	return ok, err
}

// AddPolicy adds a new access control policy with time constraints.
// It checks for existing policy first to prevent duplicates, returning a specific error if found.
// The operation is wrapped in a transaction for atomicity.
func AddPolicy(ctx *gin.Context, sub, obj, act string, beginTime, endTime time.Time) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	policy := []any{sub, obj, act, formatTime(beginTime), formatTime(endTime)}

	exists, err := enforcer.HasPolicy(policy...)
	if err != nil {
		return fmt.Errorf("failed to check existing policy: %w", err)
	}
	if exists {
		return ErrPolicyExists
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		// Use tx.AddPolicy to ensure the operation is part of the transaction.
		_, err := tx.AddPolicy(policy...)
		if err != nil {
			return fmt.Errorf("failed to add policy: %w", err)
		}
		return nil
	})
}

// RemovePolicy deletes an existing access control policy.
// It checks if the policy exists first to provide a clear error state for non-existence.
// The operation is wrapped in a transaction for atomicity.
func RemovePolicy(ctx *gin.Context, sub, obj, act string, beginTime, endTime time.Time) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	policy := []any{sub, obj, act, formatTime(beginTime), formatTime(endTime)}

	exists, err := enforcer.HasPolicy(policy...)
	if err != nil {
		return fmt.Errorf("failed to check existing policy: %w", err)
	}
	if !exists {
		return ErrPolicyNotExist
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		// Use tx.RemovePolicy to ensure the operation is part of the transaction.
		removed, err := tx.RemovePolicy(policy...)
		if err != nil {
			return fmt.Errorf("failed to remove policy: %w", err)
		}
		// A check on 'removed' ensures the expected change occurred.
		if !removed {
			return fmt.Errorf("no policy removed")
		}
		return nil
	})
}

// AddSubjectToGroup establishes a grouping relationship (e.g., User to Role or Object to Group).
// It prevents adding a grouping that already exists.
func AddSubjectToGroup(ctx *gin.Context, gType, subject, group string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	exists, err := enforcer.HasNamedGroupingPolicy(gType, subject, group)
	if err != nil {
		return fmt.Errorf("failed to check grouping policy: %w", err)
	}
	if exists {
		return ErrGroupingExists
	}
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		added, err := tx.AddNamedGroupingPolicy(gType, subject, group)
		if err != nil {
			return fmt.Errorf("failed to add grouping policy: %w", err)
		}
		if !added {
			return fmt.Errorf("no grouping policy added")
		}
		return nil
	})
}

// RemoveSubjectFromGroup deletes a grouping relationship.
// It returns an error if the grouping relationship doesn't exist.
func RemoveSubjectFromGroup(ctx *gin.Context, gType, subject, group string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	exists, err := enforcer.HasNamedGroupingPolicy(gType, subject, group)
	if err != nil {
		return fmt.Errorf("failed to check grouping policy: %w", err)
	}
	if !exists {
		return ErrGroupingNotExist
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		removed, err := tx.RemoveNamedGroupingPolicy(gType, subject, group)
		if err != nil {
			return fmt.Errorf("failed to remove grouping policy: %w", err)
		}
		if !removed {
			// Check against unexpected removal failure.
			return fmt.Errorf("no grouping policy removed")
		}
		return nil
	})
}

// GetGroupsForSubject retrieves all groups a specific subject belongs to for a given grouping type.
// The function ensures only unique groups are returned by using a map intermediate.
func GetGroupsForSubject(ctx *gin.Context, gType, subject string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	// Filter by grouping type (gType) and subject (at index 0).
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(gType, 0, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to query grouping policy: %w", err)
	}

	// Use a map to collect unique group names (which are at index 1).
	groupSet := make(map[string]struct{})
	for _, p := range policies {
		if len(p) >= 2 {
			groupSet[p[1]] = struct{}{}
		}
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}

	return groups, nil
}

// GetSubjectsForGroup retrieves all subjects that belong to a specific group for a given grouping type.
// The function ensures only unique subjects are returned by using a map intermediate.
func GetSubjectsForGroup(ctx *gin.Context, gType, group string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	// Filter by grouping type (gType) and group (at index 1).
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(gType, 1, group)
	if err != nil {
		return nil, fmt.Errorf("failed to query grouping policy: %w", err)
	}

	// Use a map to collect unique subject names (which are at index 0).
	subjectSet := make(map[string]struct{})
	for _, p := range policies {
		if len(p) >= 2 {
			subjectSet[p[0]] = struct{}{}
		}
	}

	subjects := make([]string, 0, len(subjectSet))
	for s := range subjectSet {
		subjects = append(subjects, s)
	}

	return subjects, nil
}

// formatTime converts a time.Time object to a string format compatible with the Casbin model's time check.
// Use time.RFC3339 for correct comparison in the matcher `r.time >= p.begin_time && r.time <= p.end_time`.
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// addPrefix prepends a constant prefix to a code based on its type.
// This is essential for distinguishing between users, roles, and objects
// when stored in the same Casbin `sub` or `obj` policy field.
func addPrefix(code, codeType string) string {
	switch codeType {
	case UserCode:
		return PrefixUser + code
	case RoleCode:
		return PrefixRole + code
	case MenuCode:
		return PrefixMenu + code
	default:
		return code
	}
}

// removePrefix removes the known constant prefix from a prefixed code.
// This restores the original application-specific code for use outside this layer.
func removePrefix(code, codeType string) string {
	switch codeType {
	case UserCode:
		return strings.TrimPrefix(code, PrefixUser)
	case RoleCode:
		return strings.TrimPrefix(code, PrefixRole)
	case MenuCode:
		return strings.TrimPrefix(code, PrefixMenu)
	default:
		return code
	}
}

// AssignRolesToUser assigns a list of roles to a user.
// It enforces the User-Role grouping policy (`GTypeUserRole`).
// It checks if any role is already assigned and aborts the entire transaction on conflict,
// ensuring the assignment is atomic and idempotent for existing policies.
func AssignRolesToUser(ctx *gin.Context, userCode string, roleCodes []string) (int, error) {
	if enforcer == nil {
		return 0, ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// Pre-check for existing roles to fail early and atomically.
	for _, role := range roleCodes {
		rolePrefixed := addPrefix(role, RoleCode)
		// Error ignored here since the only possible error is from Casbin, which is not critical for an existence check.
		exists, _ := enforcer.HasNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
		if exists {
			return 0, fmt.Errorf("%w: role %s to user %s", ErrRoleAlreadyAssigned, role, userCode)
		}
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, role := range roleCodes {
			rolePrefixed := addPrefix(role, RoleCode)
			// Use AddNamedGroupingPolicy for role assignment.
			_, err := tx.AddNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if err != nil {
				return fmt.Errorf("failed to assign role %s to user %s: %w", role, userCode, err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(roleCodes), nil
}

// RemoveRolesFromUser removes a list of roles from a user.
// It checks if all roles are currently assigned and aborts the entire transaction on conflict,
// ensuring the removal is atomic.
func RemoveRolesFromUser(ctx *gin.Context, userCode string, roleCodes []string) (int, error) {
	if enforcer == nil {
		return 0, ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// Pre-check for missing roles to fail early and atomically.
	for _, role := range roleCodes {
		rolePrefixed := addPrefix(role, RoleCode)
		exists, _ := enforcer.HasNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
		if !exists {
			return 0, fmt.Errorf("%w: role %s from user %s", ErrRoleNotAssigned, role, userCode)
		}
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, role := range roleCodes {
			rolePrefixed := addPrefix(role, RoleCode)
			// Use RemoveNamedGroupingPolicy for role de-assignment.
			_, err := tx.RemoveNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if err != nil {
				return fmt.Errorf("failed to remove role %s from user %s: %w", role, userCode, err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(roleCodes), nil
}

// AssignUsersToRole assigns a list of users to a role (reverse assignment).
// It enforces the User-Role grouping policy (`GTypeUserRole`).
// It checks if any user is already assigned and aborts the entire transaction on conflict.
func AssignUsersToRole(ctx *gin.Context, roleCode string, userCodes []string) (int, error) {
	if enforcer == nil {
		return 0, ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Pre-check for existing users to fail early and atomically.
	for _, user := range userCodes {
		userPrefixed := addPrefix(user, UserCode)
		exists, _ := enforcer.HasNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
		if exists {
			return 0, fmt.Errorf("%w: user %s to role %s", ErrUserAlreadyAssigned, user, roleCode)
		}
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, user := range userCodes {
			userPrefixed := addPrefix(user, UserCode)
			// Grouping policy is always (user, role) regardless of which direction the assignment is made from.
			_, err := tx.AddNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if err != nil {
				return fmt.Errorf("failed to assign user %s to role %s: %w", user, roleCode, err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(userCodes), nil
}

// RemoveUsersFromRole removes a list of users from a role.
// It checks if all users are currently assigned and aborts the entire transaction on conflict.
func RemoveUsersFromRole(ctx *gin.Context, roleCode string, userCodes []string) (int, error) {
	if enforcer == nil {
		return 0, ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Pre-check for missing users to fail early and atomically.
	for _, user := range userCodes {
		userPrefixed := addPrefix(user, UserCode)
		exists, _ := enforcer.HasNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
		if !exists {
			return 0, fmt.Errorf("%w: user %s from role %s", ErrUserNotAssigned, user, roleCode)
		}
	}

	err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, user := range userCodes {
			userPrefixed := addPrefix(user, UserCode)
			// Grouping policy is always (user, role).
			_, err := tx.RemoveNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if err != nil {
				return fmt.Errorf("failed to remove user %s from role %s: %w", user, roleCode, err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(userCodes), nil
}

// GetUserRoles retrieves the list of roles assigned to a specific user.
// It removes the internal role prefix before returning the result.
func GetUserRoles(ctx *gin.Context, userCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// Filter by subject (user) at index 0 in the g policy.
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 0, userPrefixed)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user %s: %w", userCode, err)
	}

	roleSet := make(map[string]struct{})
	for _, p := range policies {
		if len(p) >= 2 {
			// Role is at index 1. Remove the prefix for the application layer.
			roleSet[removePrefix(p[1], RoleCode)] = struct{}{}
		}
	}

	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}

	return roles, nil
}

// GetRoleUsers retrieves the list of users assigned to a specific role.
// It removes the internal user prefix before returning the result.
func GetRoleUsers(ctx *gin.Context, roleCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Filter by group (role) at index 1 in the g policy.
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 1, rolePrefixed)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role %s: %w", roleCode, err)
	}

	userSet := make(map[string]struct{})
	for _, p := range policies {
		if len(p) >= 2 {
			// User is at index 0. Remove the prefix for the application layer.
			userSet[removePrefix(p[0], UserCode)] = struct{}{}
		}
	}

	users := make([]string, 0, len(userSet))
	for u := range userSet {
		users = append(users, u)
	}

	return users, nil
}

// AssignPoliciesToRole assigns a list of policies (obj, act) to a role subject.
// It requires `beginTime` and `endTime` strings to enforce time-based access.
// Policies are checked for existence and the operation is wrapped in a transaction.
func AssignPoliciesToRole(ctx *gin.Context, roleCode string, policies [][3]string, beginTime, endTime string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Pre-check for existing policies. The third element of the policy array is unused here.
	for _, pol := range policies {
		obj, act := pol[0], pol[1]
		exists, _ := enforcer.HasPolicy(rolePrefixed, obj, act, beginTime, endTime)
		if exists {
			return fmt.Errorf("%w for role %s: %s %s", ErrPolicyExists, roleCode, obj, act)
		}
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, pol := range policies {
			obj, act := pol[0], pol[1]
			// The subject is the prefixed role.
			_, err := tx.AddPolicy(rolePrefixed, obj, act, beginTime, endTime)
			if err != nil {
				return fmt.Errorf("failed to add policy for role %s: %w", roleCode, err)
			}
		}
		return nil
	})
}

// RemovePoliciesFromRole removes a list of policies (obj, act) from a role subject.
// It requires `beginTime` and `endTime` strings to match the policy being removed.
// Policies are checked for existence and the operation is wrapped in a transaction.
func RemovePoliciesFromRole(ctx *gin.Context, roleCode string, policies [][3]string, beginTime, endTime string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Pre-check for missing policies.
	for _, pol := range policies {
		obj, act := pol[0], pol[1]
		exists, _ := enforcer.HasPolicy(rolePrefixed, obj, act, beginTime, endTime)
		if !exists {
			return fmt.Errorf("%w for role %s: %s %s", ErrPolicyNotExist, roleCode, obj, act)
		}
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, pol := range policies {
			obj, act := pol[0], pol[1]
			// The subject is the prefixed role.
			_, err := tx.RemovePolicy(rolePrefixed, obj, act, beginTime, endTime)
			if err != nil {
				return fmt.Errorf("failed to remove policy for role %s: %w", roleCode, err)
			}
		}
		return nil
	})
}

// GetPoliciesForRole retrieves all policies directly assigned to a specific role.
// The role is queried by its prefixed code in the subject field (index 0).
func GetPoliciesForRole(ctx *gin.Context, roleCode string) ([][]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	// Filter by role subject at index 0.
	policies, err := enforcer.GetFilteredPolicy(0, rolePrefixed)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies for role %s: %w", roleCode, err)
	}

	return policies, nil
}

// AssignPoliciesToUser assigns a list of policies (obj, act) directly to a user subject.
// It requires `beginTime` and `endTime` strings to enforce time-based access.
// Policies are checked for existence and the operation is wrapped in a transaction.
func AssignPoliciesToUser(ctx *gin.Context, userCode string, policies [][2]string, beginTime, endTime string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// Pre-check for existing policies.
	for _, pol := range policies {
		obj, act := pol[0], pol[1]
		exists, _ := enforcer.HasPolicy(userPrefixed, obj, act, beginTime, endTime)
		if exists {
			return fmt.Errorf("%w for user %s: %s %s", ErrPolicyExists, userCode, obj, act)
		}
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, pol := range policies {
			obj, act := pol[0], pol[1]
			// The subject is the prefixed user.
			_, err := tx.AddPolicy(userPrefixed, obj, act, beginTime, endTime)
			if err != nil {
				return fmt.Errorf("failed to add policy for user %s: %w", userCode, err)
			}
		}
		return nil
	})
}

// RemovePoliciesFromUser removes a list of policies (obj, act) directly from a user subject.
// It requires `beginTime` and `endTime` strings to match the policy being removed.
// Policies are checked for existence and the operation is wrapped in a transaction.
func RemovePoliciesFromUser(ctx *gin.Context, userCode string, policies [][2]string, beginTime, endTime string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// Pre-check for missing policies.
	for _, pol := range policies {
		obj, act := pol[0], pol[1]
		exists, _ := enforcer.HasPolicy(userPrefixed, obj, act, beginTime, endTime)
		if !exists {
			return fmt.Errorf("%w for user %s: %s %s", ErrPolicyNotExist, userCode, obj, act)
		}
	}

	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, pol := range policies {
			obj, act := pol[0], pol[1]
			// The subject is the prefixed user.
			_, err := tx.RemovePolicy(userPrefixed, obj, act, beginTime, endTime)
			if err != nil {
				return fmt.Errorf("failed to remove policy for user %s: %w", userCode, err)
			}
		}
		return nil
	})
}

// GetPoliciesForUser retrieves all effective policies for a user.
// This includes policies directly assigned to the user AND policies inherited from all assigned roles.
// This requires two policy lookups (direct user policies and policies per role) and concatenation.
func GetPoliciesForUser(ctx *gin.Context, userCode string) ([][]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	userPrefixed := addPrefix(userCode, UserCode)
	// 1. Get policies directly assigned to the user.
	userPolicies, err := enforcer.GetFilteredPolicy(0, userPrefixed)
	if err != nil {
		return nil, err
	}

	// 2. Get the roles assigned to the user.
	roles, err := GetUserRoles(ctx, userCode)
	if err != nil {
		return nil, err
	}

	// 3. Collect policies for each role.
	var rolePolicies [][]string
	for _, role := range roles {
		pols, err := GetPoliciesForRole(ctx, role)
		if err != nil {
			return nil, err
		}
		rolePolicies = append(rolePolicies, pols...)
	}

	// Combine direct user policies and inherited role policies.
	return append(userPolicies, rolePolicies...), nil
}
