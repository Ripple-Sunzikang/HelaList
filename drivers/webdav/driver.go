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
	Addition                  // 附加信息，用于webdav登陆认证
	client   *gowebdav.Client // 成品的webdav库客户端，来自gowebdav库
}

// 附加信息，用于webdav登陆认证
type Addition struct {
	Vendor   string `json:"vendor" type:"select" options:"sharepoinnt,other" default:"other"`
	Address  string `json:"address" required:"true"` // Address为服务器端地址
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootPath
	TlsInsecureSkipVerify bool `json:"tls_insecure_skip_verify" default:"false"`
}

// webdav自己的配置常量
var config = driver.Config{
	Name:        "webdav",
	LocalSort:   true,
	OnlyProxy:   true, // WebDAV需要代理访问
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

func (d *WebDav) Init(ctx context.Context) error {
	err := d.setClient()
	if err == nil {
	}
	return err
}

// init的作用是将数据库中注册的网盘写入内存的map当中，详情见/op/driver.go的driverMap
func init() {
	op.RegisterDriver(func() driver.Driver {
		return &WebDav{}
	})
}

func (d *WebDav) Config() driver.Config {
	return config
}

func (d *WebDav) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *WebDav) Drop(ctx context.Context) error {
	// todo: 定期执行服务器连接刷新
	/*
		因为数据库只有在后端启动时将服务器内容读入缓存，
		这会导致如果后续服务器端出现什么变动，HelaList端没有及时更新数据，
		将无法看到服务器端的变化
	*/
	return nil
}

// 列举指定文件夹的目录
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

func (d *WebDav) Put(ctx context.Context, dstDir model.Obj, s model.FileStreamer, up driver.UpdateProgress) error {
	callback := func(r *http.Request) {
		r.Header.Set("Content-Type", s.GetMimetype())
		r.ContentLength = s.GetSize()
	}
	reader := driver.NewLimitedUploadStream(ctx, &driver.ReaderUpdatingProgress{
		Reader:         s,
		UpdateProgress: up,
	})
	err := d.client.WriteStream(path.Join(dstDir.GetPath(), s.GetName()), reader, 0644, callback)
	return err
}

func (d *WebDav) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url, header, err := d.client.Link(file.GetPath())
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL:    url,
		Header: header,
	}, nil
}

// 用于获取指定文件的路径
func getPath(obj model.Obj) string {
	if obj.IsDir() {
		return obj.GetPath() + "/"
	}
	return obj.GetPath()
}

var _ driver.Driver = (*WebDav)(nil)
