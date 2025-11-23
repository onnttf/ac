package role

import (
	"ac/controller"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/tree"
)

type roleListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type roleListOutput struct {
	Total int64            `json:"total"`
	List  []roleListRecord `json:"list"`
}

type roleListRecord struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary List roles with pagination
// @Tags role
// @Param input query roleListInput true "input"
// @Success 200 {object} controller.Response{data=roleListOutput} "output"
// @Router /role/list [get]
func roleList(ctx *gin.Context) {
	data := map[string]any{}
	nodeMap := make(map[int64]*tree.Node)
	nodeList := make([]*tree.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodeList = append(nodeList, node)
	}

	a, b := tree.NewTreeBuilder().WithNodes(nodeList).Build()
	data["a"] = a
	data["b"] = b
	controller.Success(ctx, data)
}
