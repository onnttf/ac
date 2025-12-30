package casbin

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/model"

	casbin "github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
)

// Global enforcer with thread-safe initialization
var (
	enforcer *casbin.TransactionalEnforcer
	initOnce sync.Once
	initErr  error
)

// Casbin grouping policy identifiers
const (
	GroupingUserRole    = "g"  // User-to-Role inheritance
	GroupingObjectGroup = "g2" // Object-to-Group hierarchy
)

// Entity type identifiers for validation
const (
	EntityUser   = "user"
	EntityRole   = "role"
	EntityObject = "object"
)

// Entity prefixes for Casbin policy identification
const (
	PrefixUser   = "u"
	PrefixRole   = "r"
	PrefixObject = "o"

	PrefixSeparator = ":"
)

// Domain-specific errors for policy and role management
var (
	ErrEnforcerNotInitialized = fmt.Errorf("casbin enforcer not initialized")
	ErrPolicyAlreadyExists    = fmt.Errorf("policy already exists")
	ErrPolicyNotFound         = fmt.Errorf("policy not found")
	ErrGroupingAlreadyExists  = fmt.Errorf("grouping already exists")
	ErrGroupingNotFound       = fmt.Errorf("grouping not found")
	ErrRoleAlreadyAssigned    = fmt.Errorf("role already assigned to user")
	ErrRoleNotAssigned        = fmt.Errorf("role not assigned to user")
	ErrUserAlreadyAssigned    = fmt.Errorf("user already assigned to role")
	ErrUserNotAssigned        = fmt.Errorf("user not assigned to role")
	ErrInvalidTimeRange       = fmt.Errorf("end time must be after begin time")
	ErrInvalidPolicyFields    = fmt.Errorf("object and action are required")
	ErrInvalidUserCode        = fmt.Errorf("user code cannot be empty")
	ErrInvalidRoleCode        = fmt.Errorf("role code cannot be empty")
	ErrInvalidGroupCode       = fmt.Errorf("group code cannot be empty")
	ErrInvalidObjectCode      = fmt.Errorf("object code cannot be empty")
)

// Initialize creates the Casbin enforcer with RBAC model and GORM adapter.
// Thread-safe - can be called multiple times without side effects.
func Initialize() error {
	initOnce.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: started\n")

		gormAdapter.TurnOffAutoMigrate(database.DB)

		adapter, err := gormAdapter.NewAdapterByDBWithCustomTable(
			database.DB,
			&model.TblCasbinRule{},
			model.TableNameTblCasbinRule,
		)
		if err != nil {
			initErr = fmt.Errorf("failed to create Gorm Adapter: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create adapter failed: %v\n", err)
			return
		}

		m, err := casbinModel.NewModelFromString(`
[request_definition]
r = sub, obj, act, time

[policy_definition]
p = sub, obj, act, begin_time, end_time

[role_definition]
g = _, _ 
g2 = _, _ 

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

		enforcer, err = casbin.NewTransactionalEnforcer(m, adapter)
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin enforcer: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create enforcer failed: %v\n", err)
			return
		}

		fmt.Fprintf(os.Stdout, "INFO: casbin: init: succeeded, model=memory, policy_table=tbl_casbin_rule\n")
	})
	return initErr
}

// LoadPolicy refreshes the in-memory policy cache from database.
// Use after direct database modifications or cache inconsistencies.
func LoadPolicy(ctx *gin.Context) error {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: load policy failed: enforcer not initialized")
		return ErrEnforcerNotInitialized
	}

	logger.Infof(ctx, "casbin: loading policies from database")
	if err := enforcer.LoadPolicy(); err != nil {
		logger.Errorf(ctx, "casbin: load policy failed: error=%v", err)
		return fmt.Errorf("failed to load policies: %w", err)
	}
	logger.Infof(ctx, "casbin: policies loaded successfully")
	return nil
}

// Enforce performs authorization check with time-based policy evaluation.
// Returns true if access is granted, false if denied.
func Enforce(ctx *gin.Context, subject, object, action string, currentTime time.Time) (bool, error) {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: enforce check failed: enforcer not initialized")
		return false, ErrEnforcerNotInitialized
	}

	timeStr := formatTime(currentTime)
	logger.Debugf(ctx, "casbin: enforce check: subject=%s, object=%s, action=%s, time=%s", subject, object, action, timeStr)
	allowed, err := enforcer.Enforce(subject, object, action, timeStr)
	if err != nil {
		logger.Errorf(ctx, "casbin: enforce check failed: subject=%s, object=%s, action=%s, time=%s, error=%v", subject, object, action, timeStr, err)
		return false, fmt.Errorf("enforce check failed (subject=%s, object=%s, action=%s, time=%s): %w",
			subject, object, action, timeStr, err)
	}
	logger.Debugf(ctx, "casbin: enforce result: allowed=%v, subject=%s, object=%s, action=%s", allowed, subject, object, action)
	return allowed, nil
}

// formatTime standardizes time format for Casbin policy evaluation.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// addPrefix creates Casbin-compatible entity identifiers.
func addPrefix(code, entityType string) string {
	switch entityType {
	case EntityUser:
		return PrefixUser + PrefixSeparator + code
	case EntityRole:
		return PrefixRole + PrefixSeparator + code
	case EntityObject:
		return PrefixObject + PrefixSeparator + code
	default:
		return code
	}
}

// removePrefix extracts raw entity codes from Casbin identifiers.
func removePrefix(code, entityType string) string {
	switch entityType {
	case EntityUser:
		return strings.TrimPrefix(code, PrefixUser+PrefixSeparator)
	case EntityRole:
		return strings.TrimPrefix(code, PrefixRole+PrefixSeparator)
	case EntityObject:
		return strings.TrimPrefix(code, PrefixObject+PrefixSeparator)
	default:
		return code
	}
}

// validateCode prevents empty entity identifiers in policy operations.
func validateCode(code, entityType string) error {
	if strings.TrimSpace(code) == "" {
		switch entityType {
		case EntityUser:
			return ErrInvalidUserCode
		case EntityRole:
			return ErrInvalidRoleCode
		case EntityObject:
			return ErrInvalidObjectCode
		default:
			return fmt.Errorf("invalid entity type: %s", entityType)
		}
	}
	return nil
}

// validateCodes validates entity code batches for bulk operations.
func validateCodes(codes []string, entityType string) error {
	if len(codes) == 0 {
		switch entityType {
		case EntityUser:
			return ErrInvalidUserCode
		case EntityRole:
			return ErrInvalidRoleCode
		case EntityObject:
			return ErrInvalidObjectCode
		default:
			return fmt.Errorf("invalid entity type: %s", entityType)
		}
	}

	for _, code := range codes {
		if err := validateCode(code, entityType); err != nil {
			return err
		}
	}
	return nil
}

// buildPolicyKey generates composite keys for policy deduplication.
func buildPolicyKey(object, action string) string {
	return object + "|" + action
}

// assignPoliciesToSubject grants access policies to users or roles.
// Performs duplicate detection and validates time ranges atomically.
func assignPoliciesToSubject(ctx *gin.Context, subjectCode, subjectType string, policies []Policy) error {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: assign policies failed: enforcer not initialized")
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(subjectCode, subjectType); err != nil {
		logger.Errorf(ctx, "casbin: assign policies validation failed: subject_type=%s, subject_code=%s, error=%v", subjectType, subjectCode, err)
		return err
	}

	if subjectType != EntityUser && subjectType != EntityRole {
		logger.Errorf(ctx, "casbin: assign policies failed: invalid subject type=%s", subjectType)
		return ErrInvalidPolicyFields
	}

	subjectWithPrefix := addPrefix(subjectCode, subjectType)
	logger.Infof(ctx, "casbin: assigning policies: subject_type=%s, subject=%s, policy_count=%d", subjectType, subjectWithPrefix, len(policies))

	existingPolicies, err := enforcer.GetFilteredPolicy(0, subjectWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get existing policies: subject=%s, error=%v", subjectWithPrefix, err)
		return fmt.Errorf("failed to get existing policies for %s: %w", subjectWithPrefix, err)
	}

	// Build map of existing policies by object+action
	existingPolicyMap := make(map[string]struct{})
	for _, policyFields := range existingPolicies {
		if len(policyFields) >= 3 {
			key := buildPolicyKey(policyFields[1], policyFields[2])
			existingPolicyMap[key] = struct{}{}
		}
	}

	// Validate all policies before adding any
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policy := policies[i]

		if err := policy.Validate(); err != nil {
			logger.Errorf(ctx, "casbin: policy validation failed: subject=%s, object=%s, action=%s, begin=%s, end=%s, error=%v",
				subjectWithPrefix, policy.Object, policy.Action,
				formatTime(policy.BeginTime), formatTime(policy.EndTime), err)
			return fmt.Errorf("invalid policy (subject=%s, object=%s, action=%s, begin=%s, end=%s): %w",
				subjectWithPrefix, policy.Object, policy.Action,
				formatTime(policy.BeginTime), formatTime(policy.EndTime), err)
		}

		key := buildPolicyKey(policy.Object, policy.Action)
		if _, exists := existingPolicyMap[key]; exists {
			logger.Warnf(ctx, "casbin: policy already exists: subject=%s, object=%s, action=%s", subjectWithPrefix, policy.Object, policy.Action)
			return fmt.Errorf("policy already exists (subject=%s, object=%s, action=%s): %w",
				subjectWithPrefix, policy.Object, policy.Action, ErrPolicyAlreadyExists)
		}

		// Prevent duplicates within same request
		existingPolicyMap[key] = struct{}{}
	}

	// Add all policies in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, policy := range policies {
			_, err := tx.AddPolicy(
				subjectWithPrefix,
				policy.Object,
				policy.Action,
				formatTime(policy.BeginTime),
				formatTime(policy.EndTime),
			)
			if err != nil {
				logger.Errorf(ctx, "casbin: failed to add policy: subject=%s, object=%s, action=%s, error=%v",
					subjectWithPrefix, policy.Object, policy.Action, err)
				return fmt.Errorf("failed to add policy (subject=%s, object=%s, action=%s): %w",
					subjectWithPrefix, policy.Object, policy.Action, err)
			}
			logger.Debugf(ctx, "casbin: policy added successfully: subject=%s, object=%s, action=%s, begin=%s, end=%s",
				subjectWithPrefix, policy.Object, policy.Action,
				formatTime(policy.BeginTime), formatTime(policy.EndTime))
		}
		logger.Infof(ctx, "casbin: policies assigned successfully: subject=%s, count=%d", subjectWithPrefix, len(policies))
		return nil
	})
}

// removePoliciesFromSubject revokes specific policies from users or roles.
// Validates policy existence before removal to prevent silent failures.
func removePoliciesFromSubject(ctx *gin.Context, subjectCode, subjectType string, policies []Policy) error {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: remove policies failed: enforcer not initialized")
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(subjectCode, subjectType); err != nil {
		logger.Errorf(ctx, "casbin: remove policies validation failed: subject_type=%s, subject_code=%s, error=%v", subjectType, subjectCode, err)
		return err
	}

	subjectWithPrefix := addPrefix(subjectCode, subjectType)
	logger.Infof(ctx, "casbin: removing policies: subject_type=%s, subject=%s, policy_count=%d", subjectType, subjectWithPrefix, len(policies))

	existingPolicies, err := enforcer.GetFilteredPolicy(0, subjectWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get existing policies for removal: subject=%s, error=%v", subjectWithPrefix, err)
		return fmt.Errorf("failed to get policies for %s: %w", subjectWithPrefix, err)
	}

	existingPolicyMap := make(map[string][]string)
	for _, policyFields := range existingPolicies {
		if len(policyFields) >= 5 {
			key := buildPolicyKey(policyFields[1], policyFields[2])
			existingPolicyMap[key] = policyFields
		}
	}

	// Validate all policies exist before removing any
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policy := policies[i]

		if err := policy.Validate(); err != nil {
			logger.Errorf(ctx, "casbin: policy validation failed during removal: subject=%s, object=%s, action=%s, begin=%s, end=%s, error=%v",
				subjectWithPrefix, policy.Object, policy.Action,
				formatTime(policy.BeginTime), formatTime(policy.EndTime), err)
			return fmt.Errorf("invalid policy (subject=%s, object=%s, action=%s, begin=%s, end=%s): %w",
				subjectWithPrefix, policy.Object, policy.Action,
				formatTime(policy.BeginTime), formatTime(policy.EndTime), err)
		}

		key := buildPolicyKey(policy.Object, policy.Action)
		if _, exists := existingPolicyMap[key]; !exists {
			logger.Warnf(ctx, "casbin: policy not found for removal: subject=%s, object=%s, action=%s", subjectWithPrefix, policy.Object, policy.Action)
			return fmt.Errorf("policy not found (subject=%s, object=%s, action=%s): %w",
				subjectWithPrefix, policy.Object, policy.Action, ErrPolicyNotFound)
		}
	}

	// Remove all policies in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, policy := range policies {
			_, err := tx.RemovePolicy(
				subjectWithPrefix,
				policy.Object,
				policy.Action,
				formatTime(policy.BeginTime),
				formatTime(policy.EndTime),
			)
			if err != nil {
				logger.Errorf(ctx, "casbin: failed to remove policy: subject=%s, object=%s, action=%s, error=%v",
					subjectWithPrefix, policy.Object, policy.Action, err)
				return fmt.Errorf("failed to remove policy (subject=%s, object=%s, action=%s): %w",
					subjectWithPrefix, policy.Object, policy.Action, err)
			}
			logger.Debugf(ctx, "casbin: policy removed successfully: subject=%s, object=%s, action=%s",
				subjectWithPrefix, policy.Object, policy.Action)
		}
		logger.Infof(ctx, "casbin: policies removed successfully: subject=%s, count=%d", subjectWithPrefix, len(policies))
		return nil
	})
}

// AssignRolesToUser grants multiple roles to a user atomically.
// Prevents duplicate role assignments within the same transaction.
func AssignRolesToUser(ctx *gin.Context, userCode string, roleCodes []string) error {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: assign roles to user failed: enforcer not initialized")
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(userCode, EntityUser); err != nil {
		logger.Errorf(ctx, "casbin: assign roles to user validation failed: user_code=%s, error=%v", userCode, err)
		return err
	}

	if err := validateCodes(roleCodes, EntityRole); err != nil {
		logger.Errorf(ctx, "casbin: assign roles to user validation failed: role_codes=%v, error=%v", roleCodes, err)
		return err
	}

	userWithPrefix := addPrefix(userCode, EntityUser)
	logger.Infof(ctx, "casbin: assigning roles to user: user=%s, role_count=%d", userWithPrefix, len(roleCodes))

	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingUserRole, 0, userWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get existing roles for user: user=%s, error=%v", userWithPrefix, err)
		return fmt.Errorf("failed to get existing roles for user %s: %w", userWithPrefix, err)
	}

	existingRoles := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingRoles[grouping[1]] = struct{}{}
		}
	}

	// Check for duplicates before adding
	for _, roleCode := range roleCodes {
		roleWithPrefix := addPrefix(roleCode, EntityRole)
		if _, exists := existingRoles[roleWithPrefix]; exists {
			logger.Warnf(ctx, "casbin: role already assigned to user: user=%s, role=%s", userWithPrefix, roleWithPrefix)
			return fmt.Errorf("role already assigned (user=%s, role=%s): %w",
				userWithPrefix, roleWithPrefix, ErrRoleAlreadyAssigned)
		}
	}

	// Add all role assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, roleCode := range roleCodes {
			roleWithPrefix := addPrefix(roleCode, EntityRole)
			_, err := tx.AddNamedGroupingPolicy(GroupingUserRole, userWithPrefix, roleWithPrefix)
			if err != nil {
				logger.Errorf(ctx, "casbin: failed to assign role to user: user=%s, role=%s, error=%v", userWithPrefix, roleWithPrefix, err)
				return fmt.Errorf("failed to assign role (user=%s, role=%s): %w",
					userWithPrefix, roleWithPrefix, err)
			}
			logger.Debugf(ctx, "casbin: role assigned to user: user=%s, role=%s", userWithPrefix, roleWithPrefix)
		}
		logger.Infof(ctx, "casbin: roles assigned to user successfully: user=%s, count=%d", userWithPrefix, len(roleCodes))
		return nil
	})
}

// RemoveRolesFromUser revokes multiple roles from a user atomically.
// Validates role assignments before removal to catch inconsistencies.
func RemoveRolesFromUser(ctx *gin.Context, userCode string, roleCodes []string) error {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: remove roles from user failed: enforcer not initialized")
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(userCode, EntityUser); err != nil {
		logger.Errorf(ctx, "casbin: remove roles from user validation failed: user_code=%s, error=%v", userCode, err)
		return err
	}

	if err := validateCodes(roleCodes, EntityRole); err != nil {
		logger.Errorf(ctx, "casbin: remove roles from user validation failed: role_codes=%v, error=%v", roleCodes, err)
		return err
	}

	userWithPrefix := addPrefix(userCode, EntityUser)
	logger.Infof(ctx, "casbin: removing roles from user: user=%s, role_count=%d", userWithPrefix, len(roleCodes))

	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingUserRole, 0, userWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get existing roles for removal: user=%s, error=%v", userWithPrefix, err)
		return fmt.Errorf("failed to get existing roles for user %s: %w", userWithPrefix, err)
	}

	existingRoles := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingRoles[grouping[1]] = struct{}{}
		}
	}

	// Verify all roles exist before removing
	for _, roleCode := range roleCodes {
		roleWithPrefix := addPrefix(roleCode, EntityRole)
		if _, exists := existingRoles[roleWithPrefix]; !exists {
			logger.Warnf(ctx, "casbin: role not assigned to user for removal: user=%s, role=%s", userWithPrefix, roleWithPrefix)
			return fmt.Errorf("role not assigned (user=%s, role=%s): %w",
				userWithPrefix, roleWithPrefix, ErrRoleNotAssigned)
		}
	}

	// Remove all role assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, roleCode := range roleCodes {
			roleWithPrefix := addPrefix(roleCode, EntityRole)
			_, err := tx.RemoveNamedGroupingPolicy(GroupingUserRole, userWithPrefix, roleWithPrefix)
			if err != nil {
				logger.Errorf(ctx, "casbin: failed to remove role from user: user=%s, role=%s, error=%v", userWithPrefix, roleWithPrefix, err)
				return fmt.Errorf("failed to remove role (user=%s, role=%s): %w",
					userWithPrefix, roleWithPrefix, err)
			}
			logger.Debugf(ctx, "casbin: role removed from user: user=%s, role=%s", userWithPrefix, roleWithPrefix)
		}
		logger.Infof(ctx, "casbin: roles removed from user successfully: user=%s, count=%d", userWithPrefix, len(roleCodes))
		return nil
	})
}

// GetRolesForUser retrieves all roles for a user, including inherited ones.
// Returns deduplicated and sorted role codes.
func GetRolesForUser(ctx *gin.Context, userCode string) ([]string, error) {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: get roles for user failed: enforcer not initialized")
		return nil, ErrEnforcerNotInitialized
	}

	if err := validateCode(userCode, EntityUser); err != nil {
		logger.Errorf(ctx, "casbin: get roles for user validation failed: user_code=%s, error=%v", userCode, err)
		return nil, err
	}

	userWithPrefix := addPrefix(userCode, EntityUser)
	logger.Debugf(ctx, "casbin: getting roles for user: user=%s", userWithPrefix)

	rolesWithPrefix, err := enforcer.GetImplicitRolesForUser(userWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get roles for user: user=%s, error=%v", userWithPrefix, err)
		return nil, fmt.Errorf("failed to get roles for user %s: %w", userWithPrefix, err)
	}

	// Remove duplicates and prefixes
	roleSet := make(map[string]struct{}, len(rolesWithPrefix))
	for _, role := range rolesWithPrefix {
		roleSet[removePrefix(role, EntityRole)] = struct{}{}
	}

	roles := make([]string, 0, len(roleSet))
	for role := range roleSet {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	logger.Debugf(ctx, "casbin: retrieved roles for user: user=%s, role_count=%d, roles=%v", userWithPrefix, len(roles), roles)
	return roles, nil
}

// AssignUsersToRole grants a role to multiple users atomically.
// Performs reverse duplicate checking (user-to-role vs role-to-user).
func AssignUsersToRole(ctx *gin.Context, roleCode string, userCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(roleCode, EntityRole); err != nil {
		return err
	}

	if err := validateCodes(userCodes, EntityUser); err != nil {
		return err
	}

	roleWithPrefix := addPrefix(roleCode, EntityRole)
	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingUserRole, 1, roleWithPrefix)
	if err != nil {
		return fmt.Errorf("failed to get existing users for role %s: %w", roleWithPrefix, err)
	}

	existingUsers := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingUsers[grouping[0]] = struct{}{}
		}
	}

	// Check for duplicates before adding
	for _, userCode := range userCodes {
		userWithPrefix := addPrefix(userCode, EntityUser)
		if _, exists := existingUsers[userWithPrefix]; exists {
			return fmt.Errorf("user already assigned (role=%s, user=%s): %w",
				roleWithPrefix, userWithPrefix, ErrUserAlreadyAssigned)
		}
	}

	// Add all user assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, userCode := range userCodes {
			userWithPrefix := addPrefix(userCode, EntityUser)
			_, err := tx.AddNamedGroupingPolicy(GroupingUserRole, userWithPrefix, roleWithPrefix)
			if err != nil {
				return fmt.Errorf("failed to assign user (role=%s, user=%s): %w",
					roleWithPrefix, userWithPrefix, err)
			}
		}
		return nil
	})
}

// RemoveUsersFromRole revokes a role from multiple users atomically.
// Use for bulk user removal operations.
func RemoveUsersFromRole(ctx *gin.Context, roleCode string, userCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInitialized
	}

	if err := validateCode(roleCode, EntityRole); err != nil {
		return err
	}

	if err := validateCodes(userCodes, EntityUser); err != nil {
		return err
	}

	roleWithPrefix := addPrefix(roleCode, EntityRole)
	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingUserRole, 1, roleWithPrefix)
	if err != nil {
		return fmt.Errorf("failed to get existing users for role %s: %w", roleWithPrefix, err)
	}

	existingUsers := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingUsers[grouping[0]] = struct{}{}
		}
	}

	// Verify all users exist before removing
	for _, userCode := range userCodes {
		userWithPrefix := addPrefix(userCode, EntityUser)
		if _, exists := existingUsers[userWithPrefix]; !exists {
			return fmt.Errorf("user not assigned (role=%s, user=%s): %w",
				roleWithPrefix, userWithPrefix, ErrUserNotAssigned)
		}
	}

	// Remove all user assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, userCode := range userCodes {
			userWithPrefix := addPrefix(userCode, EntityUser)
			_, err := tx.RemoveNamedGroupingPolicy(GroupingUserRole, userWithPrefix, roleWithPrefix)
			if err != nil {
				return fmt.Errorf("failed to remove user (role=%s, user=%s): %w",
					roleWithPrefix, userWithPrefix, err)
			}
		}
		return nil
	})
}

// GetUsersForRole retrieves all users for a role, including indirect assignments.
// Returns deduplicated and sorted user codes.
func GetUsersForRole(ctx *gin.Context, roleCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInitialized
	}

	if err := validateCode(roleCode, EntityRole); err != nil {
		return nil, err
	}

	roleWithPrefix := addPrefix(roleCode, EntityRole)

	usersWithPrefix, err := enforcer.GetImplicitUsersForRole(roleWithPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role %s: %w", roleWithPrefix, err)
	}

	// Remove duplicates and prefixes
	userSet := make(map[string]struct{}, len(usersWithPrefix))
	for _, user := range usersWithPrefix {
		userSet[removePrefix(user, EntityUser)] = struct{}{}
	}

	users := make([]string, 0, len(userSet))
	for user := range userSet {
		users = append(users, user)
	}
	sort.Strings(users)

	return users, nil
}

// AssignObjectsToGroup creates resource hierarchies by grouping objects.
// Enables inheritance-based access control for object collections.
func AssignObjectsToGroup(ctx *gin.Context, groupCode string, objectCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInitialized
	}

	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return ErrInvalidGroupCode
	}

	if err := validateCodes(objectCodes, EntityObject); err != nil {
		return err
	}

	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingObjectGroup, 1, groupCode)
	if err != nil {
		return fmt.Errorf("failed to get existing objects for group %s: %w", groupCode, err)
	}

	existingObjects := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingObjects[grouping[0]] = struct{}{}
		}
	}

	// Check for duplicates before adding
	for _, objectCode := range objectCodes {
		if _, exists := existingObjects[objectCode]; exists {
			return fmt.Errorf("object already in group (group=%s, object=%s): %w",
				groupCode, objectCode, ErrGroupingAlreadyExists)
		}
	}

	// Add all object assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, objectCode := range objectCodes {
			_, err := tx.AddNamedGroupingPolicy(GroupingObjectGroup, objectCode, groupCode)
			if err != nil {
				return fmt.Errorf("failed to add object to group (group=%s, object=%s): %w",
					groupCode, objectCode, err)
			}
		}
		return nil
	})
}

// RemoveObjectsFromGroup dissolves resource hierarchies by removing objects from groups.
func RemoveObjectsFromGroup(ctx *gin.Context, groupCode string, objectCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInitialized
	}

	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return ErrInvalidGroupCode
	}

	if err := validateCodes(objectCodes, EntityObject); err != nil {
		return err
	}

	existingGroupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingObjectGroup, 1, groupCode)
	if err != nil {
		return fmt.Errorf("failed to get existing objects for group %s: %w", groupCode, err)
	}

	existingObjects := make(map[string]struct{})
	for _, grouping := range existingGroupings {
		if len(grouping) >= 2 {
			existingObjects[grouping[0]] = struct{}{}
		}
	}

	// Verify all objects exist before removing
	for _, objectCode := range objectCodes {
		if _, exists := existingObjects[objectCode]; !exists {
			return fmt.Errorf("object not in group (group=%s, object=%s): %w",
				groupCode, objectCode, ErrGroupingNotFound)
		}
	}

	// Remove all object assignments in transaction
	return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		for _, objectCode := range objectCodes {
			_, err := tx.RemoveNamedGroupingPolicy(GroupingObjectGroup, objectCode, groupCode)
			if err != nil {
				return fmt.Errorf("failed to remove object from group (group=%s, object=%s): %w",
					groupCode, objectCode, err)
			}
		}
		return nil
	})
}

// GetGroupsForObject retrieves all groups containing an object.
// Useful for debugging resource hierarchy access issues.
func GetGroupsForObject(ctx *gin.Context, objectCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInitialized
	}

	if err := validateCode(objectCode, EntityObject); err != nil {
		return nil, err
	}

	groupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingObjectGroup, 0, objectCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for object %s: %w", objectCode, err)
	}

	groupSet := make(map[string]struct{}, len(groupings))
	for _, grouping := range groupings {
		if len(grouping) >= 2 {
			groupSet[grouping[1]] = struct{}{}
		}
	}

	groups := make([]string, 0, len(groupSet))
	for group := range groupSet {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	return groups, nil
}

// GetObjectsForGroup retrieves all objects within a specific group.
// Returns sorted list for consistent ordering.
func GetObjectsForGroup(ctx *gin.Context, groupCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInitialized
	}

	if groupCode == "" {
		return nil, ErrInvalidGroupCode
	}

	groupings, err := enforcer.GetFilteredNamedGroupingPolicy(GroupingObjectGroup, 1, groupCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get objects for group %s: %w", groupCode, err)
	}

	objectSet := make(map[string]struct{}, len(groupings))
	for _, grouping := range groupings {
		if len(grouping) >= 2 {
			objectSet[grouping[0]] = struct{}{}
		}
	}

	objects := make([]string, 0, len(objectSet))
	for object := range objectSet {
		objects = append(objects, object)
	}
	sort.Strings(objects)

	return objects, nil
}

// AssignPoliciesToRole delegates access permissions through role-based policies.
func AssignPoliciesToRole(ctx *gin.Context, roleCode string, policies []Policy) error {
	return assignPoliciesToSubject(ctx, roleCode, EntityRole, policies)
}

// RemovePoliciesFromRole revokes role-based access permissions.
func RemovePoliciesFromRole(ctx *gin.Context, roleCode string, policies []Policy) error {
	return removePoliciesFromSubject(ctx, roleCode, EntityRole, policies)
}

// GetPoliciesForRole retrieves active policies for a role.
// Automatically filters expired policies and returns sorted results.
func GetPoliciesForRole(ctx *gin.Context, roleCode string) ([]Policy, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInitialized
	}

	if err := validateCode(roleCode, EntityRole); err != nil {
		return nil, err
	}

	roleWithPrefix := addPrefix(roleCode, EntityRole)
	policyFields, err := enforcer.GetFilteredPolicy(0, roleWithPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies for role %s: %w", roleWithPrefix, err)
	}

	policies := make([]Policy, 0, len(policyFields))
	for _, fields := range policyFields {
		if len(fields) < 5 {
			continue
		}

		beginTime, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			return nil, fmt.Errorf("failed to parse begin time %s: %w", fields[3], err)
		}

		endTime, err := time.Parse(time.RFC3339, fields[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time %s: %w", fields[4], err)
		}

		policies = append(policies, Policy{
			Object:    fields[1],
			Action:    fields[2],
			BeginTime: beginTime,
			EndTime:   endTime,
		})
	}

	// Filter out expired policies
	policies = filterExpiredPolicies(policies, time.Now())

	// Sort by object, action, begin time, end time
	sort.Slice(policies, func(i, j int) bool {
		if policies[i].Object != policies[j].Object {
			return policies[i].Object < policies[j].Object
		}
		if policies[i].Action != policies[j].Action {
			return policies[i].Action < policies[j].Action
		}
		if !policies[i].BeginTime.Equal(policies[j].BeginTime) {
			return policies[i].BeginTime.Before(policies[j].BeginTime)
		}
		return policies[i].EndTime.Before(policies[j].EndTime)
	})

	return policies, nil
}

// AssignPoliciesToUser grants direct user permissions bypassing role hierarchy.
// Use sparingly - prefer role-based assignments for better maintainability.
func AssignPoliciesToUser(ctx *gin.Context, userCode string, policies []Policy) error {
	return assignPoliciesToSubject(ctx, userCode, EntityUser, policies)
}

// RemovePoliciesFromUser revokes direct user permissions.
// Does not affect inherited permissions from roles.
func RemovePoliciesFromUser(ctx *gin.Context, userCode string, policies []Policy) error {
	return removePoliciesFromSubject(ctx, userCode, EntityUser, policies)
}

// GetPoliciesForUser retrieves all effective policies for a user.
// Combines direct and role-inherited permissions, filtering expired policies.
func GetPoliciesForUser(ctx *gin.Context, userCode string) ([]Policy, error) {
	if enforcer == nil {
		logger.Errorf(ctx, "casbin: get policies for user failed: enforcer not initialized")
		return nil, ErrEnforcerNotInitialized
	}

	if err := validateCode(userCode, EntityUser); err != nil {
		logger.Errorf(ctx, "casbin: get policies for user validation failed: user_code=%s, error=%v", userCode, err)
		return nil, err
	}

	userWithPrefix := addPrefix(userCode, EntityUser)
	logger.Debugf(ctx, "casbin: getting policies for user: user=%s", userWithPrefix)

	policyFields, err := enforcer.GetImplicitPermissionsForUser(userWithPrefix)
	if err != nil {
		logger.Errorf(ctx, "casbin: failed to get policies for user: user=%s, error=%v", userWithPrefix, err)
		return nil, fmt.Errorf("failed to get policies for user %s: %w", userWithPrefix, err)
	}

	policies := make([]Policy, 0, len(policyFields))
	for _, fields := range policyFields {
		if len(fields) < 5 {
			continue
		}

		beginTime, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			logger.Errorf(ctx, "casbin: failed to parse begin time: time=%s, error=%v", fields[3], err)
			return nil, fmt.Errorf("failed to parse begin time %s: %w", fields[3], err)
		}

		endTime, err := time.Parse(time.RFC3339, fields[4])
		if err != nil {
			logger.Errorf(ctx, "casbin: failed to parse end time: time=%s, error=%v", fields[4], err)
			return nil, fmt.Errorf("failed to parse end time %s: %w", fields[4], err)
		}

		policies = append(policies, Policy{
			Object:    fields[1],
			Action:    fields[2],
			BeginTime: beginTime,
			EndTime:   endTime,
		})
	}

	// Filter out expired policies
	policies = filterExpiredPolicies(policies, time.Now())

	// Sort by object, action, begin time, end time
	sort.Slice(policies, func(i, j int) bool {
		if policies[i].Object != policies[j].Object {
			return policies[i].Object < policies[j].Object
		}
		if policies[i].Action != policies[j].Action {
			return policies[i].Action < policies[j].Action
		}
		if !policies[i].BeginTime.Equal(policies[j].BeginTime) {
			return policies[i].BeginTime.Before(policies[j].BeginTime)
		}
		return policies[i].EndTime.Before(policies[j].EndTime)
	})

	logger.Debugf(ctx, "casbin: retrieved policies for user: user=%s, policy_count=%d", userWithPrefix, len(policies))
	return policies, nil
}

// filterExpiredPolicies removes expired access policies based on current time.
// Includes policies ending exactly at current time.
func filterExpiredPolicies(policies []Policy, currentTime time.Time) []Policy {
	activePolicies := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if currentTime.Before(policy.EndTime) || currentTime.Equal(policy.EndTime) {
			activePolicies = append(activePolicies, policy)
		}
	}
	return activePolicies
}

// Policy defines time-bound access control rules.
// Supports temporal permissions with automatic expiration.
type Policy struct {
	Object    string    // Resource identifier
	Action    string    // Permission type (read, write, etc.)
	BeginTime time.Time // Policy start time (UTC)
	EndTime   time.Time // Policy end time (UTC)
}

// Validate checks policy integrity and time constraints.
// Returns error for missing required fields or invalid time ranges.
func (p Policy) Validate() error {
	if p.Object == "" || p.Action == "" {
		return ErrInvalidPolicyFields
	}
	if !p.EndTime.After(p.BeginTime) {
		return ErrInvalidTimeRange
	}
	return nil
}
