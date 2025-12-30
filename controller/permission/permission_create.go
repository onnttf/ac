package permission

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"
	"ac/util"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

type permissionCreateInput struct {
	UserCode   string    `json:"user_code" binding:"omitempty,len=36"`
	RoleCode   string    `json:"role_code" binding:"omitempty,len=36"`
	ObjectCode string    `json:"object_code" binding:"required,len=36"`
	Action     string    `json:"action" binding:"required,min=1,max=50"`
	BeginTime  time.Time `json:"begin_time" binding:"required"`
	EndTime    time.Time `json:"end_time" binding:"required"`
}

type permissionCreateOutput struct {
	Id int64 `json:"id"`
}

var (
	ErrInvalidTimeRange = util.NewError(1004, "invalid time range", "end_time must be after begin_time")
	ErrObjectNotFound   = util.NewError(1003, "object not found", "the specified object does not exist")
)

// @Summary Create a new permission
// @Tags permission
// @Param input body permissionCreateInput true "input"
// @Success 200 {object} controller.Response{data=permissionCreateOutput} "output"
// @Router /permission/create [post]
func permissionCreate(ctx *gin.Context) {
	var input permissionCreateInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Check if begin time and end time are in the future
	now := time.Now()
	if input.BeginTime.Before(now) || input.EndTime.Before(now) {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("begin time and end time must be in the future"))
		return
	}

	// either user_code or role_code must be provided
	if (input.UserCode == "" && input.RoleCode == "") || (input.UserCode != "" && input.RoleCode != "") {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("must provide either user_code or role_code"))
		return
	}

	if input.EndTime.Before(input.BeginTime) {
		controller.Failure(ctx, ErrInvalidTimeRange)
		return
	}

	var subjectCode string

	// validate user existence
	if input.UserCode != "" {
		userRepo := dal.NewRepo[model.TblSubject]()
		user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ? AND type = ? AND deleted = ?", input.UserCode, model.SubjectTypeUser, model.NotDeleted)
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if user == nil {
			controller.Failure(ctx, controller.ErrInvalidInput.WithHint("user not found"))
			return
		}
		subjectCode = input.UserCode
	}

	// validate role existence
	if input.RoleCode != "" {
		roleRepo := dal.NewRepo[model.TblSubject]()
		role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ? AND type = ? AND deleted = ?", input.RoleCode, model.SubjectTypeRole, model.NotDeleted)
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if role == nil {
			controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role not found"))
			return
		}
		subjectCode = input.RoleCode
	}

	// safety check for empty subjectCode
	if subjectCode == "" {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("subject code is empty"))
		return
	}

	objectCode := input.ObjectCode

	objectRepo := dal.NewRepo[model.TblObject]()
	object, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND deleted = ?", input.ObjectCode, model.NotDeleted)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if object == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("object not found"))
		return
	}

	if input.UserCode != "" {
		if err := casbin.AssignPoliciesToUser(ctx, input.UserCode, []casbin.Policy{{Object: objectCode, Action: input.Action, BeginTime: input.BeginTime, EndTime: input.EndTime}}); err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
	} else {
		if err := casbin.AssignPoliciesToRole(ctx, input.RoleCode, []casbin.Policy{{Object: objectCode, Action: input.Action, BeginTime: input.BeginTime, EndTime: input.EndTime}}); err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
	}

	ruleRepo := dal.NewRepo[model.TblCasbinRule]()
	subjectPrefixed := subjectCode
	if input.UserCode != "" {
		subjectPrefixed = casbin.PrefixUser + subjectCode
	} else {
		subjectPrefixed = casbin.PrefixRole + subjectCode
	}
	rule, err := ruleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ?",
			"p", subjectPrefixed, objectCode, input.Action,
			input.BeginTime.Format(time.RFC3339), input.EndTime.Format(time.RFC3339))
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if rule == nil {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("created policy not found"))
		return
	}

	controller.Success(ctx, permissionCreateOutput{Id: rule.Id})
}
