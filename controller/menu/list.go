package menu

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/tree"
	"gorm.io/gorm"
)

type ListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type ListOutput struct {
	Total int64        `json:"total"`
	List  []ListRecord `json:"list"`
}

type ListRecord struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary List menus with pagination
// @Tags menu
// @Param input query ListInput true "input"
// @Response 200 {object} controller.Response{data=ListOutput} "output"
// @Router /internal-api/menu/list [get]
func internalApiMenuList(ctx *gin.Context) {
	var input ListInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: list: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	total, err := menuRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		logger.Errorf(ctx, "menu: list: failed, reason=count menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	menuList, err := menuRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	if err != nil {
		logger.Errorf(ctx, "menu: list: failed, reason=query menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]ListRecord, len(menuList))
	treeBuilder := tree.NewTreeBuilder()
	for i, v := range menuList {
		treeBuilder.AddNode(v.Code, v.ParentCode, int(v.Sort))
		list[i] = ListRecord{
			Id:   v.Id,
			Code: v.Code,
			Name: v.Name,
			Url:  v.Url,
		}
	}

	controller.Success(ctx, ListOutput{Total: total, List: list})
}
