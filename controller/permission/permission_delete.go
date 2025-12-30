package permission

import (
	"strings"
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type permissionDeleteInput struct {
	Id int64 `json:"id" binding:"required,min=1"`
}

type permissionDeleteOutput struct{}

// @Summary Delete an existing permission
// @Tags permission
// @Param input body permissionDeleteInput true "input"
// @Success 200 {object} controller.Response{data=permissionDeleteOutput} "output"
// @Router /api/permission/delete [post]
func permissionDelete(ctx *gin.Context) {
	var input permissionDeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Get existing permission
	ruleRepo := dal.NewRepo[model.TblCasbinRule]()
	rule, err := ruleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND ptype = ?", input.Id, "p")
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if rule == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("permission not found"))
		return
	}

	// Parse time
	beginTime, err1 := time.Parse(time.RFC3339, rule.V3)
	endTime, err2 := time.Parse(time.RFC3339, rule.V4)
	if err1 != nil || err2 != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("invalid permission time format"))
		return
	}

	if strings.HasPrefix(rule.V0, casbin.PrefixUser) {
		userCode := strings.TrimPrefix(rule.V0, casbin.PrefixUser)
		if err := casbin.RemovePoliciesFromUser(ctx, userCode, []casbin.Policy{{Object: rule.V1, Action: rule.V2, BeginTime: beginTime, EndTime: endTime}}); err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
	} else if strings.HasPrefix(rule.V0, casbin.PrefixRole) {
		roleCode := strings.TrimPrefix(rule.V0, casbin.PrefixRole)
		if err := casbin.RemovePoliciesFromRole(ctx, roleCode, []casbin.Policy{{Object: rule.V1, Action: rule.V2, BeginTime: beginTime, EndTime: endTime}}); err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
	} else {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("unknown subject prefix in rule"))
		return
	}

	controller.Success(ctx, permissionDeleteOutput{})
}
