package user

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userListInput struct {
	Page     int `form:"page" binding:"required,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100" default:"10"`
}

type userListOutput struct {
	Total int64          `json:"total"`
	List  []userListItem `json:"list"`
}

type userListItem struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary List users with pagination
// @Tags user
// @Param input query userListInput true "input"
// @Success 200 {object} controller.Response{data=userListOutput} "output"
// @Router /api/user/list [get]
func userList(ctx *gin.Context) {
	var input userListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	condition := map[string]any{
		"type":    model.SubjectTypeUser,
		"deleted": model.NotDeleted,
	}

	whereScopes := []func(*gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			db.Where(condition)
			return db
		},
	}

	userRepo := dal.NewRepo[model.TblSubject]()

	total, err := userRepo.Count(ctx, database.DB, whereScopes...)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	whereScopes = append(whereScopes, dal.OrderBy("id", "DESC"), dal.Paginate(input.Page, input.PageSize))
	userList, err := userRepo.Query(ctx, database.DB, whereScopes...)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]userListItem, len(userList))
	for i, u := range userList {
		list[i] = userListItem{
			Code: u.Code,
			Name: u.Name,
		}
	}

	controller.Success(ctx, userListOutput{Total: total, List: list})
}
