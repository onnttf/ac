package role

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type roleListOutput struct {
	Total int64          `json:"total"`
	List  []roleListItem `json:"list"`
}

type roleListItem struct {
	Id   int64             `json:"id"`
	Code string            `json:"code"`
	Name string            `json:"name"`
	Type model.SubjectType `json:"type"`
}

// @Summary List roles with pagination
// @Tags role
// @Param input query roleListInput true "input"
// @Success 200 {object} controller.Response{data=roleListOutput} "output"
// @Router /api/role/list [get]
func roleList(ctx *gin.Context) {
	var input roleListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblSubject]()

	total, err := roleRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND deleted = ?", model.SubjectTypeRole, model.NotDeleted)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	roleList, err := roleRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND deleted = ?", model.SubjectTypeRole, model.NotDeleted).Order("id DESC")
	}, dal.Paginate(input.Page, input.PageSize))
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]roleListItem, len(roleList))
	for i, v := range roleList {
		list[i] = roleListItem{
			Id:   v.Id,
			Code: v.Code,
			Name: v.Name,
			Type: v.Type,
		}
	}

	controller.Success(ctx, roleListOutput{Total: total, List: list})
}
