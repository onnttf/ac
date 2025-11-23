package casbin

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"ac/bootstrap/database"

	casbin "github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
)

// enforcer is the global TransactionalEnforcer instance guarded by once.
var (
	enforcer *casbin.TransactionalEnforcer
	once     sync.Once
	initErr  error
)

const (
	GTypeUserRole    = "g"
	GTypeObjectGroup = "g2"

	UserCode = "user"
	RoleCode = "role"
	MenuCode = "menu"

	PrefixUser = "u:"
	PrefixRole = "r:"
	PrefixMenu = "o:"
)

// Package-level error values for consistent error handling.
var (
	ErrEnforcerNotInit     = fmt.Errorf("casbin enforcer not initialized")
	ErrPolicyExists        = fmt.Errorf("policy already exists")
	ErrPolicyNotExist      = fmt.Errorf("policy does not exist")
	ErrGroupingExists      = fmt.Errorf("grouping policy already exists")
	ErrGroupingNotExist    = fmt.Errorf("grouping policy does not exist")
	ErrRoleAlreadyAssigned = fmt.Errorf("role already assigned to user")
	ErrRoleNotAssigned     = fmt.Errorf("role not assigned to user")
	ErrUserAlreadyAssigned = fmt.Errorf("user already assigned to role")
	ErrUserNotAssigned     = fmt.Errorf("user not assigned to role")
	ErrPolicyTimeOverlap   = fmt.Errorf("policy time window overlaps")
	ErrInvalidTimeRange    = fmt.Errorf("invalid time range")
	ErrInvalidPolicyFields = fmt.Errorf("invalid policy fields")
	ErrInvalidUserCode     = fmt.Errorf("invalid user code")
	ErrInvalidRoleCode     = fmt.Errorf("invalid role code")
	ErrInvalidGroupCode    = fmt.Errorf("invalid group code")
	ErrInvalidObjectCode   = fmt.Errorf("invalid object code")
)

// Initialize prepares the global Casbin TransactionalEnforcer.
// It loads the model, sets up the Gorm adapter, and creates the enforcer.
func Initialize() error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: started\n")
		gormadapter.TurnOffAutoMigrate(database.DB)
		adapter, err := gormadapter.NewTransactionalAdapterByDB(database.DB)
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
		enforcer, err = casbin.NewTransactionalEnforcer(m, adapter)
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin enforcer: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create enforcer failed: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: succeeded, model=memory, policy_table=casbin_rule\n")
	})
	return initErr
}

// LoadPolicy reloads policies from persistent storage into the enforcer cache.
func LoadPolicy(context *gin.Context) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("load_policy: %w", err)
	}
	return nil
}

// Enforce evaluates whether a subject may perform an action on an object at a given time.
func Enforce(context *gin.Context, subject, object, action string, currentTime time.Time) (bool, error) {
	if enforcer == nil {
		return false, ErrEnforcerNotInit
	}
	ok, err := enforcer.Enforce(subject, object, action, formatTime(currentTime))
	if err != nil {
		return false, fmt.Errorf("enforce subject=%s object=%s action=%s time=%s: %w", subject, object, action, formatTime(currentTime), err)
	}
	return ok, nil
}

// formatTime converts a time to an RFC3339 UTC string.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// timeOverlap reports whether two time windows overlap.
func timeOverlap(b1, e1, b2, e2 time.Time) bool {
	if b1.After(e1) || b2.After(e2) {
		return false
	}
	return !(b1.After(e2) || b2.After(e1))
}

// addPrefix attaches a type prefix to a code.
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

// removePrefix strips a type prefix from a code.
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

// AssignRolesToUser assigns the given roles to a user in a single transaction.
func AssignRolesToUser(context *gin.Context, userCode string, roleCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return ErrInvalidUserCode
	}
	if len(roleCodes) == 0 {
		return ErrInvalidRoleCode
	}
	for i := range roleCodes {
		roleCodes[i] = strings.TrimSpace(roleCodes[i])
		if roleCodes[i] == "" {
			return ErrInvalidRoleCode
		}
	}

	userPrefixed := addPrefix(userCode, UserCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 0, userPrefixed)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[1]] = struct{}{}
		}
	}
	for _, role := range roleCodes {
		rolePrefixed := addPrefix(role, RoleCode)
		if _, ok := existing[rolePrefixed]; ok {
			return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, ErrRoleAlreadyAssigned)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, role := range roleCodes {
			rolePrefixed := addPrefix(role, RoleCode)
			_, err := transaction.AddNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if err != nil {
				return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RemoveRolesFromUser removes the given roles from a user in a single transaction.
func RemoveRolesFromUser(context *gin.Context, userCode string, roleCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return ErrInvalidUserCode
	}
	if len(roleCodes) == 0 {
		return ErrInvalidRoleCode
	}
	for i := range roleCodes {
		roleCodes[i] = strings.TrimSpace(roleCodes[i])
		if roleCodes[i] == "" {
			return ErrInvalidRoleCode
		}
	}

	userPrefixed := addPrefix(userCode, UserCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 0, userPrefixed)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[1]] = struct{}{}
		}
	}
	for _, role := range roleCodes {
		rolePrefixed := addPrefix(role, RoleCode)
		if _, ok := existing[rolePrefixed]; !ok {
			return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, ErrRoleNotAssigned)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, role := range roleCodes {
			rolePrefixed := addPrefix(role, RoleCode)
			_, transactionError := transaction.RemoveNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if transactionError != nil {
				return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, transactionError)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetRolesForUser returns the role codes currently assigned to a user.
func GetRolesForUser(context *gin.Context, userCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return nil, ErrInvalidUserCode
	}

	userPrefixed := addPrefix(userCode, UserCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 0, userPrefixed)
	if err != nil {
		return nil, fmt.Errorf("query_grouping type=%s subject=%s: %w", GTypeUserRole, userPrefixed, err)
	}

	roleSet := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			roleSet[removePrefix(policyEntry[1], RoleCode)] = struct{}{}
		}
	}

	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}
	sort.Strings(roles)

	return roles, nil
}

// AssignUsersToRole assigns the given users to a role in a single transaction.
func AssignUsersToRole(context *gin.Context, roleCode string, userCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return ErrInvalidRoleCode
	}
	if len(userCodes) == 0 {
		return ErrInvalidUserCode
	}
	for i := range userCodes {
		userCodes[i] = strings.TrimSpace(userCodes[i])
		if userCodes[i] == "" {
			return ErrInvalidUserCode
		}
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 1, rolePrefixed)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[0]] = struct{}{}
		}
	}
	for _, user := range userCodes {
		userPrefixed := addPrefix(user, UserCode)
		if _, ok := existing[userPrefixed]; ok {
			return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, ErrUserAlreadyAssigned)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, user := range userCodes {
			userPrefixed := addPrefix(user, UserCode)
			_, transactionError := transaction.AddNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if transactionError != nil {
				return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, transactionError)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RemoveUsersFromRole removes the given users from a role in a single transaction.
func RemoveUsersFromRole(context *gin.Context, roleCode string, userCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return ErrInvalidRoleCode
	}
	if len(userCodes) == 0 {
		return ErrInvalidUserCode
	}
	for i := range userCodes {
		userCodes[i] = strings.TrimSpace(userCodes[i])
		if userCodes[i] == "" {
			return ErrInvalidUserCode
		}
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 1, rolePrefixed)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[0]] = struct{}{}
		}
	}
	for _, user := range userCodes {
		userPrefixed := addPrefix(user, UserCode)
		if _, ok := existing[userPrefixed]; !ok {
			return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, ErrUserNotAssigned)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, user := range userCodes {
			userPrefixed := addPrefix(user, UserCode)
			_, transactionError := transaction.RemoveNamedGroupingPolicy(GTypeUserRole, userPrefixed, rolePrefixed)
			if transactionError != nil {
				return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeUserRole, userPrefixed, rolePrefixed, transactionError)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetUsersForRole returns the user codes currently assigned to a role.
func GetUsersForRole(context *gin.Context, roleCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil, ErrInvalidRoleCode
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeUserRole, 1, rolePrefixed)
	if err != nil {
		return nil, fmt.Errorf("query_grouping type=%s group=%s: %w", GTypeUserRole, rolePrefixed, err)
	}

	userSet := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			userSet[removePrefix(policyEntry[0], UserCode)] = struct{}{}
		}
	}

	users := make([]string, 0, len(userSet))
	for u := range userSet {
		users = append(users, u)
	}
	sort.Strings(users)

	return users, nil
}

// AssignObjectsToGroup assigns the given objects to a group in a single transaction.
func AssignObjectsToGroup(context *gin.Context, groupCode string, objectCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return ErrInvalidGroupCode
	}
	if len(objectCodes) == 0 {
		return ErrInvalidObjectCode
	}
	for i := range objectCodes {
		objectCodes[i] = strings.TrimSpace(objectCodes[i])
		if objectCodes[i] == "" {
			return ErrInvalidObjectCode
		}
	}

	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeObjectGroup, 1, groupCode)
	if err != nil {
		return fmt.Errorf("query_grouping type=%s group=%s: %w", GTypeObjectGroup, groupCode, err)
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[0]] = struct{}{}
		}
	}
	for _, object := range objectCodes {
		if _, ok := existing[object]; ok {
			return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeObjectGroup, object, groupCode, ErrGroupingExists)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, object := range objectCodes {
			_, txErr := transaction.AddNamedGroupingPolicy(GTypeObjectGroup, object, groupCode)
			if txErr != nil {
				return fmt.Errorf("add_group type=%s subject=%s group=%s: %w", GTypeObjectGroup, object, groupCode, txErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RemoveObjectsFromGroup removes the given objects from a group in a single transaction.
func RemoveObjectsFromGroup(context *gin.Context, groupCode string, objectCodes []string) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return ErrInvalidGroupCode
	}
	if len(objectCodes) == 0 {
		return ErrInvalidObjectCode
	}
	for i := range objectCodes {
		objectCodes[i] = strings.TrimSpace(objectCodes[i])
		if objectCodes[i] == "" {
			return ErrInvalidObjectCode
		}
	}

	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeObjectGroup, 1, groupCode)
	if err != nil {
		return fmt.Errorf("query_grouping type=%s group=%s: %w", GTypeObjectGroup, groupCode, err)
	}
	existing := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			existing[policyEntry[0]] = struct{}{}
		}
	}
	for _, object := range objectCodes {
		if _, ok := existing[object]; !ok {
			return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeObjectGroup, object, groupCode, ErrGroupingNotExist)
		}
	}

	err = enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, object := range objectCodes {
			_, txErr := transaction.RemoveNamedGroupingPolicy(GTypeObjectGroup, object, groupCode)
			if txErr != nil {
				return fmt.Errorf("remove_group type=%s subject=%s group=%s: %w", GTypeObjectGroup, object, groupCode, txErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetGroupsForObject returns the group codes an object currently belongs to.
func GetGroupsForObject(context *gin.Context, objectCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	objectCode = strings.TrimSpace(objectCode)
	if objectCode == "" {
		return nil, ErrInvalidObjectCode
	}

	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeObjectGroup, 0, objectCode)
	if err != nil {
		return nil, fmt.Errorf("query_grouping type=%s subject=%s: %w", GTypeObjectGroup, objectCode, err)
	}

	groupSet := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			groupSet[policyEntry[1]] = struct{}{}
		}
	}
	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups, nil
}

// GetObjectsForGroup returns the object codes currently in a group.
func GetObjectsForGroup(context *gin.Context, groupCode string) ([]string, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return nil, ErrInvalidGroupCode
	}

	policies, err := enforcer.GetFilteredNamedGroupingPolicy(GTypeObjectGroup, 1, groupCode)
	if err != nil {
		return nil, fmt.Errorf("query_grouping type=%s group=%s: %w", GTypeObjectGroup, groupCode, err)
	}

	objectSet := make(map[string]struct{})
	for _, policyEntry := range policies {
		if len(policyEntry) >= 2 {
			objectSet[policyEntry[0]] = struct{}{}
		}
	}
	objects := make([]string, 0, len(objectSet))
	for s := range objectSet {
		objects = append(objects, s)
	}
	sort.Strings(objects)
	return objects, nil
}

// AssignPoliciesToRole assigns object-action-time policies to a role.
func AssignPoliciesToRole(context *gin.Context, roleCode string, policies []Policy) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return ErrInvalidRoleCode
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	existingPolicies, err := enforcer.GetFilteredPolicy(0, rolePrefixed)
	if err != nil {
		return err
	}
	existSet := make(map[string]struct{})
	for _, policyEntry := range existingPolicies {
		if len(policyEntry) >= 5 {
			key := policyEntry[1] + "|" + policyEntry[2] + "|" + policyEntry[3] + "|" + policyEntry[4]
			existSet[key] = struct{}{}
		}
	}
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policyItem := policies[i]
		if err := policyItem.Validate(); err != nil {
			return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", rolePrefixed, policyItem.Object, policyItem.Action, formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime), err)
		}
		object, action := policyItem.Object, policyItem.Action
		beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
		key := object + "|" + action + "|" + beginStr + "|" + endStr
		if _, ok := existSet[key]; ok {
			return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", rolePrefixed, object, action, beginStr, endStr, ErrPolicyExists)
		}
		for _, policyEntry := range existingPolicies {
			if len(policyEntry) >= 5 && policyEntry[1] == object && policyEntry[2] == action {
				tb1, err := time.Parse(time.RFC3339, policyEntry[3])
				if err != nil {
					return err
				}
				te1, err := time.Parse(time.RFC3339, policyEntry[4])
				if err != nil {
					return err
				}
				overlap := timeOverlap(tb1, te1, policyItem.BeginTime, policyItem.EndTime)
				if overlap {
					return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", rolePrefixed, object, action, beginStr, endStr, ErrPolicyTimeOverlap)
				}
			}
		}
	}

	return enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, policyItem := range policies {
			object, action := policyItem.Object, policyItem.Action
			beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
			_, err := transaction.AddPolicy(rolePrefixed, object, action, beginStr, endStr)
			if err != nil {
				return fmt.Errorf("failed to add policy for role %s: %w", roleCode, err)
			}
		}
		return nil
	})
}

// RemovePoliciesFromRole removes object-action-time policies from a role.
func RemovePoliciesFromRole(context *gin.Context, roleCode string, policies []Policy) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return ErrInvalidRoleCode
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	existingPolicies, err := enforcer.GetFilteredPolicy(0, rolePrefixed)
	if err != nil {
		return err
	}
	existSet := make(map[string]struct{})
	for _, policyEntry := range existingPolicies {
		if len(policyEntry) >= 5 {
			key := policyEntry[1] + "|" + policyEntry[2] + "|" + policyEntry[3] + "|" + policyEntry[4]
			existSet[key] = struct{}{}
		}
	}
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policyItem := policies[i]
		object, action := policyItem.Object, policyItem.Action
		beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
		key := object + "|" + action + "|" + beginStr + "|" + endStr
		if _, ok := existSet[key]; !ok {
			return fmt.Errorf("remove_policy subject=%s object=%s action=%s begin=%s end=%s: %w", rolePrefixed, object, action, beginStr, endStr, ErrPolicyNotExist)
		}
	}

	return enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, policyItem := range policies {
			object, action := policyItem.Object, policyItem.Action
			beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
			_, err := transaction.RemovePolicy(rolePrefixed, object, action, beginStr, endStr)
			if err != nil {
				return fmt.Errorf("failed to remove policy for role %s: %w", roleCode, err)
			}
		}
		return nil
	})
}

// GetPoliciesForRole returns policies directly assigned to a role.
func GetPoliciesForRole(context *gin.Context, roleCode string) ([]Policy, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	if strings.TrimSpace(roleCode) == "" {
		return nil, ErrInvalidRoleCode
	}

	rolePrefixed := addPrefix(roleCode, RoleCode)
	policiesRaw, err := enforcer.GetFilteredPolicy(0, rolePrefixed)
	if err != nil {
		return nil, fmt.Errorf("query_policies subject=%s: %w", rolePrefixed, err)
	}
	var policies []Policy
	for _, entry := range policiesRaw {
		if len(entry) < 5 {
			continue
		}
		tb, err := time.Parse(time.RFC3339, entry[3])
		if err != nil {
			return nil, err
		}
		te, err := time.Parse(time.RFC3339, entry[4])
		if err != nil {
			return nil, err
		}
		policies = append(policies, Policy{Object: entry[1], Action: entry[2], BeginTime: tb, EndTime: te})
	}
	sort.Slice(policies, func(i, j int) bool {
		pi, pj := policies[i], policies[j]
		if pi.Object != pj.Object {
			return pi.Object < pj.Object
		}
		if pi.Action != pj.Action {
			return pi.Action < pj.Action
		}
		if !pi.BeginTime.Equal(pj.BeginTime) {
			return pi.BeginTime.Before(pj.BeginTime)
		}
		return pi.EndTime.Before(pj.EndTime)
	})
	return policies, nil
}

// AssignPoliciesToUser assigns object-action-time policies to a user.
func AssignPoliciesToUser(context *gin.Context, userCode string, policies []Policy) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return ErrInvalidUserCode
	}

	userPrefixed := addPrefix(userCode, UserCode)
	existingPolicies, err := enforcer.GetFilteredPolicy(0, userPrefixed)
	if err != nil {
		return fmt.Errorf("query_policies subject=%s: %w", userPrefixed, err)
	}
	existSet := make(map[string]struct{})
	for _, policyEntry := range existingPolicies {
		if len(policyEntry) >= 5 {
			key := policyEntry[1] + "|" + policyEntry[2] + "|" + policyEntry[3] + "|" + policyEntry[4]
			existSet[key] = struct{}{}
		}
	}
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policyItem := policies[i]
		if err := policyItem.Validate(); err != nil {
			return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, policyItem.Object, policyItem.Action, formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime), err)
		}
		object, action := policyItem.Object, policyItem.Action
		beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
		key := object + "|" + action + "|" + beginStr + "|" + endStr
		if _, ok := existSet[key]; ok {
			return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, object, action, beginStr, endStr, ErrPolicyExists)
		}
		for _, policyEntry := range existingPolicies {
			if len(policyEntry) >= 5 && policyEntry[1] == object && policyEntry[2] == action {
				tb1, err := time.Parse(time.RFC3339, policyEntry[3])
				if err != nil {
					return err
				}
				te1, err := time.Parse(time.RFC3339, policyEntry[4])
				if err != nil {
					return err
				}
				overlap := timeOverlap(tb1, te1, policyItem.BeginTime, policyItem.EndTime)
				if overlap {
					return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, object, action, formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime), ErrPolicyTimeOverlap)
				}
			}
		}
	}

	return enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, policyItem := range policies {
			object, action := policyItem.Object, policyItem.Action
			beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
			_, err := transaction.AddPolicy(userPrefixed, object, action, beginStr, endStr)
			if err != nil {
				return fmt.Errorf("assign_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, object, action, beginStr, endStr, err)
			}
		}
		return nil
	})
}

// RemovePoliciesFromUser removes object-action-time policies from a user.
func RemovePoliciesFromUser(context *gin.Context, userCode string, policies []Policy) error {
	if enforcer == nil {
		return ErrEnforcerNotInit
	}

	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return ErrInvalidUserCode
	}

	userPrefixed := addPrefix(userCode, UserCode)
	existingPolicies, err := enforcer.GetFilteredPolicy(0, userPrefixed)
	if err != nil {
		return fmt.Errorf("query_policies subject=%s: %w", userPrefixed, err)
	}
	existSet := make(map[string]struct{})
	for _, policyEntry := range existingPolicies {
		if len(policyEntry) >= 5 {
			key := policyEntry[1] + "|" + policyEntry[2] + "|" + policyEntry[3] + "|" + policyEntry[4]
			existSet[key] = struct{}{}
		}
	}
	for i := range policies {
		policies[i].Object = strings.TrimSpace(policies[i].Object)
		policies[i].Action = strings.TrimSpace(policies[i].Action)
		policyItem := policies[i]
		if err := policyItem.Validate(); err != nil {
			return fmt.Errorf("remove_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, policyItem.Object, policyItem.Action, formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime), err)
		}
		object, action := policyItem.Object, policyItem.Action
		beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
		key := object + "|" + action + "|" + beginStr + "|" + endStr
		if _, ok := existSet[key]; !ok {
			return fmt.Errorf("remove_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, object, action, beginStr, endStr, ErrPolicyNotExist)
		}
	}

	return enforcer.WithTransaction(context, func(transaction *casbin.Transaction) error {
		for _, policyItem := range policies {
			object, action := policyItem.Object, policyItem.Action
			beginStr, endStr := formatTime(policyItem.BeginTime), formatTime(policyItem.EndTime)
			_, err := transaction.RemovePolicy(userPrefixed, object, action, beginStr, endStr)
			if err != nil {
				return fmt.Errorf("remove_policy subject=%s object=%s action=%s begin=%s end=%s: %w", userPrefixed, object, action, beginStr, endStr, err)
			}
		}
		return nil
	})
}

// GetPoliciesForUser returns effective policies for a user (direct + inherited from roles).
func GetPoliciesForUser(context *gin.Context, userCode string) ([]Policy, error) {
	if enforcer == nil {
		return nil, ErrEnforcerNotInit
	}

	if strings.TrimSpace(userCode) == "" {
		return nil, ErrInvalidUserCode
	}

	userPrefixed := addPrefix(userCode, UserCode)
	userPoliciesRaw, err := enforcer.GetFilteredPolicy(0, userPrefixed)
	if err != nil {
		return nil, fmt.Errorf("query_policies subject=%s: %w", userPrefixed, err)
	}
	var userPolicies []Policy
	for _, entry := range userPoliciesRaw {
		if len(entry) < 5 {
			continue
		}
		tb, err := time.Parse(time.RFC3339, entry[3])
		if err != nil {
			return nil, err
		}
		te, err := time.Parse(time.RFC3339, entry[4])
		if err != nil {
			return nil, err
		}
		userPolicies = append(userPolicies, Policy{Object: entry[1], Action: entry[2], BeginTime: tb, EndTime: te})
	}

	roles, err := GetRolesForUser(context, userCode)
	if err != nil {
		return nil, fmt.Errorf("query_grouping type=%s subject=%s: %w", GTypeUserRole, userPrefixed, err)
	}
	var rolePolicies []Policy
	for _, role := range roles {
		rolePolicyList, err := GetPoliciesForRole(context, role)
		if err != nil {
			return nil, fmt.Errorf("query_policies subject=%s: %w", addPrefix(role, RoleCode), err)
		}
		rolePolicies = append(rolePolicies, rolePolicyList...)
	}
	combined := append(userPolicies, rolePolicies...)
	sort.Slice(combined, func(i, j int) bool {
		pi, pj := combined[i], combined[j]
		if pi.Object != pj.Object {
			return pi.Object < pj.Object
		}
		if pi.Action != pj.Action {
			return pi.Action < pj.Action
		}
		if !pi.BeginTime.Equal(pj.BeginTime) {
			return pi.BeginTime.Before(pj.BeginTime)
		}
		return pi.EndTime.Before(pj.EndTime)
	})
	return combined, nil
}

// Policy describes an object-action rule within a time window.
type Policy struct {
	Object    string
	Action    string
	BeginTime time.Time
	EndTime   time.Time
}

// Validate checks the policy fields and time window for correctness.
func (p Policy) Validate() error {
	if p.Object == "" || p.Action == "" {
		return ErrInvalidPolicyFields
	}
	if !p.EndTime.After(p.BeginTime) {
		return ErrInvalidTimeRange
	}
	return nil
}
