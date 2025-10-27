package utils

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DataFolder = "data"
)

// UseWorkingDir 是否使用工作目录而不是可执行文件目录
var UseWorkingDir = false

// DotPathToStandardPath 将","分割的路径转换为标准路径
func DotPathToStandardPath(path string) string {
	// 将逗号替换为斜杠，并在前面添加 /
	return "/" + strings.Replace(path, ",", "/", -1)
}

// FillSlash 给路径补全`/`
func FillSlash(path string) string {
	// 如果路径已经是 /，则直接返回
	if path == "/" {
		return path
	}
	// 在路径末尾添加 /
	return path + "/"
}

// RemoveSlash 移除路径最后的`/`
func RemoveSlash(path string) string {
	// 如果路径长度大于 1，移除末尾的 /
	if len(path) > 1 {
		return strings.TrimSuffix(path, "/")
	}
	// 否则直接返回（保留单个 /）
	return path
}

// SplitPath 分割路径为列表
func SplitPath(path string) []string {
	// 如果路径为空或不以 / 开头，返回空列表
	if len(path) == 0 || path[0] != '/' {
		return []string{}
	}

	// 如果路径只是 /，返回包含 / 的列表
	if path == "/" {
		return []string{"/"}
	}

	// 按 / 分割路径
	pathSplit := strings.Split(path, "/")
	// 将第一个元素设置为 /
	pathSplit[0] = "/"
	return pathSplit
}

// FormSlash 将path中的反斜杠'\'替换为'/'
func FormSlash(old string) string {
	// 替换所有反斜杠为正斜杠，然后清理路径
	return path.Clean(strings.ReplaceAll(old, "\\", "/"))
}

// RelativePath 获取相对可执行文件的路径
func RelativePath(name string) string {
	// 如果设置了使用工作目录，直接返回原路径
	if UseWorkingDir {
		return name
	}

	// 如果是绝对路径，直接返回
	if filepath.IsAbs(name) {
		return name
	}
	// 获取可执行文件所在目录，并拼接相对路径
	e, _ := os.Executable()
	return filepath.Join(filepath.Dir(e), name)
}

// DataPath 获取用于存储持久化数据文件的相对路径
func DataPath(child string) string {
	// 获取数据文件夹的路径
	dataPath := RelativePath(DataFolder)
	// 如果文件夹不存在则创建
	if !Exists(dataPath) {
		_ = os.MkdirAll(dataPath, 0700)
	}

	// 如果子路径是绝对路径，直接返回
	if filepath.IsAbs(child) {
		return child
	}

	// 拼接数据文件夹路径和子路径
	return filepath.Join(dataPath, child)
}

// MkdirIfNotExist 如果目录不存在则创建目录
func MkdirIfNotExist(ctx context.Context, p string) {
	// 检查目录是否存在，不存在则递归创建
	if !Exists(p) {
		_ = os.MkdirAll(p, 0700)
	}
}

// SlashClean 等价于 path.Clean("/" + name)，但效率略高
func SlashClean(name string) string {
	// 如果路径不以 / 开头，则添加 /
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	// 清理路径
	return path.Clean(name)
}

// Ext 返回路径使用的文件扩展名，不包含点号
func Ext(name string) string {
	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(filepath.Ext(name))
	// 移除扩展名前的点号
	if len(ext) > 0 {
		ext = ext[1:]
	}
	return ext
}
