package object

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

type objectCreateInput struct {
	Name       string `json:"name" binding:"required,min=1,max=50"`
	ParentCode string `json:"parent_code" binding:"omitempty,len=36"`
}

type objectCreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new object
// @Tags object
// @Param input body objectCreateInput true "input"
// @Success 200 {object} controller.Response{data=objectCreateOutput} "output"
// @Router /api/object/create [post]
func objectCreate(ctx *gin.Context) {
	var input objectCreateInput
	if err := ctx.ShouldBind(&input); err != nil {

		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	now := time.Now()
	newValue := &model.TblObject{
		Code:       util.GenerateCode(),
		Name:       input.Name,
		Type:       model.ObjectTypeMenu,
		ParentCode: input.ParentCode,
		Sort:       1,
		Status:     model.StatusEnabled.Int64(),
		Deleted:    model.NotDeleted,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	objectRepo := dal.NewRepo[model.TblObject]()
	if input.ParentCode != "" {

		parent, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ? AND deleted = ?", input.ParentCode, model.NotDeleted)
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent object not found"))
			return
		}
		if parent == nil {
			controller.Failure(ctx, controller.ErrSystemError.WithHint("parent object not found"))
			return
		}

		lastChild, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_code = ? AND deleted = ?", input.ParentCode, model.NotDeleted).Order("sort DESC")
		})
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}
		if lastChild != nil {
			newValue.Sort = lastChild.Sort + 1
		}
	}

	if err := objectRepo.Insert(ctx, database.DB, newValue); err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, objectCreateOutput{
		Code: newValue.Code,
	})
}
