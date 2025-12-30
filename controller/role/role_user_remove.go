package role

import (
	"slices"
	"strings"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleUserRemoveInput struct {
	RoleCode  string   `json:"role_code" binding:"required,len=36"`
	UserCodes []string `json:"user_codes" binding:"required,min=1,dive,len=36"`
}

type roleUserRemoveOutput struct{}

// @Summary Remove users from a role
// @Tags role
// @Param input body roleUserRemoveInput true "input"
// @Success 200 {object} controller.Response{data=roleUserRemoveOutput} "output"
// @Router /api/role/user/remove [post]
func roleUserRemove(ctx *gin.Context) {
	var input roleUserRemoveInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Step 1: Validate RoleCode existence
	roleRepo := dal.NewRepo[model.TblSubject]()
	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND type = ? AND deleted = ?", input.RoleCode, model.SubjectTypeRole, model.NotDeleted)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role not found"))
		return
	}

	// Step 2: Validate that all users are valid (not deleted)
	userRepo := dal.NewRepo[model.TblSubject]()
	users, err := userRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN ? AND type = ? AND deleted = ?", input.UserCodes, model.SubjectTypeUser, model.NotDeleted)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Step 3: Create a map of the valid users
	validUsersMap := make(map[string]struct{})
	for _, user := range users {
		validUsersMap[user.Code] = struct{}{}
	}

	// Step 4: Check which users are invalid
	var invalidUsers []string
	for _, userCode := range input.UserCodes {
		if _, exists := validUsersMap[userCode]; !exists {
			invalidUsers = append(invalidUsers, userCode)
		}
	}

	if len(invalidUsers) > 0 {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("users not found or deleted: "+strings.Join(invalidUsers, ",")))
		return
	}

	assignedUsers, err := casbin.GetUsersForRole(ctx, input.RoleCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Step 6: Check which users are not assigned the role
	var notAssignedUsers []string
	for _, userCode := range input.UserCodes {
		if !slices.Contains(assignedUsers, userCode) {
			notAssignedUsers = append(notAssignedUsers, userCode)
		}
	}

	// If there are users who don't have the role, return an error
	if len(notAssignedUsers) > 0 {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("users not assigned the role: "+strings.Join(notAssignedUsers, ",")))
		return
	}

	if err := casbin.RemoveUsersFromRole(ctx, input.RoleCode, input.UserCodes); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleUserRemoveOutput{})
}
