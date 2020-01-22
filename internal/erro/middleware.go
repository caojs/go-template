package erro

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Handler(ctx *gin.Context) {
	defer func () {
		errs := ctx.Errors
		if len(errs) > 0 {
			status := ctx.Writer.Status()

			// Error should not have status ok
			if status == http.StatusOK {
				status = 500
			}

			ctx.JSON(status, errs.Last())
		}
	}()

	ctx.Next()
}