package user

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"
	"ac/util"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateInput struct {
	Name     string `json:"name" binding:"required,min=6,max=50" example:"Alice"`
	Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=6,max=8" example:"123456"`
}

type CreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new user
// @Tags user
// @Param input body CreateInput true "input"
// @Response 200 {object} controller.Response{data=CreateOutput} "output"
// @Router /internal-api/user/create [post]
func internalApiUserCreate(ctx *gin.Context) {
	var input CreateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "user: create: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	emailCount, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where(model.TblUser{Email: input.Email})
	})
	if err != nil {
		logger.Errorf(ctx, "user: create: failed, reason=query email, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if emailCount > 0 {
		logger.Warnf(ctx, "user: create: failed, reason=email already exists, email=%s", input.Email)
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("email already exists"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf(ctx, "user: create: failed, reason=generate password hash, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	now := time.NowUTC()
	newValue := &model.TblUser{
		Code:         util.GenerateCode(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := userRepo.Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "user: create: failed, reason=insert user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "user: create: succeeded, id=%d, code=%s, email=%s",
		newValue.Id, newValue.Code, newValue.Email)

	controller.Success(ctx, CreateOutput{
		Code: newValue.Code,
	})
}
