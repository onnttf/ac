package menu

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/tree"
	"gorm.io/gorm"
)

type menuListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type menuListOutput struct {
	Total int64            `json:"total"`
	List  []menuListRecord `json:"list"`
}

type menuListRecord struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary List menus with pagination
// @Tags menu
// @Param input query menuListInput true "input"
// @Response 200 {object} controller.Response{data=menuListOutput} "output"
// @Router /menu/list [get]
func menuList(ctx *gin.Context) {
	var input menuListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	total, err := menuRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	menuList, err := menuRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]menuListRecord, len(menuList))
	treeBuilder := tree.NewTreeBuilder()
	for i, v := range menuList {
		treeBuilder.AddNode(v.Code, v.ParentCode, int(v.Sort))
		list[i] = menuListRecord{
			Id:   v.Id,
			Code: v.Code,
			Name: v.Name,
			Url:  v.Url,
		}
	}

	controller.Success(ctx, menuListOutput{Total: total, List: list})
}
