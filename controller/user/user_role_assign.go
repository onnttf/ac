package user

import (
	"strings"

	"ac/controller"
	"ac/service/casbin"
	"ac/service/role"
	"ac/service/user"

	"github.com/gin-gonic/gin"
)

type userRoleAssignInput struct {
	UserCode  string   `json:"user_code" binding:"required,len=36"`
	RoleCodes []string `json:"role_codes" binding:"required,min=1,dive,len=36"`
}

type userRoleAssignOutput struct{}

// @Summary Assign roles to a user
// @Tags user
// @Param input body userRoleAssignInput true "input"
// @Success 200 {object} controller.Response{data=userRoleAssignOutput} "output"
// @Router /api/user/role/assign [post]
func userRoleAssign(ctx *gin.Context) {
	var input userRoleAssignInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	if err := user.Verify(ctx, input.UserCode); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	code2Error, err := role.BatchVerify(ctx, input.RoleCodes)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if len(code2Error) > 0 {
		var invalidRoles []string
		for code := range code2Error {
			invalidRoles = append(invalidRoles, code)
		}
		controller.Failure(ctx, controller.ErrSystemError.WithMsg("invalid roles: "+strings.Join(invalidRoles, ", ")))
		return
	}

	assignedRoles, err := casbin.GetRolesForUser(ctx, input.UserCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	assignedSet := make(map[string]struct{}, len(assignedRoles))
	for _, role := range assignedRoles {
		assignedSet[role] = struct{}{}
	}

	var alreadyAssigned []string
	for _, roleCode := range input.RoleCodes {
		if _, exists := assignedSet[roleCode]; exists {
			alreadyAssigned = append(alreadyAssigned, roleCode)
		}
	}

	if len(alreadyAssigned) > 0 {
		msg := "roles already assigned: " + strings.Join(alreadyAssigned, ", ")
		controller.Failure(ctx, controller.ErrSystemError.WithMsg(msg))
		return
	}

	if err := casbin.AssignRolesToUser(ctx, input.UserCode, input.RoleCodes); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userRoleAssignOutput{})
}
