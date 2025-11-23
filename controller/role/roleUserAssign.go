package role

import (
	"slices"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleUserAssignInput struct {
	RoleCode  string   `json:"role_code" binding:"required,len=36"`
	UserCodes []string `json:"user_codes" binding:"required,min=1,dive,len=36"`
}

type roleUserAssignOutput struct{}

// @Summary Assign users to a role
// @Tags role
// @Param input body roleUserAssignInput true "input"
// @Success 200 {object} controller.Response{data=roleUserAssignOutput} "output"
// @Router /role/user/assign [post]
func roleUserAssign(ctx *gin.Context) {
	var input roleUserAssignInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Validate RoleCode existence
	roleRepo := dal.NewRepo[model.TblRole]()
	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND deleted = 0", input.RoleCode)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role not found"))
		return
	}

	assignedUsers, err := casbin.GetUsersForRole(ctx, input.RoleCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Ensure no user is already assigned
	for _, userCode := range input.UserCodes {
		if slices.Contains(assignedUsers, userCode) {
			controller.Failure(ctx, controller.ErrInvalidInput.WithHint("user already assigned to the role: "+userCode))
			return
		}
	}

	// Validate that each user exists and is not deleted using batch query
	userRepo := dal.NewRepo[model.TblUser]()
	users, err := userRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN ? AND deleted = 0", input.UserCodes)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Check if all users are found
	validUsers := make(map[string]struct{})
	for _, user := range users {
		validUsers[user.Code] = struct{}{}
	}

	for _, userCode := range input.UserCodes {
		if _, exists := validUsers[userCode]; !exists {
			controller.Failure(ctx, controller.ErrInvalidInput.WithHint("user not found: "+userCode))
			return
		}
	}

	if err := casbin.AssignUsersToRole(ctx, input.RoleCode, input.UserCodes); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleUserAssignOutput{})
}
