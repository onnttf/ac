package role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/service/role"
	"ac/service/system"
	"gorm.io/gorm"

	"ac/custom/input"
	"ac/custom/output"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"time"

	"github.com/labstack/echo/v4"
)

type Role struct {
	ID          int64     `json:"ID"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SystemCode  string    `json:"system_code"`
	Code        string    `json:"code"`
	ModifiedBy  string    `json:"modified_by"`
	UpdatedAt   time.Time `json:"update_at"`
}

func RegisterRoutes(g *echo.Group) {
	g.GET("/query", queryFunc)
	g.POST("/add", addFunc)
	g.POST("/delete", deleteFunc)
}

func deleteFunc(ctx echo.Context) error {
	body := struct {
		SystemCode string `json:"system_code" validate:"required,gt=0"`
		Code       string `json:"code" validate:"required,gt=0"`
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

	err := dal.NewRepo[model.Resource]().Delete(ctx.Request().Context(), database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.Code})
	})
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	return output.Success(ctx, nil)
}

func addFunc(ctx echo.Context) error {
	body := struct {
		SystemCode  string `json:"system_code" validate:"required,gt=0"`
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
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

	code, err := role.GenerateCode(ctx.Request().Context())
	if code == "" {
		if err != nil {
			logger.Warnf(ctx.Request().Context(), "failed to generate unique code: %s", err.Error())
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Unable to generate the code. Please try again later"))
	}

	now := util.UTCNow()
	newValue := &model.Role{
		SystemCode:  body.SystemCode,
		Name:        body.Name,
		Code:        code,
		Description: body.Description,
		ModifiedBy:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   gorm.DeletedAt{},
	}
	if err := dal.NewRepo[model.Role]().Insert(ctx.Request().Context(), database.DB, newValue); err != nil {
		logger.Errorf(ctx.Request().Context(), "fail to insert role: %s", err.Error())
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, Role{
		ID:          newValue.ID,
		Name:        newValue.Name,
		Description: newValue.Description,
		SystemCode:  newValue.SystemCode,
		Code:        newValue.Code,
		ModifiedBy:  newValue.ModifiedBy,
		UpdatedAt:   newValue.UpdatedAt,
	})
}

func queryFunc(ctx echo.Context) error {
	body := struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	total, err := dal.NewRepo[model.System]().Count(ctx.Request().Context(), database.DB)
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	recordList, err := dal.NewRepo[model.Role]().QueryList(ctx.Request().Context(), database.DB, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]Role, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, Role{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			SystemCode:  v.SystemCode,
			Code:        v.Code,
			ModifiedBy:  v.ModifiedBy,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return output.Success(ctx, map[string]any{
		"list":  list,
		"total": total,
	})
}
