package menu

import (
	"fmt"

	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/util"

	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type CreateInput struct {
	Name       string `json:"name" binding:"required,min=1,max=50"`
	Url        string `json:"url" binding:"required,min=1,max=200"`
	ParentCode string `json:"parent_code"`
}

type CreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new menu
// @Tags menu
// @Param input body CreateInput true "input"
// @Response 200 {object} controller.Response{data=CreateOutput} "output"
// @Router /internal-api/menu/create [post]
func internalApiMenuCreate(ctx *gin.Context) {
	var input CreateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: create: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	if err := validateUrl(ctx, input.Url); err != nil {
		logger.Errorf(ctx, "menu: create: failed, reason=check url, error=%v", err)
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
			logger.Errorf(ctx, "menu: create: failed, reason=query menu, error=%v, parent_code=%s", err, input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent menu not found"))
			return
		}
		if parent == nil {
			logger.Warnf(ctx, "menu: create: failed, reason=parent menu not found, parent_code=%s", input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent menu not found"))
			return
		}
		lastChild, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_code = ?", input.ParentCode).Order("sort DESC")
		})
		if err != nil {
			logger.Errorf(ctx, "menu: create: failed, reason=query last child menu, error=%v, parent_code=%s", err, input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if lastChild != nil {
			newValue.Sort = lastChild.Sort + 1
		}
	}

	if err := menuRepo.Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "menu: create: failed, reason=insert menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "menu: create: succeeded, id=%d, code=%s, name=%s",
		newValue.Id, newValue.Code, newValue.Name)

	controller.Success(ctx, CreateOutput{
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
