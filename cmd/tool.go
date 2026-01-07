package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/initializer"
	"github.com/PokeForum/PokeForum/internal/pkg/logging"
	"github.com/PokeForum/PokeForum/internal/utils"
)

// ToolCMD system toolkit command group | 系统工具包命令组
var ToolCMD = &cobra.Command{
	Use:   "tool",
	Short: "System toolkit | 系统工具包",
	Long:  `Provides command-line tools for various system management functions | 提供各种系统管理功能的命令行工具`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is specified, show menu | 如果没有指定子命令，显示菜单
		return showToolMenu()
	},
}

// CreateSuperAdminCMD create super admin command | 创建超级管理员命令
var CreateSuperAdminCMD = &cobra.Command{
	Use:   "create-super-admin",
	Short: "Create super admin account | 创建超级管理员账号",
	Long:  `Interactively create super admin account, requires configuration file to be set up first | 交互式创建超级管理员账号，需要先完成配置文件设置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createSuperAdmin()
	},
}

// showToolMenu displays system toolkit menu | 显示系统工具包菜单
func showToolMenu() error {
	fmt.Println("=== System Toolkit | 系统工具包 ===")
	fmt.Println("Please select an operation | 请选择要执行的操作:")
	fmt.Println("1. Create super admin account | 创建超级管理员账号")
	fmt.Println("0. Exit | 退出")
	fmt.Print("Please enter option number | 请输入选项序号: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input | 读取输入失败: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid option | 无效的选项: %s", input)
	}

	switch choice {
	case 1:
		return createSuperAdmin()
	case 0:
		fmt.Println("Exiting system toolkit | 退出系统工具包")
		return nil
	default:
		return fmt.Errorf("invalid option | 无效的选项: %d", choice)
	}
}

// createSuperAdmin creates super admin account | 创建超级管理员账号
func createSuperAdmin() error {
	fmt.Println("=== Create Super Admin Account | 创建超级管理员账号 ===")

	// Check if configuration file exists | 检查配置文件是否存在
	configPath := configs.ConfigPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file %s does not exist, please create configuration file and fill in database connection information first | 配置文件 %s 不存在，请先创建配置文件并填充数据库连接信息", configPath, configPath)
	}

	// Initialize configuration and database connection | 初始化配置和数据库连接
	fmt.Println("Initializing configuration | 正在初始化配置...")
	configs.VP = initializer.Viper(configPath)

	// Initialize logger | 初始化日志
	configs.Log = logging.Zap()

	// Initialize database | 初始化数据库
	configs.DB = initializer.DB()
	if configs.DB == nil {
		return fmt.Errorf("database initialization failed, please check the database connection information in the configuration file | 数据库初始化失败，请检查配置文件中的数据库连接信息")
	}
	defer func(DB *ent.Client) {
		err := DB.Close()
		if err != nil {
			configs.Log.Warn(err.Error())
		}
	}(configs.DB)

	// Automatically migrate database | 自动迁移数据库
	fmt.Println("Migrating database structure | 正在迁移数据库结构...")
	initializer.AutoMigrate(configs.DB)

	// Get user input | 获取用户输入
	reader := bufio.NewReader(os.Stdin)

	// Get email | 获取邮箱
	email, err := getInput(reader, "Please enter super admin email | 请输入超级管理员邮箱: ")
	if err != nil {
		return err
	}

	// Check if email already exists - use specific email search | 检查邮箱是否已存在 - 使用具体邮箱检索
	fmt.Println("Checking if email already exists | 正在检查邮箱是否已存在...")
	emailExists, err := configs.DB.User.Query().Where(user.EmailEQ(email)).Exist(context.Background())
	if err != nil {
		return fmt.Errorf("failed to query email | 查询邮箱失败: %w", err)
	}
	if emailExists {
		return fmt.Errorf("email %s is already in use | 邮箱 %s 已被使用", email, email)
	}

	// Get username | 获取用户名
	username, err := getInput(reader, "Please enter username | 请输入用户名: ")
	if err != nil {
		return err
	}

	// Check if username already exists - use specific username search | 检查用户名是否已存在 - 使用具体用户名检索
	fmt.Println("Checking if username already exists | 正在检查用户名是否已存在...")
	usernameExists, err := configs.DB.User.Query().Where(user.UsernameEQ(username)).Exist(context.Background())
	if err != nil {
		return fmt.Errorf("failed to query username | 查询用户名失败: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("username %s is already in use | 用户名 %s 已被使用", username, username)
	}

	// Get password | 获取密码
	fmt.Println(utils.GetPasswordStrengthTips())
	fmt.Println()
	password, err := getInput(reader, "Please enter password | 请输入密码: ")
	if err != nil {
		return err
	}

	// Validate password strength | 验证密码强度
	if err := utils.ValidateStrongPassword(password); err != nil {
		return fmt.Errorf("password does not meet security requirements | 密码不符合安全要求: %w", err)
	}

	// Confirm password | 确认密码
	confirmPassword, err := getInput(reader, "Please confirm password | 请确认密码: ")
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("the two passwords entered do not match | 两次输入的密码不一致")
	}

	// Generate password salt | 生成密码盐
	passwordSalt := utils.GeneratePasswordSalt()

	// Combine password with salt | 拼接密码和盐
	combinedPassword := utils.CombinePasswordWithSalt(password, passwordSalt)

	// Hash password | 加密密码
	fmt.Println("Encrypting password | 正在加密密码...")
	hashedPassword, err := utils.HashPassword(combinedPassword)
	if err != nil {
		return fmt.Errorf("password encryption failed | 密码加密失败: %w", err)
	}

	// Create super admin | 创建超级管理员
	fmt.Println("Creating super admin account | 正在创建超级管理员账号...")
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
		return fmt.Errorf("failed to create super admin | 创建超级管理员失败: %w", err)
	}

	fmt.Printf("✅ Super admin account created successfully | 超级管理员账号创建成功！\n")
	fmt.Printf("User ID | 用户ID: %d\n", newUser.ID)
	fmt.Printf("Email | 邮箱: %s\n", newUser.Email)
	fmt.Printf("Username | 用户名: %s\n", newUser.Username)
	fmt.Printf("Role | 角色: %s\n", newUser.Role)

	return nil
}

// getInput gets user input | 获取用户输入
func getInput(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input | 读取输入失败: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("input cannot be empty | 输入不能为空")
	}

	return input, nil
}

func init() {
	// Add ToolCMD as a subcommand of RootCMD | 将 ToolCMD 添加为 RootCMD 的子命令
	RootCMD.AddCommand(ToolCMD)

	// Add CreateSuperAdminCMD as a subcommand of ToolCMD | 将 CreateSuperAdminCMD 添加为 ToolCMD 的子命令
	ToolCMD.AddCommand(CreateSuperAdminCMD)
}
