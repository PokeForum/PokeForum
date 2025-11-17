package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/ent/userbalancelog"
	"github.com/PokeForum/PokeForum/ent/usersigninstatus"
	_const "github.com/PokeForum/PokeForum/internal/const"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/internal/schema"
)

// SigninService 签到服务实现
type SigninService struct {
	db              *ent.Client
	cache           cache.ICacheService
	redisLock       *cache.RedisLock
	logger          *zap.Logger
	settingsService ISettingsService
	asyncTask       *SigninAsyncTask
}

// ISigninService 签到服务接口
type ISigninService interface {
	// Signin 执行签到
	Signin(ctx context.Context, userID int64) (*schema.SigninResult, error)
	// GetSigninStatus 获取签到状态
	GetSigninStatus(ctx context.Context, userID int64) (*schema.SigninStatus, error)
	// GetDailyRanking 获取每日排行榜
	GetDailyRanking(ctx context.Context, date string, limit int) ([]*schema.SigninRankingItem, error)
	// GetContinuousRanking 获取连续签到排行榜
	GetContinuousRanking(ctx context.Context, limit int) ([]*schema.SigninRankingItem, error)
}

// NewSigninService 创建签到服务实例
func NewSigninService(
	db *ent.Client,
	cacheService cache.ICacheService,
	redisLock *cache.RedisLock,
	logger *zap.Logger,
	settingsService ISettingsService,
	asyncTask *SigninAsyncTask,
) ISigninService {
	return &SigninService{
		db:              db,
		cache:           cacheService,
		redisLock:       redisLock,
		logger:          logger,
		settingsService: settingsService,
		asyncTask:       asyncTask,
	}
}

// Signin 执行签到
func (s *SigninService) Signin(ctx context.Context, userID int64) (*schema.SigninResult, error) {
	traceID := tracing.GetTraceID(ctx)
	s.logger.Info("开始处理签到请求",
		zap.Int64("user_id", userID),
		tracing.WithTraceIDField(ctx))

	// 检查签到功能是否启用
	enabled, err := s.isSigninEnabled(ctx)
	if err != nil {
		s.logger.Error("检查签到功能状态失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, err
	}

	if !enabled {
		s.logger.Warn("签到功能未启用",
			zap.Int64("user_id", userID),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("签到功能未启用")
	}

	// 检查用户是否存在
	userExists, err := s.db.User.Query().Where(user.ID(int(userID))).Exist(ctx)
	if err != nil {
		s.logger.Error("检查用户存在性失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("检查用户失败")
	}

	if !userExists {
		s.logger.Warn("用户不存在",
			zap.Int64("user_id", userID),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("用户不存在")
	}

	// 获取分布式锁，防止重复签到
	lockKey := fmt.Sprintf("signin:lock:%d", userID)
	lockValue := fmt.Sprintf("%s:%d", traceID, time.Now().Unix())

	lockAcquired, err := s.redisLock.Lock(ctx, lockKey, lockValue, &cache.LockOptions{
		Expiration: 10 * time.Second,
		Timeout:    5 * time.Second,
	})

	if err != nil {
		s.logger.Error("获取签到锁失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("获取资源失败")
	}

	if !lockAcquired {
		s.logger.Warn("获取签到锁超时，可能有重复签到请求",
			zap.Int64("user_id", userID),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("获取资源失败")
	}

	// 确保释放锁
	defer func() {
		err = s.redisLock.Unlock(ctx, lockKey, lockValue)
		if err != nil {
			s.logger.Error("释放签到锁失败",
				zap.Int64("user_id", userID),
				zap.Error(err),
				tracing.WithTraceIDField(ctx))
		}
	}()

	// 检查今日是否已签到
	today := time.Now().Format("2006-01-02")
	todaySigned, err := s.isTodaySigned(ctx, userID, today)
	if err != nil {
		s.logger.Error("检查今日签到状态失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("操作失败")
	}

	if todaySigned {
		s.logger.Info("用户今日已签到",
			zap.Int64("user_id", userID),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("今日已签到")
	}

	// 获取用户签到状态并计算连续签到天数
	status, err := s.getUserSigninStatus(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户签到状态失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 计算连续签到天数和奖励
	continuousDays, totalDays := s.calculateContinuousDays(status, today)
	rewardPoints, rewardExperience, err := s.calculateReward(ctx, continuousDays)
	if err != nil {
		s.logger.Error("计算签到奖励失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 更新Redis中的签到状态
	err = s.updateRedisSigninStatus(ctx, userID, today, continuousDays, totalDays)
	if err != nil {
		s.logger.Error("更新Redis签到状态失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("操作失败")
	}

	// 更新排行榜
	err = s.updateRanking(ctx, userID, today, rewardPoints)
	if err != nil {
		s.logger.Error("更新排行榜失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		// 排行榜更新失败不影响签到流程
	}

	// 异步写入数据库
	signDate, _ := time.Parse("2006-01-02", today)
	task := &SigninTask{
		UserID:         userID,
		SignDate:       signDate,
		ContinuousDays: continuousDays,
		TotalDays:      totalDays,
		TraceID:        traceID,
	}

	err = s.asyncTask.SubmitTask(task)
	if err != nil {
		s.logger.Error("提交异步任务失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))

		// 如果是队列满的错误，返回用户友好的提示
		if err.Error() == "服务繁忙，请稍后再试" {
			return nil, errors.New("服务繁忙，请稍后再试")
		}

		// 其他异步任务失败也不应该继续签到流程，确保数据一致性
		return nil, errors.New("系统繁忙，请稍后重试")
	}

	// 更新用户积分和经验（同步操作）
	err = s.updateUserBalance(ctx, userID, rewardPoints, rewardExperience)
	if err != nil {
		s.logger.Error("更新用户积分失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		return nil, errors.New("操作失败")
	}

	// 构建返回结果
	result := &schema.SigninResult{
		IsSuccess:        true,
		RewardPoints:     rewardPoints,
		RewardExperience: rewardExperience,
		ContinuousDays:   continuousDays,
		TotalDays:        totalDays,
		Message:          s.buildSigninMessage(continuousDays, rewardPoints),
	}

	s.logger.Info("签到处理完成",
		zap.Int64("user_id", userID),
		zap.Int("reward_points", rewardPoints),
		zap.Int("continuous_days", continuousDays),
		tracing.WithTraceIDField(ctx))

	return result, nil
}

// isSigninEnabled 检查签到功能是否启用
func (s *SigninService) isSigninEnabled(ctx context.Context) (bool, error) {
	value, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIsEnable, "false")
	if err != nil {
		return false, err
	}
	return value == _const.SettingBoolTrue.String(), nil
}

// isTodaySigned 检查今日是否已签到
func (s *SigninService) isTodaySigned(ctx context.Context, userID int64, today string) (bool, error) {
	statusKey := fmt.Sprintf("signin:status:%d", userID)
	lastSign, err := s.cache.HGet(statusKey, "last_sign")
	if err != nil {
		return false, err
	}
	return lastSign == today, nil
}

// getUserSigninStatus 获取用户签到状态
func (s *SigninService) getUserSigninStatus(ctx context.Context, userID int64) (*schema.SigninStatus, error) {
	statusKey := fmt.Sprintf("signin:status:%d", userID)

	// 尝试从Redis获取
	statusMap, err := s.cache.HGetAll(statusKey)
	if err != nil {
		s.logger.Error("从Redis获取签到状态失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		// Redis失败时从数据库获取
		return s.getSigninStatusFromDB(ctx, userID)
	}

	if len(statusMap) == 0 {
		// Redis中没有记录，从数据库获取
		return s.getSigninStatusFromDB(ctx, userID)
	}

	// 解析Redis数据
	continuousDays, _ := strconv.Atoi(statusMap["continuous"])
	totalDays, _ := strconv.Atoi(statusMap["total"])

	var lastSigninDate *time.Time
	if lastSignStr := statusMap["last_sign"]; lastSignStr != "" {
		if date, err := time.Parse("2006-01-02", lastSignStr); err == nil {
			lastSigninDate = &date
		}
	}

	today := time.Now().Format("2006-01-02")
	isTodaySigned := statusMap["last_sign"] == today

	// 计算明日预计奖励
	tomorrowReward := s.calculateTomorrowReward(ctx, continuousDays+1)

	return &schema.SigninStatus{
		IsTodaySigned:  isTodaySigned,
		LastSigninDate: lastSigninDate,
		ContinuousDays: continuousDays,
		TotalDays:      totalDays,
		TomorrowReward: tomorrowReward,
	}, nil
}

// getSigninStatusFromDB 从数据库获取签到状态
func (s *SigninService) getSigninStatusFromDB(ctx context.Context, userID int64) (*schema.SigninStatus, error) {
	status, err := s.db.UserSigninStatus.Query().
		Where(usersigninstatus.UserID(userID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 用户首次签到，返回默认状态
			tomorrowReward := s.calculateTomorrowReward(ctx, 1)
			return &schema.SigninStatus{
				IsTodaySigned:  false,
				LastSigninDate: nil,
				ContinuousDays: 0,
				TotalDays:      0,
				TomorrowReward: tomorrowReward,
			}, nil
		}
		return nil, err
	}

	// 计算连续签到天数
	today := time.Now().Format("2006-01-02")
	isTodaySigned := status.LastSigninDate.Format("2006-01-02") == today
	continuousDays, totalDays := s.calculateContinuousDays(&schema.SigninStatus{
		IsTodaySigned:  isTodaySigned,
		LastSigninDate: &status.LastSigninDate,
		ContinuousDays: status.ContinuousDays,
		TotalDays:      status.TotalDays,
		TomorrowReward: 0,
	}, today)

	// 计算明日预计奖励
	tomorrowReward := s.calculateTomorrowReward(ctx, continuousDays+1)

	return &schema.SigninStatus{
		IsTodaySigned:  isTodaySigned,
		LastSigninDate: &status.LastSigninDate,
		ContinuousDays: continuousDays,
		TotalDays:      totalDays,
		TomorrowReward: tomorrowReward,
	}, nil
}

// calculateContinuousDays 计算连续签到天数
func (s *SigninService) calculateContinuousDays(status *schema.SigninStatus, today string) (continuousDays, totalDays int) {
	if status.LastSigninDate == nil {
		// 首次签到
		return 1, 1
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	lastSignDate := status.LastSigninDate.Format("2006-01-02")

	if lastSignDate == yesterday {
		// 连续签到
		return status.ContinuousDays + 1, status.TotalDays + 1
	} else {
		// 中断签到，重新开始
		return 1, status.TotalDays + 1
	}
}

// calculateReward 计算签到奖励
func (s *SigninService) calculateReward(ctx context.Context, continuousDays int) (points, experience int, err error) {
	// 获取签到模式
	mode, err := s.settingsService.GetSettingByKey(ctx, _const.SigninMode, "fixed")
	if err != nil {
		return 0, 0, err
	}

	switch mode {
	case "fixed":
		return s.calculateFixedReward(ctx)
	case "increment":
		return s.calculateIncrementReward(ctx, continuousDays)
	case "random":
		return s.calculateRandomReward(ctx)
	default:
		return 0, 0, errors.New("无效的签到模式")
	}
}

// calculateFixedReward 计算固定奖励
func (s *SigninService) calculateFixedReward(ctx context.Context) (points, experience int, err error) {
	pointsStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninFixedReward, "10")
	if err != nil {
		return 0, 0, err
	}

	points, err = strconv.Atoi(pointsStr)
	if err != nil || points < 1 {
		return 0, 0, errors.New("签到配置无效")
	}

	// 计算经验值奖励
	experienceRatioStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninExperienceReward, "1")
	if err != nil {
		experienceRatioStr = "1" // 默认1:1
	}

	experienceRatio, err := strconv.ParseFloat(experienceRatioStr, 64)
	if err != nil {
		experienceRatio = 1.0
	}

	experience = int(float64(points) * experienceRatio)

	return points, experience, nil
}

// calculateIncrementReward 计算递增奖励
func (s *SigninService) calculateIncrementReward(ctx context.Context, continuousDays int) (points, experience int, err error) {
	baseStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIncrementBase, "5")
	if err != nil {
		return 0, 0, err
	}

	stepStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIncrementStep, "1")
	if err != nil {
		return 0, 0, err
	}

	// 获取递增周期配置
	cycleStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIncrementCycle, "7")
	if err != nil {
		return 0, 0, err
	}

	base, err := strconv.Atoi(baseStr)
	if err != nil || base < 1 {
		return 0, 0, errors.New("签到配置无效")
	}

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		step = 1 // 默认步长为1
	}

	cycle, err := strconv.Atoi(cycleStr)
	if err != nil || cycle < 1 {
		cycle = 7 // 默认7天周期
	}

	// 计算在当前周期内的有效天数（从1开始）
	effectiveDay := (continuousDays-1)%cycle + 1

	// 基于周期内有效天数计算奖励
	points = base + (effectiveDay-1)*step
	if points < 1 {
		points = 1 // 确保至少1个积分
	}

	// 计算经验值奖励
	experienceRatioStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninExperienceReward, "1")
	if err != nil {
		experienceRatioStr = "1"
	}

	experienceRatio, err := strconv.ParseFloat(experienceRatioStr, 64)
	if err != nil {
		experienceRatio = 1.0
	}

	experience = int(float64(points) * experienceRatio)

	return points, experience, nil
}

// calculateRandomReward 计算随机奖励
func (s *SigninService) calculateRandomReward(ctx context.Context) (points, experience int, err error) {
	minStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninRandomMin, "5")
	if err != nil {
		return 0, 0, err
	}

	maxStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninRandomMax, "20")
	if err != nil {
		return 0, 0, err
	}

	min, err := strconv.Atoi(minStr)
	if err != nil || min < 1 {
		return 0, 0, errors.New("签到配置无效")
	}

	max, err := strconv.Atoi(maxStr)
	if err != nil || max < min {
		return 0, 0, errors.New("签到配置无效")
	}

	// 使用当前时间作为随机种子，确保随机性
	rand.Seed(time.Now().UnixNano())
	points = rand.Intn(max-min+1) + min

	// 计算经验值奖励
	experienceRatioStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninExperienceReward, "1")
	if err != nil {
		experienceRatioStr = "1"
	}

	experienceRatio, err := strconv.ParseFloat(experienceRatioStr, 64)
	if err != nil {
		experienceRatio = 1.0
	}

	experience = int(float64(points) * experienceRatio)

	return points, experience, nil
}

// calculateTomorrowReward 计算明日预计奖励
func (s *SigninService) calculateTomorrowReward(ctx context.Context, continuousDays int) int {
	// 获取签到模式
	mode, err := s.settingsService.GetSettingByKey(ctx, _const.SigninMode, "fixed")
	if err != nil {
		return 1 // 默认返回1
	}

	switch mode {
	case "fixed":
		pointsStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninFixedReward, "10")
		if err != nil {
			return 1
		}
		points, err := strconv.Atoi(pointsStr)
		if err != nil || points < 1 {
			return 1
		}
		return points
	case "increment":
		baseStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIncrementBase, "5")
		if err != nil {
			return 1
		}
		stepStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninIncrementStep, "1")
		if err != nil {
			return 1
		}
		base, err := strconv.Atoi(baseStr)
		if err != nil || base < 1 {
			return 1
		}
		step, err := strconv.Atoi(stepStr)
		if err != nil {
			step = 1
		}
		tomorrowPoints := base + continuousDays*step
		if tomorrowPoints < 1 {
			return 1
		}
		return tomorrowPoints
	case "random":
		minStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninRandomMin, "5")
		if err != nil {
			return 1
		}
		maxStr, err := s.settingsService.GetSettingByKey(ctx, _const.SigninRandomMax, "20")
		if err != nil {
			return 1
		}
		min, err := strconv.Atoi(minStr)
		if err != nil || min < 1 {
			return 1
		}
		max, err := strconv.Atoi(maxStr)
		if err != nil || max < min {
			return 1
		}
		return max // 返回最大值作为明日预计
	default:
		return 1
	}
}

// updateRedisSigninStatus 更新Redis中的签到状态
func (s *SigninService) updateRedisSigninStatus(ctx context.Context, userID int64, today string, continuousDays, totalDays int) error {
	statusKey := fmt.Sprintf("signin:status:%d", userID)

	// 使用HMSET批量更新
	fieldValues := map[string]interface{}{
		"last_sign":  today,
		"continuous": continuousDays,
		"total":      totalDays,
	}

	return s.cache.HMSet(statusKey, fieldValues)
}

// updateRanking 更新排行榜
func (s *SigninService) updateRanking(ctx context.Context, userID int64, today string, rewardPoints int) error {
	// 更新每日奖励排行榜
	dailyRankingKey := fmt.Sprintf("signin:reward:%s", today)
	err := s.cache.ZAdd(dailyRankingKey, fmt.Sprintf("%d", userID), float64(rewardPoints))
	if err != nil {
		return err
	}

	// 设置过期时间（30天）
	_, err = s.cache.Expire(dailyRankingKey, 30*24*3600)
	if err != nil {
		return err
	}

	// 更新连续签到排行榜
	continuousRankingKey := "signin:continuous:ranking"
	err = s.cache.ZAdd(continuousRankingKey, fmt.Sprintf("%d", userID), float64(rewardPoints))
	if err != nil {
		return err
	}

	// 设置过期时间（永久）
	return nil
}

// updateUserBalance 更新用户积分和经验
func (s *SigninService) updateUserBalance(ctx context.Context, userID int64, points, experience int) error {
	// 这里需要调用用户服务来更新积分
	// 由于项目结构限制，暂时直接操作数据库
	// 实际项目中应该通过UserService来处理

	// 更新用户积分
	_, err := s.db.User.Update().
		Where(user.ID(int(userID))).
		AddPoints(points).
		AddExperience(experience).
		Save(ctx)

	if err != nil {
		return err
	}

	// 记录积分变动日志
	err = s.createBalanceLog(ctx, userID, "points", points, "签到奖励")
	if err != nil {
		s.logger.Error("记录积分变动日志失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		// 日志记录失败不影响主流程
	}

	// 记录经验变动日志
	err = s.createBalanceLog(ctx, userID, "experience", experience, "签到奖励")
	if err != nil {
		s.logger.Error("记录经验变动日志失败",
			zap.Int64("user_id", userID),
			zap.Error(err),
			tracing.WithTraceIDField(ctx))
		// 日志记录失败不影响主流程
	}

	return nil
}

// createBalanceLog 创建余额变动日志
func (s *SigninService) createBalanceLog(ctx context.Context, userID int64, logType string, amount int, reason string) error {
	// 获取用户当前余额
	user, err := s.db.User.Get(ctx, int(userID))
	if err != nil {
		return err
	}

	var beforeAmount, afterAmount int
	var enumType userbalancelog.Type
	switch logType {
	case "points":
		beforeAmount = user.Points
		afterAmount = user.Points + amount
		enumType = userbalancelog.TypePoints
	case "experience":
		beforeAmount = user.Experience
		afterAmount = user.Experience + amount
		enumType = userbalancelog.TypeExperience
	default:
		return errors.New("无效的日志类型")
	}

	// 创建变动日志
	return s.db.UserBalanceLog.Create().
		SetUserID(int(userID)).
		SetType(enumType).
		SetAmount(amount).
		SetBeforeAmount(beforeAmount).
		SetAfterAmount(afterAmount).
		SetReason(reason).
		SetRelatedType("signin").
		Exec(ctx)
}

// buildSigninMessage 构建签到提示信息
func (s *SigninService) buildSigninMessage(continuousDays, rewardPoints int) string {
	if continuousDays == 1 {
		return fmt.Sprintf("签到成功！获得%d积分，连续签到1天", rewardPoints)
	} else if continuousDays < 7 {
		return fmt.Sprintf("签到成功！获得%d积分，连续签到%d天，继续加油！", rewardPoints, continuousDays)
	} else if continuousDays < 30 {
		return fmt.Sprintf("签到成功！获得%d积分，连续签到%d天，太棒了！", rewardPoints, continuousDays)
	} else {
		return fmt.Sprintf("签到成功！获得%d积分，连续签到%d天，你是签到达人！", rewardPoints, continuousDays)
	}
}

// GetSigninStatus 获取用户签到状态
func (s *SigninService) GetSigninStatus(ctx context.Context, userID int64) (*schema.SigninStatus, error) {
	s.logger.Info("开始获取签到状态",
		zap.Int64("user_id", userID),
		tracing.WithTraceIDField(ctx))

	return s.getUserSigninStatus(ctx, userID)
}

// GetDailyRanking 获取每日签到排行榜
func (s *SigninService) GetDailyRanking(ctx context.Context, date string, limit int) ([]*schema.SigninRankingItem, error) {
	s.logger.Info("获取每日签到排行榜",
		zap.String("date", date),
		zap.Int("limit", limit),
		tracing.WithTraceIDField(ctx))

	// TODO: 实现每日排行榜逻辑
	return []*schema.SigninRankingItem{}, nil
}

// GetContinuousRanking 获取连续签到排行榜
func (s *SigninService) GetContinuousRanking(ctx context.Context, limit int) ([]*schema.SigninRankingItem, error) {
	s.logger.Info("获取连续签到排行榜",
		zap.Int("limit", limit),
		tracing.WithTraceIDField(ctx))

	// TODO: 实现连续签到排行榜逻辑
	return []*schema.SigninRankingItem{}, nil
}
