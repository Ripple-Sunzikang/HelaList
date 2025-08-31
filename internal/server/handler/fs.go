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

// 获取文件信息
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

type ListReq struct {
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
	Refresh  bool   `json:"refresh"`
}

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

// 只列举文件夹
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

type FsCopyMoveReq struct {
	SrcPath string `json:"src_path" binding:"required"`
	DstPath string `json:"dst_path" binding:"required"`
}

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

func FsPutHandler(c *gin.Context) {
	dstPath := c.PostForm("path")
	if dstPath == "" {
		common.ErrorResponse(c, errors.New("destination path is required"), 400)
		return
	}

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

	// 获取当前用户信息并检查权限
	user := c.Request.Context().Value(configs.UserKey).(*model.User)

	// 转换成绝对路径
	reqPath, err := user.JoinPath(dstPath)
	if err != nil {
		common.ErrorResponse(c, err, 403)
		return
	}

	// 创建一个符合fs.PutDirectly要求的model.FileStreamer对象
	fileStream := &stream.FileStream{
		Ctx: c.Request.Context(),
		Obj: &model.Object{
			Name:         fileHeader.Filename,
			Size:         fileHeader.Size,
			ModifiedTime: time.Now(),
		},
		Reader:  file,
		Closers: utils.NewClosers(file),
	}

	if err := fs.PutDirectly(c.Request.Context(), reqPath, fileStream); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	// 成功响应
	common.SuccessResponse(c)
}

type RenameReq struct {
	Path string `json:"path" binding:"required"`
	Name string `json:"name" binding:"required"`
}

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

type FsRemoveReq struct {
	Path string `json:"path" binding:"required"`
}

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

type LinkReq struct {
	Path string `json:"path" form:"path" binding:"required"`
}

func FsLinkHandler(c *gin.Context) {
	var req LinkReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

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

	link, _, err := fs.Link(c.Request.Context(), reqPath, model.LinkArgs{})
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	common.SuccessResponse(c, link)
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
