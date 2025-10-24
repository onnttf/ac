package casbin

import (
	"fmt"
	"os"
	"sync"
	"time"

	"ac/bootstrap/database"
	"ac/bootstrap/logger"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
)

var (
	Enforcer *casbin.Enforcer
	once     sync.Once
	initErr  error
)

const (
	GTypeUserRole    = "g"
	GTypeObjectGroup = "g2"
)

func Initialize() error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: started\n")

		gormadapter.TurnOffAutoMigrate(database.DB)

		adapter, err := gormadapter.NewAdapterByDB(database.DB)
		if err != nil {
			initErr = fmt.Errorf("failed to create Gorm Adapter: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create adapter failed: %v\n", err)
			return
		}

		m, err := model.NewModelFromString(`
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

		enforcer, err := casbin.NewEnforcer(m, adapter)
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin Enforcer: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: create enforcer failed: %v\n", err)
			return
		}

		if err := enforcer.LoadPolicy(); err != nil {
			initErr = fmt.Errorf("failed to load Casbin policy from DB: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: casbin: init: load policy failed: %v\n", err)
			return
		}

		Enforcer = enforcer
		fmt.Fprintf(os.Stdout, "INFO: casbin: init: succeeded, model=memory, policy_table=casbin_rule\n")
	})
	return initErr
}

func LoadPolicy(ctx *gin.Context) error {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "policy: load: failed, reason=uninitialized enforcer, error=%v", err)
		return err
	}
	if err := Enforcer.LoadPolicy(); err != nil {
		logger.Errorf(ctx, "policy: load: failed, reason=load policy, error=%v", err)
		return err
	}
	return nil
}

func Enforce(ctx *gin.Context, sub, obj, act string) (bool, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "policy: enforce: failed, reason=uninitialized enforcer, error=%v", err)
		return false, err
	}

	ok, err := Enforcer.Enforce(sub, obj, act, formatTime(time.Now().UTC()))
	if err != nil {
		logger.Errorf(ctx, "policy: enforce: failed, reason=enforce policy, error=%v", err)
	}
	return ok, err
}

func AddPolicy(ctx *gin.Context, sub, obj, act string, beginTime, endTime time.Time) (bool, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "policy: add: failed, reason=uninitialized enforcer, error=%v", err)
		return false, err
	}

	added, err := Enforcer.AddPolicy(sub, obj, act, formatTime(beginTime), formatTime(endTime))
	if err != nil {
		logger.Errorf(ctx, "policy: add: failed, reason=add policy, error=%v", err)
		return false, err
	}

	if added {
		if err := Enforcer.SavePolicy(); err != nil {
			_, rollbackErr := Enforcer.RemovePolicy(sub, obj, act, formatTime(beginTime), formatTime(endTime))
			if rollbackErr != nil {
				logger.Errorf(ctx, "policy: add: failed to save policy: %v; rollback also failed: %v", err, rollbackErr)
				return false, fmt.Errorf("failed to save policy: %v; rollback also failed: %v", err, rollbackErr)
			}
			logger.Errorf(ctx, "policy: add: failed to save policy: %v", err)
			return false, fmt.Errorf("failed to save policy: %w", err)
		}
	}

	return added, nil
}

func RemovePolicy(ctx *gin.Context, sub, obj, act string, beginTime, endTime time.Time) (bool, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "policy: remove: failed, reason=uninitialized enforcer, error=%v", err)
		return false, err
	}

	removed, err := Enforcer.RemovePolicy(sub, obj, act, formatTime(beginTime), formatTime(endTime))
	if err != nil {
		logger.Errorf(ctx, "policy: remove: failed, reason=remove policy, error=%v", err)
		return false, err
	}

	if removed {
		if err := Enforcer.SavePolicy(); err != nil {
			_, rollbackErr := Enforcer.AddPolicy(sub, obj, act, formatTime(beginTime), formatTime(endTime))
			if rollbackErr != nil {
				logger.Errorf(ctx, "policy: remove: failed to save policy removal: %v; rollback also failed: %v", err, rollbackErr)
				return false, fmt.Errorf("failed to save policy removal: %v; rollback also failed: %v", err, rollbackErr)
			}
			logger.Errorf(ctx, "policy: remove: failed to save policy removal: %v", err)
			return false, fmt.Errorf("failed to save policy removal: %w", err)
		}
	}

	return removed, nil
}

func AddSubjectToGroup(ctx *gin.Context, gType, subject, group string) (bool, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "group: add subject: failed, reason=uninitialized enforcer, error=%v", err)
		return false, err
	}

	added, err := Enforcer.AddNamedGroupingPolicy(gType, subject, group)
	if err != nil {
		logger.Errorf(ctx, "group: add subject: failed, reason=add grouping policy, error=%v", err)
		return false, fmt.Errorf("failed to add subject %s to group %s in policy %s: %w", subject, group, gType, err)
	}

	if added {
		if err := Enforcer.SavePolicy(); err != nil {
			_, rollbackErr := Enforcer.RemoveNamedGroupingPolicy(gType, subject, group)
			if rollbackErr != nil {
				logger.Errorf(ctx, "group: add subject: failed to save grouping policy: %v; rollback also failed: %v", err, rollbackErr)
				return false, fmt.Errorf("failed to save grouping policy: %v; rollback also failed: %v", err, rollbackErr)
			}
			logger.Errorf(ctx, "group: add subject: failed to save grouping policy: %v", err)
			return false, fmt.Errorf("failed to save grouping policy: %w", err)
		}
	}

	return added, nil
}

func GetGroupsForSubject(ctx *gin.Context, gType, subject string) ([]string, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "group: get groups for subject: failed, reason=uninitialized enforcer, error=%v", err)
		return nil, err
	}

	policy, err := Enforcer.GetFilteredNamedGroupingPolicy(gType, 0, subject)
	if err != nil {
		logger.Errorf(ctx, "group: get groups for subject: failed, reason=query grouping policy, error=%v", err)
		return nil, fmt.Errorf("failed to query grouping policy %s: %w", gType, err)
	}

	groupSet := make(map[string]struct{})
	for _, rule := range policy {
		if len(rule) >= 2 {
			groupSet[rule[1]] = struct{}{}
		}
	}

	groupList := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groupList = append(groupList, g)
	}

	return groupList, nil
}

func GetSubjectsForGroup(ctx *gin.Context, gType, group string) ([]string, error) {
	if Enforcer == nil {
		err := fmt.Errorf("casbin enforcer not initialized")
		logger.Errorf(ctx, "group: get subjects for group: failed, reason=uninitialized enforcer, error=%v", err)
		return nil, err
	}

	policy, err := Enforcer.GetFilteredNamedGroupingPolicy(gType, 1, group)
	if err != nil {
		logger.Errorf(ctx, "group: get subjects for group: failed, reason=query grouping policy, error=%v", err)
		return nil, fmt.Errorf("failed to query grouping policy %s: %w", gType, err)
	}

	subjectSet := make(map[string]struct{})
	for _, rule := range policy {
		if len(rule) >= 2 {
			subjectSet[rule[0]] = struct{}{}
		}
	}

	subjectList := make([]string, 0, len(subjectSet))
	for s := range subjectSet {
		subjectList = append(subjectList, s)
	}

	return subjectList, nil
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
