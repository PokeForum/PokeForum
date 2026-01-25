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

// UseWorkingDir determines whether to use working directory instead of executable file directory | 是否使用工作目录而不是可执行文件目录
var UseWorkingDir = false

// DotPathToStandardPath converts comma-separated path to standard path | 将","分割的路径转换为标准路径
func DotPathToStandardPath(path string) string {
	// Replace commas with slashes and add / at the beginning | 将逗号替换为斜杠，并在前面添加 /
	return "/" + strings.ReplaceAll(path, ",", "/")
}

// FillSlash adds `/` to the end of path | 给路径补全`/`
func FillSlash(path string) string {
	// If path is already /, return directly | 如果路径已经是 /，则直接返回
	if path == "/" {
		return path
	}
	// Add / at the end of path | 在路径末尾添加 /
	return path + "/"
}

// RemoveSlash removes the trailing `/` from path | 移除路径最后的`/`
func RemoveSlash(path string) string {
	// If path length is greater than 1, remove trailing / | 如果路径长度大于 1，移除末尾的 /
	if len(path) > 1 {
		return strings.TrimSuffix(path, "/")
	}
	// Otherwise return directly (keep single /) | 否则直接返回（保留单个 /）
	return path
}

// SplitPath splits path into a list | 分割路径为列表
func SplitPath(path string) []string {
	// If path is empty or doesn't start with /, return empty list | 如果路径为空或不以 / 开头，返回空列表
	if len(path) == 0 || path[0] != '/' {
		return []string{}
	}

	// If path is just /, return list containing / | 如果路径只是 /，返回包含 / 的列表
	if path == "/" {
		return []string{"/"}
	}

	// Split path by / | 按 / 分割路径
	pathSplit := strings.Split(path, "/")
	// Set first element to / | 将第一个元素设置为 /
	pathSplit[0] = "/"
	return pathSplit
}

// FormSlash replaces backslashes '\' with '/' in path | 将path中的反斜杠'\'替换为'/'
func FormSlash(old string) string {
	// Replace all backslashes with forward slashes, then clean the path | 替换所有反斜杠为正斜杠，然后清理路径
	return path.Clean(strings.ReplaceAll(old, "\\", "/"))
}

// MkdirIfNotExist creates directory if it doesn't exist | 如果目录不存在则创建目录
func MkdirIfNotExist(ctx context.Context, p string) {
	// Check if directory exists, if not, create it recursively | 检查目录是否存在，不存在则递归创建
	if !Exists(p) {
		_ = os.MkdirAll(p, 0o700) //nolint:errcheck // Directory creation failure will be reported in subsequent operations
	}
}

// SlashClean is equivalent to path.Clean("/" + name), but slightly more efficient | 等价于 path.Clean("/" + name)，但效率略高
func SlashClean(name string) string {
	// If path doesn't start with /, add / | 如果路径不以 / 开头，则添加 /
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	// Clean the path | 清理路径
	return path.Clean(name)
}

// Ext returns the file extension used by the path, excluding the dot | 返回路径使用的文件扩展名，不包含点号
func Ext(name string) string {
	// Get file extension and convert to lowercase | 获取文件扩展名并转换为小写
	ext := strings.ToLower(filepath.Ext(name))
	// Remove the dot before extension | 移除扩展名前的点号
	if len(ext) > 0 {
		ext = ext[1:]
	}
	return ext
}
