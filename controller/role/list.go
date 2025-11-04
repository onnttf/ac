package role

import (
	"ac/controller"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/tree"
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

// @Summary List roles with pagination
// @Tags role
// @Param input query ListInput true "input"
// @Response 200 {object} controller.Response{data=ListOutput} "output"
// @Router /internal-api/role/list [get]
func internalApiRoleList(ctx *gin.Context) {
	// repo := dal.NewRepo[model.TblRole]()
	data := map[string]any{}
	// 1. 查询所有与树相关的数据，只进行一次数据库查询
	// rows, _ := repo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
	// 	return db.Where("tree_id = 2")
	// })

	// 2. 使用 map 来聚合节点信息，确保每个节点只被处理一次
	nodeMap := make(map[int64]*tree.Node)

	// 3. 将处理后的 map 转换为切片，传递给 TreeBuilder
	nodeList := make([]*tree.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodeList = append(nodeList, node)
	}

	// 4. 构建树
	a, b := tree.NewTreeBuilder().WithNodes(nodeList).Build()
	data["a"] = a
	data["b"] = b
	controller.Success(ctx, data)
	// var input ListInput
	// if err := ctx.ShouldBind(&input); err != nil {
	// 	logger.Errorf(ctx, "role: list: failed, reason=invalid input, error=%v", err)
	// 	controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
	// 	return
	// }

	// roleRepo := dal.NewRepo[model.TblRole]()

	// total, err := roleRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
	// 	return db
	// })
	// if err != nil {
	// 	logger.Errorf(ctx, "role: list: failed, reason=count role, error=%v", err)
	// 	controller.Failure(ctx, controller.ErrSystemError.WithError(err))
	// 	return
	// }

	// roleList, err := roleRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	// if err != nil {
	// 	logger.Errorf(ctx, "role: list: failed, reason=query role, error=%v", err)
	// 	controller.Failure(ctx, controller.ErrSystemError.WithError(err))
	// 	return
	// }

	// list := make([]ListRecord, len(roleList))
	// for i, u := range roleList {
	// 	list[i] = ListRecord{
	// 		Id:   u.Id,
	// 		Code: u.Code,
	// 		Name: u.Name,
	// 		Url:  u.Url,
	// 	}
	// }

	// controller.Success(ctx, ListOutput{Total: total, List: list})
}
