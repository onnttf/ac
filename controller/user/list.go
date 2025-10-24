package user

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
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
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary List users with pagination
// @Tags user
// @Param input query ListInput true "input"
// @Response 200 {object} controller.Response{data=ListOutput} "output"
// @Router /internal-api/user/list [get]
func internalApiUserList(ctx *gin.Context) {
	var input ListInput
	if err := ctx.ShouldBindQuery(&input); err != nil {
		logger.Errorf(ctx, "user: list: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	total, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		logger.Errorf(ctx, "user: list: failed, reason=count user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	userList, err := userRepo.Query(ctx, database.DB, dal.Paginate(input.Page, input.PageSize), dal.OrderBy("id", "DESC"))
	if err != nil {
		logger.Errorf(ctx, "user: list: failed, reason=query user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	list := make([]ListRecord, len(userList))
	for i, u := range userList {
		list[i] = ListRecord{
			Code:  u.Code,
			Name:  u.Name,
			Email: u.Email,
		}
	}

	controller.Success(ctx, ListOutput{Total: total, List: list})
}
