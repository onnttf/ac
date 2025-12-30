package user

import (
	"ac/controller"
	"ac/model"
	"ac/service/casbin"
	"ac/service/role"
	"ac/service/user"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/container"
)

type userRoleInput struct {
	UserCode string `form:"user_code" binding:"required,len=36"`
}

type userRoleOutput struct {
	List []userRoleOutputListItem `json:"list"`
}

type userRoleOutputListItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Query roles assigned to a user
// @Tags user
// @Param input query userRoleInput true "input"
// @Success 200 {object} controller.Response{data=userRoleOutput} "output"
// @Router /api/user/role [get]
func userRole(ctx *gin.Context) {
	var input userRoleInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	if err := user.Verify(ctx, input.UserCode); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	roleCodeList, err := casbin.GetRolesForUser(ctx, input.UserCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	data := userRoleOutput{
		List: make([]userRoleOutputListItem, 0),
	}

	if len(roleCodeList) > 0 {
		roleList, err := role.BatchFetch(ctx, roleCodeList)
		if err != nil {
			controller.Failure(ctx, controller.ErrSystemError.WithError(err))
			return
		}

		roleCode2Role := container.ToMap(roleList, func(r model.TblSubject) string {
			return r.Code
		})

		for _, v := range roleCodeList {
			role, ok := roleCode2Role[v]
			if !ok {
				continue
			}
			data.List = append(data.List, userRoleOutputListItem{
				Code: role.Code,
				Name: role.Name,
			})
		}
	}

	controller.Success(ctx, data)
}
