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

type roleUserRemoveOutput struct {
	RemovedCount int `json:"removed_count"`
}

// @Summary Remove users from a role
// @Tags role
// @Param input body roleUserRemoveInput true "input"
// @Response 200 {object} controller.Response{data=roleUserRemoveOutput} "output"
// @Router /role/user/remove [post]
func roleUserRemove(ctx *gin.Context) {
	var input roleUserRemoveInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Step 1: Validate RoleCode existence
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

	// Step 2: Validate that all users are valid (not deleted)
	userRepo := dal.NewRepo[model.TblUser]()
	users, err := userRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN ? AND deleted = 0", input.UserCodes)
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

	// Step 5: Get all users assigned to this role (no need to validate the role existence again)
	assignedUsers, err := casbin.GetRoleUsers(ctx, input.RoleCode)
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

	// Step 7: Remove the users from the role
	removedCount, err := casbin.RemoveUsersFromRole(ctx, input.RoleCode, input.UserCodes)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Return the count of removed users
	controller.Success(ctx, roleUserRemoveOutput{
		RemovedCount: removedCount,
	})
}
