package user

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/input"
	"ac/custom/output"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"ac/service/casbin"
	"ac/service/system"
	"ac/service/user"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Permission struct {
	System        string `json:"system"`
	User          string `json:"user"`
	Index         string `json:"index"`
	indexPartList []string
	IndexName     string    `json:"index_name"`
	Action        string    `json:"action"`
	BeginTime     time.Time `json:"begin_time"`
	EndTime       time.Time `json:"end_time"`
}

func permissionGetFunc(ctx echo.Context) error {
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
		logger.Errorf(ctx.Request().Context(), "failed to get permission list, err: %s", err.Error())
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]Permission, 0, len(permissionList))
	tmpResourceCodeList := make([]string, 0, len(permissionList))
	for _, v := range permissionList {
		indexPartList := strings.Split(v.Index, "/")
		permission := Permission{
			System:        v.System,
			User:          v.Subject,
			Index:         v.Index,
			indexPartList: indexPartList,
			Action:        v.Action,
			BeginTime:     v.BeginTime,
			EndTime:       v.EndTime,
		}
		list = append(list, permission)
		tmpResourceCodeList = append(tmpResourceCodeList, indexPartList...)
	}
	resourceCodeList := make([]string, 0, len(tmpResourceCodeList))
	for _, v := range tmpResourceCodeList {
		if v == "" || v == "*" {
			continue
		}
		resourceCodeList = append(resourceCodeList, v)
	}
	resourceRecordList, err := dal.NewRepo[model.Resource]().QueryList(ctx.Request().Context(), database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("system_code = ? and code in ?", body.SystemCode, resourceCodeList)
	})
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}
	resourceCode2Resource := util.ToMap(resourceRecordList, func(t model.Resource) string {
		return t.Code
	})

	for i, permission := range list {
		nameList := make([]string, 0, len(permission.indexPartList))
		for _, part := range permission.indexPartList {
			if part == "*" {
				nameList = append(nameList, "全部")
			} else if resource, ok := resourceCode2Resource[part]; ok {
				nameList = append(nameList, resource.Name)
			} else {
				nameList = append(nameList, "未知")
			}
		}
		list[i].IndexName = strings.Join(nameList, "/")
	}

	return output.Success(ctx, map[string]any{
		"list": list,
	})
}
