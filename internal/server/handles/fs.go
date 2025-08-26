package handles

type MkdirOrLinkReq struct {
	Path string `json:"path" form:"path"`
}

// func FsMkdir(c *gin.Context) {
// 	// 创建文件的路径
// 	var request MkdirOrLinkReq
// 	if err := c.ShouldBind(&request); err != nil {
// 		common.ErrorResponse(c, err, 403)
// 		return
// 	}
// 	user := c.Request.Context().Value(configs.UserKey).(*model.User)
// 	reqPath, err := user.JoinPath(request.Path)
// 	if err != nil {
// 		common.ErrorResponse(c, err, 403)
// 		return
// 	}
// 	if err := fs.MakeDir(c.Request.Context(), reqPath); err != nil {
// 		common.ErrorResponse(c, err, 500)
// 	}
// 	common.SuccessResponse(c)
// }
