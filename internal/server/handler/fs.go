package handler

import (
	"HelaList/configs"
	"HelaList/internal/fs"
	"HelaList/internal/model"
	"HelaList/internal/server/common"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListReq represents the request for listing files
type ListReq struct {
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
	Refresh  bool   `json:"refresh"`
}

// FsListHandler handles file listing
func FsListHandler(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	// 从 URL 路径参数获取路径
	req.Path = c.Param("path")

	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if user.IsGuest() && user.Disabled {
		common.ErrorResponse(c, errors.New("guest user is disabled"), 401)
		return
	}

	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// meta, err := op.GetNearestMeta(reqPath)
	// if err != nil {
	// 	if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 		common.ErrorResponse(c, err, 500)
	// 		return
	// 	}
	// }

	objs, err := fs.List(c.Request.Context(), reqPath, &fs.ListArgs{Refresh: req.Refresh})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	// // 现在并不很需要分页
	// total := len(objs)
	// start := (req.Page - 1) * req.PerPage
	// if start < 0 {
	// 	start = 0
	// }
	// end := start + req.PerPage
	// if end > total {
	// 	end = total
	// }
	// paginatedObjs := objs[start:end]

	resp := FsListResp{
		// 直接转换完整的 objs 列表，而不是分页后的 paginatedObjs
		Content: toObjsResp(objs, reqPath),
		Total:   int64(len(objs)),
		Write:   true,
	}

	common.SuccessResponse(c, resp)
}

// FsDirsHandler handles directory listing
func FsDirsHandler(c *gin.Context) {
	var req struct {
		Path      string `json:"path" form:"path"`
		Password  string `json:"password" form:"password"`
		ForceRoot bool   `json:"force_root" form:"force_root"`
	}
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	// 从 URL 路径参数获取路径
	req.Path = c.Param("path")

	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	reqPath := req.Path
	if !req.ForceRoot {
		var err error
		reqPath, err = user.JoinPath(req.Path)
		if err != nil {
			common.ErrorResponse(c, err, 403)
			return
		}
	} else if !user.IsAdmin() {
		common.ErrorResponse(c, errors.New("permission denied"), 403)
		return
	}

	// meta, err := op.GetNearestMeta(reqPath)
	// if err != nil {
	// 	if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 		common.ErrorResponse(c, err, 500)
	// 		return
	// 	}
	// }

	objs, err := fs.List(c.Request.Context(), reqPath, &fs.ListArgs{})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	dirs := filterDirs(objs)
	common.SuccessResponse(c, dirs)
}

// FsGetHandler handles getting file information
func FsGetHandler(c *gin.Context) {
	var req struct {
		Path     string `json:"path" form:"path"`
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	// 从 URL 路径参数获取路径
	req.Path = c.Param("path")

	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if user.IsGuest() && user.Disabled {
		common.ErrorResponse(c, errors.New("guest user is disabled"), 401)
		return
	}

	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// meta, err := op.GetNearestMeta(reqPath)
	// if err != nil {
	// 	if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 		common.ErrorResponse(c, err, 500)
	// 		return
	// 	}
	// }

	obj, err := fs.Get(c.Request.Context(), reqPath)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	resp := toObjResp(obj, reqPath)
	common.SuccessResponse(c, resp)
}

// 从文件和文件夹的组合中，找出所有文件夹
func filterDirs(objs []model.Obj) []DirResp {
	var dirs []DirResp
	for _, obj := range objs {
		if obj.IsDir() {
			dirs = append(dirs, DirResp{
				Name:     obj.GetName(),
				Modified: obj.GetModifiedTime(),
			})
		}
	}
	return dirs
}

// 将Obj转为api响应对象

func toObjsResp(objs []model.Obj, parent string) []ObjResp {
	var resp []ObjResp
	for _, obj := range objs {
		resp = append(resp, ObjResp{
			Id:       obj.GetId().String(),
			Path:     obj.GetPath(),
			Name:     obj.GetName(),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.GetModifiedTime(),
			Created:  obj.GetCreatedTime(),
		})
	}
	return resp
}

type MkdirOrLinkReq struct {
	Path string `json:"path" form:"path"`
}

func FsMkdir(c *gin.Context) {
	var req MkdirOrLinkReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}
	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	if err := fs.MakeDir(c.Request.Context(), reqPath); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
	common.SuccessResponse(c)
}

func toObjResp(obj model.Obj, parent string) ObjResp {
	return ObjResp{
		Id:       obj.GetId().String(),
		Path:     obj.GetPath(),
		Name:     obj.GetName(),
		Size:     obj.GetSize(),
		IsDir:    obj.IsDir(),
		Modified: obj.GetModifiedTime(),
		Created:  obj.GetCreatedTime(),
	}
}

type RenameReq struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Overwrite bool   `json:"overwrite"`
}

func checkRelativePath(path string) error {
	if strings.ContainsAny(path, "/\\") || path == "" || path == "." || path == ".." {
		return errors.New("relativePath")
	}
	return nil
}

// api的响应对象

type ObjResp struct {
	Id       string    `json:"id"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
}

type DirResp struct {
	Name     string    `json:"name"`
	Modified time.Time `json:"modified"`
}

type FsListResp struct {
	Content []ObjResp `json:"content"`
	Total   int64     `json:"total"`
	Write   bool      `json:"write"`
}
