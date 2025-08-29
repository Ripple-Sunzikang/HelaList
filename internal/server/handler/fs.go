package handler

import (
	"HelaList/configs"
	"HelaList/internal/fs"
	"HelaList/internal/model"
	"HelaList/internal/server/common"
	"HelaList/internal/stream"
	"errors"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
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

	objs, err := fs.List(c.Request.Context(), reqPath, &fs.ListArgs{Refresh: req.Refresh})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	resp := FsListResp{
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

// FsCopyMoveReq defines the request structure for copy and move operations.
type FsCopyMoveReq struct {
	SrcPath string `json:"src_path" binding:"required"`
	DstPath string `json:"dst_path" binding:"required"`
}

// FsCopyHandler handles the file/folder copy operation.
func FsCopyHandler(c *gin.Context) {
	var req FsCopyMoveReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	user := c.Request.Context().Value(configs.UserKey).(*model.User)

	srcPath, err := user.JoinPath(req.SrcPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}
	dstPath, err := user.JoinPath(req.DstPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	if err := fs.Copy(c.Request.Context(), srcPath, dstPath); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	common.SuccessResponse(c)
}

// FsMoveHandler handles the file/folder move operation.
func FsMoveHandler(c *gin.Context) {
	var req FsCopyMoveReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	user := c.Request.Context().Value(configs.UserKey).(*model.User)

	srcPath, err := user.JoinPath(req.SrcPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}
	dstPath, err := user.JoinPath(req.DstPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	if err := fs.Move(c.Request.Context(), srcPath, dstPath); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	common.SuccessResponse(c)
}

// FsPutHandler handles the direct file upload operation.
func FsPutHandler(c *gin.Context) {
	// 从 multipart form 中获取目标路径
	dstPath := c.PostForm("path")
	if dstPath == "" {
		common.ErrorResponse(c, errors.New("destination path is required"), 400)
		return
	}

	// 从 multipart form 中获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		common.ErrorResponse(c, errors.New("file is required in multipart form"), 400)
		return
	}

	// 打开上传的文件以获取其 io.Reader
	file, err := fileHeader.Open()
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
	// file 是一个 multipart.File, 它实现了 io.ReadCloser 接口

	// 获取当前用户信息并检查权限
	user := c.Request.Context().Value(configs.UserKey).(*model.User)

	// 转换成绝对路径
	reqPath, err := user.JoinPath(dstPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// 创建一个符合 fs.PutDirectly 要求的 model.FileStreamer 对象
	fileStream := &stream.FileStream{
		Ctx: c.Request.Context(),
		Obj: &model.Object{
			Name:         fileHeader.Filename,
			Size:         fileHeader.Size,
			ModifiedTime: time.Now(),
		},
		Reader: file,
		// 必须将 file 添加到 Closers 中，以便在操作结束后正确关闭文件句柄
		Closers: utils.NewClosers(file),
	}

	// 调用核心的 put 方法
	if err := fs.PutDirectly(c.Request.Context(), reqPath, fileStream); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	// 成功响应
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

// RenameReq defines the request structure for rename operation.
type RenameReq struct {
	Path string `json:"path" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// FsRenameHandler handles the file/folder rename operation.
func FsRenameHandler(c *gin.Context) {
	var req RenameReq
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

	if err := fs.Rename(c.Request.Context(), reqPath, req.Name); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	common.SuccessResponse(c)
}

// FsRemoveReq defines the request structure for remove operation.
type FsRemoveReq struct {
	Path string `json:"path" binding:"required"`
}

// FsRemoveHandler handles the file/folder remove operation.
func FsRemoveHandler(c *gin.Context) {
	var req FsRemoveReq
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

	if err := fs.Remove(c.Request.Context(), reqPath); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	common.SuccessResponse(c)
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
