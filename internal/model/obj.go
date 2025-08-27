package model

import (
	"HelaList/configs"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dlclark/regexp2"
	"github.com/google/uuid"
)

/*
严格来说，Obj接口才是文件最抽象的形式。其次才是Object。
*/
type Obj interface {
	GetSize() int64
	GetName() string
	GetModifiedTime() time.Time
	GetCreatedTime() time.Time
	IsDir() bool
	// GetHash() // 哈希还没实现，所以暂时不做
	GetId() uuid.UUID
	GetPath() string
}

type ObjUnwrap interface {
	Unwrap() Obj
}

// 将原始文件名进行转换/映射
type ObjWrapName struct {
	Name string
	Obj
}

func UnwrapObj(obj Obj) Obj {
	if unwrap, ok := obj.(ObjUnwrap); ok {
		obj = unwrap.Unwrap()
	}
	return obj
}

// 便利conf.FilenameCharMap，对所有传入的name逐项替换
func MappingName(name string) string {
	for k, v := range configs.FilenameCharMap {
		name = strings.ReplaceAll(name, k, v)
	}
	return name
}

func WrapObjName(objs Obj) Obj {
	return &ObjWrapName{
		Name: MappingName(objs.GetName()),
		Obj:  objs,
	}
}

func WrapObjsName(objs []Obj) {
	for i := 0; i < len(objs); i++ {
		objs[i] = &ObjWrapName{Name: MappingName(objs[i].GetName()), Obj: objs[i]}
	}
}

type SetPath interface {
	SetPath(path string)
}

// Merge
func NewObjMerge() *ObjMerge {
	return &ObjMerge{
		set: mapset.NewSet[string](),
	}
}

type ObjMerge struct {
	regs []*regexp2.Regexp
	set  mapset.Set[string]
}

func (om *ObjMerge) Merge(objs []Obj, objs_ ...Obj) []Obj {
	newObjs := make([]Obj, 0, len(objs)+len(objs_))
	newObjs = om.insertObjs(om.insertObjs(newObjs, objs...), objs_...)
	return newObjs
}

func (om *ObjMerge) insertObjs(objs []Obj, objs_ ...Obj) []Obj {
	for _, obj := range objs_ {
		if om.clickObj(obj) {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (om *ObjMerge) clickObj(obj Obj) bool {
	for _, reg := range om.regs {
		if isMatch, _ := reg.MatchString(obj.GetName()); isMatch {
			return false
		}
	}
	return om.set.Add(obj.GetName())
}

func (om *ObjMerge) InitHideReg(hides string) {
	rs := strings.Split(hides, "\n")
	om.regs = make([]*regexp2.Regexp, 0, len(rs))
	for _, r := range rs {
		om.regs = append(om.regs, regexp2.MustCompile(r, regexp2.None))
	}
}

func (om *ObjMerge) Reset() {
	om.set.Clear()
}
