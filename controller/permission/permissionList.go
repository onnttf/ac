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

type permissionListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type permissionListOutput struct {
	Total int64                  `json:"total"`
	List  []permissionListRecord `json:"list"`
}

type permissionListRecord struct {
	Id          int64     `json:"id"`
	SubjectCode string    `json:"subject_code"`
	ObjectCode  string    `json:"object_code"`
	Action      string    `json:"action"`
	BeginTime   time.Time `json:"begin_time"`
	EndTime     time.Time `json:"end_time"`
}

// @Summary List policies with pagination
// @Tags permission
// @Param input query permissionListInput true "input"
// @Success 200 {object} controller.Response{data=permissionListOutput} "output"
// @Router /permission/list [get]
func permissionList(ctx *gin.Context) {
	var input permissionListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	ruleRepo := dal.NewRepo[model.TblCasbinRule]()

	total, err := ruleRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("ptype = ?", "p")
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	ruleList, err := ruleRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("ptype = ?", "p").Order("id DESC")
	}, dal.Paginate(input.Page, input.PageSize))
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]permissionListRecord, 0, len(ruleList))
	for _, rule := range ruleList {
		beginTime, err1 := time.Parse(time.RFC3339, rule.V3)
		endTime, err2 := time.Parse(time.RFC3339, rule.V4)
		if err1 != nil || err2 != nil {
			continue
		}
		list = append(list, permissionListRecord{
			Id:          rule.Id,
			SubjectCode: rule.V0,
			ObjectCode:  rule.V1,
			Action:      rule.V2,
			BeginTime:   beginTime,
			EndTime:     endTime,
		})
	}

	controller.Success(ctx, permissionListOutput{Total: total, List: list})
}
