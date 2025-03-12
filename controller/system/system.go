package system

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/input"
	"ac/custom/output"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"ac/service/system"
	"time"

	"github.com/labstack/echo/v4"
)

type System struct {
	ID          int64     `json:"ID"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	ModifiedBy  string    `json:"modified_by"`
	UpdatedAt   time.Time `json:"update_at"`
}

func RegisterRoutes(g *echo.Group) {
	g.GET("/query", queryFunc)
	g.POST("/add", addFunc)
}

func addFunc(ctx echo.Context) error {
	body := struct {
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	code, err := system.GenerateCode(ctx.Request().Context())
	if code == "" {
		if err != nil {
			logger.Warnf(ctx.Request().Context(), "failed to generate unique code: %s", err.Error())
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Unable to generate the code. Please try again later"))
	}

	now := util.UTCNow()
	newValue := &model.System{
		Code:        code,
		Name:        body.Name,
		Description: body.Description,
		ModifiedBy:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.System]().Insert(ctx.Request().Context(), database.DB, newValue); err != nil {
		logger.Errorf(ctx.Request().Context(), "fail to insert system: %s", err.Error())
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, System{
		ID:          newValue.ID,
		Name:        newValue.Name,
		Code:        newValue.Code,
		Description: newValue.Description,
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

	recordList, err := dal.NewRepo[model.System]().QueryList(ctx.Request().Context(), database.DB, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]System, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, System{
			ID:          v.ID,
			Name:        v.Name,
			Code:        v.Code,
			Description: v.Description,
			ModifiedBy:  v.ModifiedBy,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return output.Success(ctx, map[string]any{
		"list":  list,
		"total": total,
	})
}
