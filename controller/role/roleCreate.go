package role

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/util"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleCreateInput struct {
	Name       string `json:"name" binding:"required,min=1,max=50"`
	ParentCode string `json:"parent_code"`
}

type roleCreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new role
// @Tags role
// @Param input body roleCreateInput true "input"
// @Success 200 {object} controller.Response{data=roleCreateOutput} "output"
// @Router /role/create [post]
func roleCreate(ctx *gin.Context) {
	var input roleCreateInput
	if err := ctx.ShouldBind(&input); err != nil {

		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	now := time.Now()
	newValue := &model.TblRole{
		Code:       util.GenerateCode(),
		Name:       input.Name,
		ParentCode: input.ParentCode,
		Sort:       1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	roleRepo := dal.NewRepo[model.TblRole]()
	if input.ParentCode != "" {

		parent, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", input.ParentCode)
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent role not found"))
			return
		}
		if parent == nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent role not found"))
			return
		}

		lastChild, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_code = ?", input.ParentCode).Order("sort DESC")
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if lastChild != nil {
			newValue.Sort = lastChild.Sort + 1
		}
	}

	if err := roleRepo.Insert(ctx, database.DB, newValue); err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleCreateOutput{
		Code: newValue.Code,
	})
}
