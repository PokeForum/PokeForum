package utils

import (
	"io"
	"os"
	"path/filepath"
)

// Exists 检查指定的文件或目录是否存在
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (*os.File, error) {
	// 获取文件所在的目录
	basePath := filepath.Dir(path)
	// 如果目录不存在，递归创建
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			return nil, err
		}
	}

	// 创建文件
	return os.Create(path) //nolint:gosec // 路径由调用方控制，已在调用处验证安全性
}

// CreatNestedFolder 使用给定的路径创建文件夹，如果目录不存在则递归创建
func CreatNestedFolder(path string) error {
	// 如果目录不存在，递归创建
	if !Exists(path) {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsEmpty 返回给定目录是否为空目录
func IsEmpty(name string) (bool, error) {
	// 打开目录
	f, err := os.Open(name) //nolint:gosec // 路径由调用方控制，已在调用处验证安全性
	if err != nil {
		return false, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// 尝试读取一个目录项，如果返回 EOF 说明目录为空
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	// 如果有错误或目录不为空都返回 false
	return false, err
}

// CallbackReader 是一个包装了回调函数的读取器，在每次读取时调用回调函数
type CallbackReader struct {
	// reader 底层的读取器
	reader io.Reader
	// callback 每次读取后调用的回调函数，参数为读取的字节数
	callback func(int64)
}

// NewCallbackReader 创建一个新的 CallbackReader 实例
func NewCallbackReader(reader io.Reader, callback func(int64)) *CallbackReader {
	return &CallbackReader{
		reader:   reader,
		callback: callback,
	}
}

// Read 实现 io.Reader 接口，读取数据并调用回调函数
func (r *CallbackReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.callback(int64(n))
	return
}
