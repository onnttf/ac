package role

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleFetchInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type roleFetchOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Fetch a role by code
// @Tags role
// @Param input query roleFetchInput true "input"
// @Success 200 {object} controller.Response{data=roleFetchOutput} "output"
// @Router /role/fetch [get]
func roleFetch(ctx *gin.Context) {
	var input roleFetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	controller.Success(ctx, roleFetchOutput{
		Code: role.Code,
		Name: role.Name,
	})
}
