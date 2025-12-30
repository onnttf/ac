package object

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/tree"
	"gorm.io/gorm"
)

type objectListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type objectListOutput struct {
	Total int64            `json:"total"`
	List  []objectListItem `json:"list"`
}

type objectListItem struct {
	Id   int64            `json:"id"`
	Code string           `json:"code"`
	Name string           `json:"name"`
	Type model.ObjectType `json:"type"`
}

// @Summary List objects with pagination
// @Tags object
// @Param input query objectListInput true "input"
// @Success 200 {object} controller.Response{data=objectListOutput} "output"
// @Router /api/object/list [get]
func objectList(ctx *gin.Context) {
	var input objectListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	objectRepo := dal.NewRepo[model.TblObject]()

	total, err := objectRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	objectList, err := objectRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]objectListItem, len(objectList))
	treeBuilder := tree.NewTreeBuilder()
	for i, v := range objectList {
		treeBuilder.AddNode(v.Code, v.ParentCode, int(v.Sort))
		list[i] = objectListItem{
			Id:   v.Id,
			Code: v.Code,
			Name: v.Name,
		}
	}

	controller.Success(ctx, objectListOutput{Total: total, List: list})
}
