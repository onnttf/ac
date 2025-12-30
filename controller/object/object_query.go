package object

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type objectQueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Name     string `form:"name" binding:"omitempty,min=1"`
}

type objectQueryOutput struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Query objects by fields
// @Tags object
// @Param input query objectQueryInput false "input"
// @Success 200 {object} controller.Response{data=objectQueryOutput} "output"
// @Router /api/object/query [get]
func objectQuery(ctx *gin.Context) {
	var input objectQueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	objectRepo := dal.NewRepo[model.TblObject]()

	object, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Name != "" {
			db = db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if object == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithMsg("object not found"))
		return
	}

	controller.Success(ctx, objectQueryOutput{
		Id:   object.Id,
		Code: object.Code,
		Name: object.Name,
	})
}
