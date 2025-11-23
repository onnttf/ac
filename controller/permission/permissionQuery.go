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

type permissionQueryInput struct {
	Page        int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	SubjectCode string `form:"subject_code" binding:"omitempty,len=36"`
	ObjectCode  string `form:"object_code" binding:"omitempty"`
	Action      string `form:"action" binding:"omitempty"`
}

type permissionQueryOutput struct {
	Id          int64     `json:"id"`
	SubjectCode string    `json:"subject_code"`
	ObjectCode  string    `json:"object_code"`
	Action      string    `json:"action"`
	BeginTime   time.Time `json:"begin_time"`
	EndTime     time.Time `json:"end_time"`
}

// @Summary Query policies by fields
// @Tags permission
// @Param input query permissionQueryInput false "input"
// @Success 200 {object} controller.Response{data=permissionQueryOutput} "output"
// @Router /permission/query [get]
func permissionQuery(ctx *gin.Context) {
	var input permissionQueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	ruleRepo := dal.NewRepo[model.TblCasbinRule]()

	rule, err := ruleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		db = db.Where("ptype = ?", "p")
		if input.SubjectCode != "" {
			db = db.Where("v0 = ?", input.SubjectCode)
		}
		if input.ObjectCode != "" {
			db = db.Where("v1 = ?", input.ObjectCode)
		}
		if input.Action != "" {
			db = db.Where("v2 = ?", input.Action)
		}
		return db
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

	controller.Success(ctx, permissionQueryOutput{
		Id:          rule.Id,
		SubjectCode: rule.V0,
		ObjectCode:  rule.V1,
		Action:      rule.V2,
		BeginTime:   beginTime,
		EndTime:     endTime,
	})
}
