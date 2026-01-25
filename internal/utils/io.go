package utils

import (
	"io"
	"os"
	"path/filepath"
)

// Exists checks if the specified file or directory exists | 检查指定的文件或目录是否存在
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile creates a file for the given path, recursively creating directories if they don't exist | 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (*os.File, error) {
	// Get the directory where the file is located | 获取文件所在的目录
	basePath := filepath.Dir(path)
	// If the directory doesn't exist, create it recursively | 如果目录不存在，递归创建
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			return nil, err
		}
	}

	// Create the file | 创建文件
	return os.Create(path) //nolint:gosec // Path is controlled by caller, security is verified at call site
}

// CreatNestedFolder creates a folder using the given path, recursively creating directories if they don't exist | 使用给定的路径创建文件夹，如果目录不存在则递归创建
func CreatNestedFolder(path string) error {
	// If the directory doesn't exist, create it recursively | 如果目录不存在，递归创建
	if !Exists(path) {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsEmpty returns whether the given directory is empty | 返回给定目录是否为空目录
func IsEmpty(name string) (bool, error) {
	// Open the directory | 打开目录
	f, err := os.Open(name) //nolint:gosec // Path is controlled by caller, security is verified at call site
	if err != nil {
		return false, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Try to read one directory entry, if EOF is returned, the directory is empty | 尝试读取一个目录项，如果返回 EOF 说明目录为空
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	// Return false if there's an error or the directory is not empty | 如果有错误或目录不为空都返回 false
	return false, err
}

// CallbackReader is a reader wrapper with a callback function that gets called on each read | 是一个包装了回调函数的读取器，在每次读取时调用回调函数
type CallbackReader struct {
	// reader is the underlying reader | reader 底层的读取器
	reader io.Reader
	// callback is the function called after each read, parameter is the number of bytes read | callback 每次读取后调用的回调函数，参数为读取的字节数
	callback func(int64)
}

// NewCallbackReader creates a new CallbackReader instance | 创建一个新的 CallbackReader 实例
func NewCallbackReader(reader io.Reader, callback func(int64)) *CallbackReader {
	return &CallbackReader{
		reader:   reader,
		callback: callback,
	}
}

// Read implements the io.Reader interface, reads data and calls the callback function | 实现 io.Reader 接口，读取数据并调用回调函数
func (r *CallbackReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.callback(int64(n))
	return
}
