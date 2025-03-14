package resource

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/input"
	"ac/custom/output"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"ac/service/resource"
	"ac/service/system"
	"time"

	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
)

type Resource struct {
	ID                 int64     `json:"id"`
	SystemCode         string    `json:"system_code"`
	ResourceCode       string    `json:"resource_code"`
	Name               string    `json:"name"`
	ParentResourceCode string    `json:"parent_resource_code"`
	Description        string    `json:"description"`
	ModifiedBy         string    `json:"modified_by"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func RegisterRoutes(g *echo.Group) {
	g.GET("/query", queryFunc)

	g.POST("/add", addFunc)
	g.POST("/update", updateFunc)
	g.POST("/delete", deleteFunc)
}

func deleteFunc(ctx echo.Context) error {
	body := struct {
		SystemCode   string `json:"system_code" validate:"required,gt=0"`
		ResourceCode string `json:"resource_code" validate:"required,gt=0"`
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
	if ok, err := resource.Validate(ctx.Request().Context(), body.SystemCode, body.ResourceCode); !ok {
		if err != nil {
			logger.Errorf(ctx.Request().Context(), "failed to validate resource, err: %s, system code: %s, code: %s", err.Error(), body.SystemCode, body.ResourceCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
	}

	err := dal.NewRepo[model.Resource]().Delete(ctx.Request().Context(), database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.ResourceCode})
	})
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	return output.Success(ctx, nil)
}

func updateFunc(ctx echo.Context) error {
	body := struct {
		SystemCode         string `json:"system_code" validate:"required,gt=0"`
		ResourceCode       string `json:"resource_code" validate:"required,gt=0"`
		Name               string `json:"name" validate:"required,gt=0"`
		Description        string `json:"description"`
		ParentResourceCode string `json:"parent_resource_code"`
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
	if ok, err := resource.Validate(ctx.Request().Context(), body.SystemCode, body.ResourceCode); !ok {
		if err != nil {
			logger.Errorf(ctx.Request().Context(), "failed to validate resource, err: %s, system code: %s, code: %s", err.Error(), body.SystemCode, body.ResourceCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
	}
	if body.ParentResourceCode != "" {
		if ok, err := resource.Validate(ctx.Request().Context(), body.SystemCode, body.ParentResourceCode); !ok {
			if err != nil {
				logger.Errorf(ctx.Request().Context(), "failed to validate resource, err: %s, system code: %s, code: %s", err.Error(), body.SystemCode, body.ParentResourceCode)
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
		}
	}

	if err := dal.NewRepo[model.Resource]().Update(ctx.Request().Context(), database.DB, &model.Resource{
		Name:        body.Name,
		Description: &body.Description,
		ParentCode:  &body.ParentResourceCode,
		ModifiedBy:  "",
		UpdatedAt:   util.UTCNow(),
	}, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.ResourceCode}).Limit(1)
	}); err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	return output.Success(ctx, nil)
}

func addFunc(ctx echo.Context) error {
	body := struct {
		SystemCode         string `json:"system_code" validate:"required,gt=0"`
		Name               string `json:"name" validate:"required,gt=0"`
		Description        string `json:"description"`
		ParentResourceCode string `json:"parent_resource_code"`
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
	if body.ParentResourceCode != "" {
		if ok, err := resource.Validate(ctx.Request().Context(), body.SystemCode, body.ParentResourceCode); !ok {
			if err != nil {
				logger.Errorf(ctx.Request().Context(), "failed to validate resource, err: %s, system code: %s, code: %s", err.Error(), body.SystemCode, body.ParentResourceCode)
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
		}
	}

	code, err := resource.GenerateCode(ctx.Request().Context())
	if code == "" {
		if err != nil {
			logger.Warnf(ctx.Request().Context(), "failed to generate unique code: %s", err.Error())
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Unable to generate the code. Please try again later"))
	}

	now := util.UTCNow()
	newValue := &model.Resource{
		SystemCode:  body.SystemCode,
		Code:        code,
		Name:        body.Name,
		ParentCode:  &body.ParentResourceCode,
		Description: &body.Description,
		ModifiedBy:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   gorm.DeletedAt{},
	}
	if err := dal.NewRepo[model.Resource]().Insert(ctx.Request().Context(), database.DB, newValue); err != nil {
		logger.Errorf(ctx.Request().Context(), "fail to insert Index: %s", err.Error())
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, Resource{
		ID:                 newValue.ID,
		SystemCode:         newValue.SystemCode,
		ResourceCode:       newValue.Code,
		Name:               newValue.Name,
		ParentResourceCode: *newValue.ParentCode,
		Description:        *newValue.Description,
		ModifiedBy:         newValue.ModifiedBy,
		UpdatedAt:          newValue.UpdatedAt,
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

	recordList, err := dal.NewRepo[model.Resource]().QueryList(ctx.Request().Context(), database.DB, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]Resource, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, Resource{
			ID:                 v.ID,
			SystemCode:         v.SystemCode,
			ResourceCode:       v.Code,
			Name:               v.Name,
			ParentResourceCode: *v.ParentCode,
			Description:        *v.Description,
			ModifiedBy:         v.ModifiedBy,
			UpdatedAt:          v.UpdatedAt,
		})
	}

	return output.Success(ctx, map[string]any{
		"list":  list,
		"total": total,
	})
}
