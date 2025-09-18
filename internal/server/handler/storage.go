package handler

import (
	"HelaList/internal/model"
	"HelaList/internal/op"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateStorageHandler(c *gin.Context) {
	var storage model.Storage
	if err := c.ShouldBindJSON(&storage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := op.CreateStorage(c.Request.Context(), storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func UpdateStorageHandler(c *gin.Context) {
	var storage model.Storage
	if err := c.ShouldBindJSON(&storage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := op.UpdateStorage(c.Request.Context(), storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func GetStorageByMountPathHandler(c *gin.Context) {
	mountPath := c.Param("mountPath")
	driver, err := op.GetStorageByMountPath(mountPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, driver.GetStorage())
}

func GetAllStoragesHandler(c *gin.Context) {
	storages := op.GetAllStorages()
	var storageModels []model.Storage
	for _, d := range storages {
		storageModels = append(storageModels, *d.GetStorage())
	}
	c.JSON(http.StatusOK, storageModels)
}

func HasStorageHandler(c *gin.Context) {
	mountPath := c.Param("mountPath")
	has := op.HasStorage(mountPath)
	c.JSON(http.StatusOK, gin.H{"has": has})
}

func GetStorageVirtualFilesByPathHandler(c *gin.Context) {
	prefix := c.Query("prefix")
	files := op.GetStorageVirtualFilesByPath(prefix)
	c.JSON(http.StatusOK, files)
}

func LoadStorageHandler(c *gin.Context) {
	var storage model.Storage
	if err := c.ShouldBindJSON(&storage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := op.LoadStorage(c.Request.Context(), storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "loaded"})
}

func DeleteStorageHandler(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "存储ID不能为空"})
		return
	}

	err := op.DeleteStorage(c.Request.Context(), storageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
