package role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/model"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

func BatchFetch(ctx *gin.Context, codeList []string) ([]model.TblSubject, error) {
	if len(codeList) == 0 {
		err := fmt.Errorf("id list is empty")
		logger.Warnf(
			ctx,
			"role: batch fetch: invalid params, reason=%v, error=%v, code_list=%v",
			"empty id list",
			err,
			codeList,
		)
		return nil, err
	}

	recordList, err := dal.NewRepo[model.TblSubject]().Query(
		ctx,
		database.DB,
		func(db *gorm.DB) *gorm.DB {
			return db.Where("code IN ?", codeList)
		},
	)
	if err != nil {
		logger.Errorf(
			ctx,
			"role: batch fetch: failed, reason=%v, error=%v, code_list=%v",
			"db query error",
			err,
			codeList,
		)
		return nil, err
	}

	logger.Infof(
		ctx,
		"role: batch fetch: succeeded, code_list=%v, count=%d",
		codeList,
		len(recordList),
	)

	return recordList, nil
}
