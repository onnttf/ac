package permission

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type permissionUpdateInput struct {
	Id        int64     `json:"id" binding:"required,min=1"`   // permission ID
	Action    string    `json:"action" binding:"required"`     // new operation type
	BeginTime time.Time `json:"begin_time" binding:"required"` // new start time
	EndTime   time.Time `json:"end_time" binding:"required"`   // new end time
}

type permissionUpdateOutput struct{}

// @Summary Update an existing permission (action or time range)
// @Tags permission
// @Param input body permissionUpdateInput true "input"
// @Response 200 {object} controller.Response{data=permissionUpdateOutput} "output"
// @Router /permission/update [post]
func permissionUpdate(ctx *gin.Context) {
	var input permissionUpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Get the current time for validation
	now := time.Now()

	// Ensure times are in the future
	if input.BeginTime.Before(now) || input.EndTime.Before(now) {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("Begin time and end time must be in the future"))
		return
	}

	// Validate time range
	if input.EndTime.Before(input.BeginTime) {
		controller.Failure(ctx, ErrInvalidTimeRange)
		return
	}

	// Get the existing permission
	ruleRepo := dal.NewRepo[model.TblCasbinRule]()
	rule, err := ruleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND ptype = ?", input.Id, "p")
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if rule == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("Permission not found"))
		return
	}

	// Safety check
	if rule.V0 == "" || rule.V1 == "" {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("Invalid Casbin rule data"))
		return
	}

	// Parse old time values
	oldBeginTime, err1 := time.Parse(time.RFC3339, rule.V3)
	oldEndTime, err2 := time.Parse(time.RFC3339, rule.V4)
	if err1 != nil || err2 != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("Invalid time format in rule"))
		return
	}

	// Remove the old policy
	err = casbin.RemovePolicy(ctx, rule.V0, rule.V1, rule.V2, oldBeginTime, oldEndTime)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("Failed to remove old policy"))
		return
	}

	// Add the new policy
	err = casbin.AddPolicy(ctx, rule.V0, rule.V1, input.Action, input.BeginTime, input.EndTime)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err).WithHint("Failed to add new policy"))
		return
	}

	controller.Success(ctx, permissionUpdateOutput{})
}
