package webdav

import (
	"HelaList/internal/driver"
	"HelaList/internal/model"
	"HelaList/internal/op"
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path"

	"github.com/OpenListTeam/OpenList/v4/pkg/gowebdav"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	// "github.com/the-plate/gowebdav"
)

type WebDav struct {
	model.Storage
	Addition                  // 附加信息
	client   *gowebdav.Client // 成品的webdav库客户端
}

func (d *WebDav) Config() driver.Config {
	return config
}

func (d *WebDav) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *WebDav) Init(ctx context.Context) error {
	err := d.setClient()
	if err == nil {
		// 执行定时刷新操作的
	}
	return err
}

func (d *WebDav) Drop(ctx context.Context) error {
	// 检查定时刷新
	return nil
}

func (d *WebDav) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.client.ReadDir(dir.GetPath())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src os.FileInfo) (model.Obj, error) {
		return &model.Object{
			Name:         src.Name(),
			Size:         src.Size(),
			ModifiedTime: src.ModTime(),
			IsFolder:     src.IsDir(),
		}, nil
	})
}

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

func (d *WebDav) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return d.client.MkdirAll(path.Join(parentDir.GetPath(), dirName), 0644)
}

func (d *WebDav) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Rename(getPath(srcObj), path.Join(dstDir.GetPath(), srcObj.GetName()), true)
}

func (d *WebDav) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return d.client.Rename(getPath(srcObj), path.Join(path.Dir(srcObj.GetPath()), newName), true)
}

func (d *WebDav) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Copy(getPath(srcObj), path.Join(dstDir.GetPath(), srcObj.GetName()), true)
}

func (d *WebDav) Remove(ctx context.Context, obj model.Obj) error {
	return d.client.RemoveAll(getPath(obj))
}

// func (d *WebDav) Put(ctx context.Context, dstDir model.Obj, s model.FileStreamer, up driver.UpdateProgress) error {
//
// }

var _ driver.Driver = (*WebDav)(nil)

type Addition struct {
	Vendor   string `json:"vendor" type:"select" options:"sharepoinnt,other" default:"other"`
	Address  string `json:"address" required:"true"` // Address为服务器连接
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootPath
	TlsInsecureSkipVerify bool `json:"tls_insecure_skip_verify" default:"false"`
}

var config = driver.Config{
	Name:        "webdav",
	LocalSort:   true,
	OnlyProxy:   true,
	DefaultRoot: "/",
}

// 设置驱动层客户端，调用gowebdav
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
// 用于替换utils.SliceConvert
// func ConvertSlices[S any, D any](sourceS []S, convert func(sourceS S) (D, error)) ([]D, error) {
// 	res := make([]D, 0, len(sourceS))
// 	for i := range sourceS {
// 		desti, err := convert(sourceS[i])
// 		if err != nil {
// 			return nil, err
// 		}
// 		res = append(res, desti)
// 	}
// 	return res, nil
// }

// 这是待会儿要用到的妙妙工具
func getPath(obj model.Obj) string {
	if obj.IsDir() {
		return obj.GetPath() + "/"
	}
	return obj.GetPath()
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &WebDav{}
	})
}
