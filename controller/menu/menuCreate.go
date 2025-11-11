package menu

import (
	"fmt"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/util"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type menuCreateInput struct {
	Name       string `json:"name" binding:"required,min=1,max=50"`
	Url        string `json:"url" binding:"required,min=1,max=200"`
	ParentCode string `json:"parent_code"`
}

type menuCreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new menu
// @Tags menu
// @Param input body menuCreateInput true "input"
// @Response 200 {object} controller.Response{data=menuCreateOutput} "output"
// @Router /menu/create [post]
func menuCreate(ctx *gin.Context) {
	var input menuCreateInput
	if err := ctx.ShouldBind(&input); err != nil {

		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	if err := validateUrl(ctx, input.Url); err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	now := time.NowUTC()
	newValue := &model.TblMenu{
		Code:       util.GenerateCode(),
		Name:       input.Name,
		ParentCode: input.ParentCode,
		Sort:       1,
		Url:        input.Url,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	menuRepo := dal.NewRepo[model.TblMenu]()
	if input.ParentCode != "" {

		parent, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", input.ParentCode)
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent menu not found"))
			return
		}
		if parent == nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent menu not found"))
			return
		}

		lastChild, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
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

	if err := menuRepo.Insert(ctx, database.DB, newValue); err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, menuCreateOutput{
		Code: newValue.Code,
	})
}

func validateUrl(ctx *gin.Context, url string) error {
	count, err := dal.NewRepo[model.TblMenu]().Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where(model.TblMenu{Url: url})
	})
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("url '%s' already exists", url)
	}

	return nil
}
