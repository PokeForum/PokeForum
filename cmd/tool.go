package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/initializer"
	"github.com/PokeForum/PokeForum/internal/pkg/logging"
	"github.com/PokeForum/PokeForum/internal/utils"
	"github.com/spf13/cobra"
)

// ToolCMD 系统工具包命令组
var ToolCMD = &cobra.Command{
	Use:   "tool",
	Short: "系统工具包",
	Long:  `提供各种系统管理功能的命令行工具`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 如果没有指定子命令，显示菜单
		return showToolMenu()
	},
}

// CreateSuperAdminCMD 创建超级管理员命令
var CreateSuperAdminCMD = &cobra.Command{
	Use:   "create-super-admin",
	Short: "创建超级管理员账号",
	Long:  `交互式创建超级管理员账号，需要先完成配置文件设置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createSuperAdmin()
	},
}

// showToolMenu 显示系统工具包菜单
func showToolMenu() error {
	fmt.Println("=== 系统工具包 ===")
	fmt.Println("请选择要执行的操作:")
	fmt.Println("1. 创建超级管理员账号")
	fmt.Println("0. 退出")
	fmt.Print("请输入选项序号: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取输入失败: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("无效的选项: %s", input)
	}

	switch choice {
	case 1:
		return createSuperAdmin()
	case 0:
		fmt.Println("退出系统工具包")
		return nil
	default:
		return fmt.Errorf("无效的选项: %d", choice)
	}
}

// createSuperAdmin 创建超级管理员账号
func createSuperAdmin() error {
	fmt.Println("=== 创建超级管理员账号 ===")

	// 检查配置文件是否存在
	configPath := configs.ConfigPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件 %s 不存在，请先创建配置文件并填充数据库连接信息", configPath)
	}

	// 初始化配置和数据库连接
	fmt.Println("正在初始化配置...")
	configs.VP = initializer.Viper(configPath)

	// 初始化日志
	configs.Log = logging.Zap()

	// 初始化数据库
	configs.DB = initializer.DB()
	if configs.DB == nil {
		return fmt.Errorf("数据库初始化失败，请检查配置文件中的数据库连接信息")
	}
	defer func(DB *ent.Client) {
		err := DB.Close()
		if err != nil {
			configs.Log.Warn(err.Error())
		}
	}(configs.DB)

	// 自动迁移数据库
	fmt.Println("正在迁移数据库结构...")
	initializer.AutoMigrate(configs.DB)

	// 获取用户输入
	reader := bufio.NewReader(os.Stdin)

	// 获取邮箱
	email, err := getInput(reader, "请输入超级管理员邮箱: ")
	if err != nil {
		return err
	}

	// 检查邮箱是否已存在 - 使用具体邮箱检索
	fmt.Println("正在检查邮箱是否已存在...")
	emailExists, err := configs.DB.User.Query().Where(user.EmailEQ(email)).Exist(context.Background())
	if err != nil {
		return fmt.Errorf("查询邮箱失败: %w", err)
	}
	if emailExists {
		return fmt.Errorf("邮箱 %s 已被使用", email)
	}

	// 获取用户名
	username, err := getInput(reader, "请输入用户名: ")
	if err != nil {
		return err
	}

	// 检查用户名是否已存在 - 使用具体用户名检索
	fmt.Println("正在检查用户名是否已存在...")
	usernameExists, err := configs.DB.User.Query().Where(user.UsernameEQ(username)).Exist(context.Background())
	if err != nil {
		return fmt.Errorf("查询用户名失败: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("用户名 %s 已被使用", username)
	}

	// 获取密码
	fmt.Println(utils.GetPasswordStrengthTips())
	fmt.Println()
	password, err := getInput(reader, "请输入密码: ")
	if err != nil {
		return err
	}

	// 验证密码强度
	if err := utils.ValidateStrongPassword(password); err != nil {
		return fmt.Errorf("密码不符合安全要求: %w", err)
	}

	// 确认密码
	confirmPassword, err := getInput(reader, "请确认密码: ")
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 生成密码盐
	passwordSalt := utils.GeneratePasswordSalt()

	// 拼接密码和盐
	combinedPassword := utils.CombinePasswordWithSalt(password, passwordSalt)

	// 加密密码
	fmt.Println("正在加密密码...")
	hashedPassword, err := utils.HashPassword(combinedPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建超级管理员
	fmt.Println("正在创建超级管理员账号...")
	newUser, err := configs.DB.User.Create().
		SetEmail(email).
		SetUsername(username).
		SetPassword(hashedPassword).
		SetPasswordSalt(passwordSalt).
		SetRole("SuperAdmin").
		SetEmailVerified(true).
		SetStatus("Normal").
		Save(context.Background())

	if err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}

	fmt.Printf("✅ 超级管理员账号创建成功！\n")
	fmt.Printf("用户ID: %d\n", newUser.ID)
	fmt.Printf("邮箱: %s\n", newUser.Email)
	fmt.Printf("用户名: %s\n", newUser.Username)
	fmt.Printf("角色: %s\n", newUser.Role)

	return nil
}

// getInput 获取用户输入
func getInput(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	return input, nil
}

func init() {
	// 将 ToolCMD 添加为 RootCMD 的子命令
	RootCMD.AddCommand(ToolCMD)

	// 将 CreateSuperAdminCMD 添加为 ToolCMD 的子命令
	ToolCMD.AddCommand(CreateSuperAdminCMD)
}
