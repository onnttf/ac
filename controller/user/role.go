package user

import (
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/input"
	"ac/custom/output"
	"ac/service/casbin"
	"ac/service/system"
	"ac/service/user"
	"github.com/labstack/echo/v4"
	"time"
)

type Role struct {
	System    string
	User      string
	Resource  string
	Action    string
	BeginTime time.Time
	EndTime   time.Time
}

func roleGetFunc(ctx echo.Context) error {
	body := struct {
		SystemCode string `json:"system_code" validate:"required,gt=0"`
		UserCode   string `json:"user_code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx.Request().Context(), body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx.Request().Context(), "failed to validate system, err: %s, code: %s", err.Error(), body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}
	if ok, err := user.Validate(ctx.Request().Context(), body.SystemCode, body.UserCode); !ok {
		if err != nil {
			logger.Errorf(ctx.Request().Context(), "failed to validate user, err: %s, system code: %s, code: %s", err.Error(), body.SystemCode, body.UserCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid user code"))
	}

	permissionList, err := casbin.GetSubjectPermissionList(ctx.Request().Context(), body.SystemCode, body.UserCode)
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]Permission, 0, len(permissionList))
	for _, v := range permissionList {
		list = append(list, Permission{
			System:    v.System,
			User:      v.Subject,
			Index:     v.Index,
			Action:    v.Action,
			BeginTime: v.BeginTime,
			EndTime:   v.EndTime,
		})
	}

	return output.Success(ctx, map[string]any{
		"list": list,
	})
}
