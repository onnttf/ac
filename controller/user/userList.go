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
	Total int64            `json:"total"`
	List  []userListRecord `json:"list"`
}

type userListRecord struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary List users with pagination
// @Tags user
// @Param input query userListInput true "input"
// @Success 200 {object} controller.Response{data=userListOutput} "output"
// @Router /user/list [get]
func userList(ctx *gin.Context) {
	var input userListInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	total, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	userList, err := userRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]userListRecord, len(userList))
	for i, u := range userList {
		list[i] = userListRecord{
			Code:  u.Code,
			Name:  u.Name,
			Email: u.Email,
		}
	}

	controller.Success(ctx, userListOutput{Total: total, List: list})
}
