package role

import (
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
	ParentCode string `json:"parent_code"`
}

type CreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new role
// @Tags role
// @Param input body CreateInput true "input"
// @Response 200 {object} controller.Response{data=CreateOutput} "output"
// @Router /internal-api/role/create [post]
func internalApiRoleCreate(ctx *gin.Context) {
	var input CreateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "role: create: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	now := time.NowUTC()
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
			logger.Errorf(ctx, "role: create: failed, reason=query role, error=%v, parent_code=%s", err, input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent role not found"))
			return
		}
		if parent == nil {
			logger.Warnf(ctx, "role: create: failed, reason=parent role not found, parent_code=%s", input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent role not found"))
			return
		}
		lastChild, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_code = ?", input.ParentCode).Order("sort DESC")
		})
		if err != nil {
			logger.Errorf(ctx, "role: create: failed, reason=query last child role, error=%v, parent_code=%s", err, input.ParentCode)
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if lastChild != nil {
			newValue.Sort = lastChild.Sort + 1
		}
	}

	if err := roleRepo.Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "role: create: failed, reason=insert role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "role: create: succeeded, id=%d, code=%s, name=%s",
		newValue.Id, newValue.Code, newValue.Name)

	controller.Success(ctx, CreateOutput{
		Code: newValue.Code,
	})
}
