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
	UserCode   string    `json:"user_code" binding:"omitempty,len=36"` // user unique code
	RoleCode   string    `json:"role_code" binding:"omitempty,len=36"` // role unique code
	ObjectCode string    `json:"object_code" binding:"required"`       // resource code
	Action     string    `json:"action" binding:"required"`            // operation type
	BeginTime  time.Time `json:"begin_time" binding:"required"`        // start time
	EndTime    time.Time `json:"end_time" binding:"required"`          // end time
}

type permissionCreateOutput struct {
	Id int64 `json:"id"`
}

var (
	ErrInvalidTimeRange = util.NewError(1004, "invalid time range", "end_time must be after begin_time")
	ErrMenuNotFound     = util.NewError(1003, "menu not found", "the specified menu does not exist")
)

// @Summary Create a new permission
// @Tags permission
// @Param input body permissionCreateInput true "input"
// @Response 200 {object} controller.Response{data=permissionCreateOutput} "output"
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
		userRepo := dal.NewRepo[model.TblUser]()
		user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ? AND deleted = 0", input.UserCode)
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
		roleRepo := dal.NewRepo[model.TblRole]()
		role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ? AND deleted = 0", input.RoleCode)
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

	menuRepo := dal.NewRepo[model.TblMenu]()
	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND deleted = 0", input.ObjectCode)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		controller.Failure(ctx, ErrMenuNotFound)
		return
	}

	// add casbin policy with the updated subjectCode and objectCode
	err = casbin.AddPolicy(ctx, subjectCode, objectCode, input.Action, input.BeginTime, input.EndTime)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// retrieve the created policy record
	ruleRepo := dal.NewRepo[model.TblCasbinRule]()
	rule, err := ruleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ?",
			"p", subjectCode, objectCode, input.Action,
			input.BeginTime.Format(time.RFC3339), input.EndTime.Format(time.RFC3339))
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, permissionCreateOutput{Id: rule.Id})
}
