package permission

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type permissionFetchInput struct {
	Id int64 `form:"id" binding:"required,min=1"`
}

type permissionFetchOutput struct {
	Id          int64     `json:"id"`
	SubjectCode string    `json:"subject_code"`
	ObjectCode  string    `json:"object_code"`
	Action      string    `json:"action"`
	BeginTime   time.Time `json:"begin_time"`
	EndTime     time.Time `json:"end_time"`
}

// @Summary Fetch a permission by ID
// @Tags permission
// @Param input query permissionFetchInput true "input"
// @Success 200 {object} controller.Response{data=permissionFetchOutput} "output"
// @Router /api/permission/fetch [get]
func permissionFetch(ctx *gin.Context) {
	var input permissionFetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

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

	beginTime, err1 := time.Parse(time.RFC3339, rule.V3)
	endTime, err2 := time.Parse(time.RFC3339, rule.V4)
	if err1 != nil || err2 != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithHint("invalid permission time format"))
		return
	}

	controller.Success(ctx, permissionFetchOutput{
		Id:          rule.Id,
		SubjectCode: rule.V0,
		ObjectCode:  rule.V1,
		Action:      rule.V2,
		BeginTime:   beginTime,
		EndTime:     endTime,
	})
}
