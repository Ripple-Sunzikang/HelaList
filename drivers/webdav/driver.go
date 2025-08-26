package webdav

import (
	"HelaList/internal/model"
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path"

	"github.com/the-plate/gowebdav"
)

type WebDav struct {
	model.Storage
	client   *gowebdav.Client // 成品的webdav库客户端
	Addition                  // 附加信息
}

type Addition struct {
	Vendor   string `json:"vendor" type:"select" options:"sharepoinnt,other" default:"other"`
	Address  string `json:"address" required:"true"` // Address为服务器连接
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	//driver.RootPath
	TlsInsecureSkipVerify bool `json:"tls_insecure_skip_verify" default:"false"`
}

func (wd *WebDav) setClient() error {
	c := gowebdav.NewClient(wd.Address, wd.Username, wd.Password)
	c.SetTransport(&http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: wd.TlsInsecureSkipVerify},
	})

	wd.client = c
	return nil
}

// 将一个切片，通过某种函数转换关系，转换为另一个切片
// 因为你需要把gowebdav的文件类型转换为你的文件类型
func ConvertSlices[S any, D any](sourceS []S, convert func(sourceS S) (D, error)) ([]D, error) {
	res := make([]D, 0, len(sourceS))
	for i := range sourceS {
		desti, err := convert(sourceS[i])
		if err != nil {
			return nil, err
		}
		res = append(res, desti)
	}
	return res, nil
}

// 这是待会儿要用到的妙妙工具
func getPath(obj model.Obj) string {
	if obj.IsDir() {
		return obj.GetPath() + "/"
	}
	return obj.GetPath()
}

func (wd *WebDav) Init(ctx context.Context) error {
	err := wd.setClient()
	if err == nil {
		// 执行重复验证
	}
	return err
}

func (wd *WebDav) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := wd.client.ReadDir(ctx, dir.GetPath())
	if err != nil {
		return nil, err
	}
	return ConvertSlices(files, func(src os.FileInfo) (model.Obj, error) {
		return &model.Object{
			Name:         src.Name(),
			Size:         src.Size(),
			ModifiedTime: src.ModTime(),
			IsFolder:     src.IsDir(),
		}, nil
	})
}

func (wd *WebDav) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return wd.client.MkdirAll(ctx, path.Join(parentDir.GetPath(), dirName), 0644)
}

func (wd *WebDav) Move(ctx context.Context, sourceObj model.Obj, destiObj model.Obj) error {
	return wd.client.Rename(ctx, getPath(sourceObj), path.Join(destiObj.GetPath(), sourceObj.GetName()), true)
}

func (wd *WebDav) Rename(ctx context.Context, sourceObj model.Obj, newName string) error {
	return wd.client.Rename(ctx, getPath(sourceObj), path.Join(path.Dir(sourceObj.GetPath()), newName), true)
}

func (wd *WebDav) Copy(ctx context.Context, sourceObj, destiDir model.Obj) error {
	return wd.client.Copy(ctx, getPath(sourceObj), path.Join(destiDir.GetPath(), sourceObj.GetName()), true)
}

func (wd *WebDav) Remove(ctx context.Context, obj model.Obj) error {
	return wd.client.RemoveAll(ctx, getPath(obj))
}

// 文件链接的事情等会儿再说，你连link的类型都没写呢。
// func (d *WebDav) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
// 	url, header, err := d.client.Link(file.GetPath())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &model.Link{
// 		URL:    url,
// 		Header: header,
// 	}, nil
// }

// 上传的事情等会儿再说
// func (wd *WebDav) Put(ctx context.Context, dstDir model.Obj, s model.FileStreamer, up driver.UpdateProgress) error {
// 	callback := func(r *http.Request) {
// 		r.Header.Set("Content-Type", s.GetMimetype())
// 		r.ContentLength = s.GetSize()
// 	}
// 	reader := driver.NewLimitedUploadStream(ctx, &driver.ReaderUpdatingProgress{
// 		Reader:         s,
// 		UpdateProgress: up,
// 	})
// 	err := wd.client.WriteStream(path.Join(dstDir.GetPath(), s.GetName()), reader, 0644, callback)
// 	return err
// }
