package api

import (
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Database interface {
	FindByFilenameAndType(filename string, imageType image.Type) ([]image.Image, bool, error)
}

type getThumbsRequest struct {
	Filename string `uri:"filename" binding:"required"`
}

func GetThumbsHandler(db Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := getThumbsRequest{}
		if err := ctx.ShouldBindUri(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if t, exists, err := db.FindByFilenameAndType(request.Filename, image.Thumb); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
		} else if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("%s was not found", request.Filename),
			})
		} else {
			ctx.JSON(http.StatusOK, t)
		}
	}
}
