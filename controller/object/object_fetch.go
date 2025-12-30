package object

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type objectFetchInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type objectFetchOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// @Summary Fetch a object by code
// @Tags object
// @Param input query objectFetchInput true "input"
// @Success 200 {object} controller.Response{data=objectFetchOutput} "output"
// @Router /api/object/fetch [get]
func objectFetch(ctx *gin.Context) {
	var input objectFetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	objectRepo := dal.NewRepo[model.TblObject]()

	object, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if object == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithMsg("object not found"))
		return
	}

	controller.Success(ctx, objectFetchOutput{
		Code: object.Code,
		Name: object.Name,
	})
}
