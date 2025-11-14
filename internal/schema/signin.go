package schema

import (
	"time"
)

// SigninRequest 签到请求
type SigninRequest struct {
	// 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int64 `json:"user_id" example:"1001"`
}

// SigninResponse 签到响应
type SigninResponse struct {
	// 响应状态码
	Code int `json:"code" example:"200"`
	// 响应消息
	Message string `json:"message" example:"签到成功"`
	// 响应数据
	Data *SigninResult `json:"data"`
}

// SigninStatusRequest 获取签到状态请求
type SigninStatusRequest struct {
	// 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int64 `json:"user_id" example:"1001"`
}

// SigninStatusResponse 获取签到状态响应
type SigninStatusResponse struct {
	// 响应状态码
	Code int `json:"code" example:"200"`
	// 响应消息
	Message string `json:"message" example:"获取成功"`
	// 响应数据
	Data *SigninStatus `json:"data"`
}

// DailyRankingRequest 每日排行榜请求
type DailyRankingRequest struct {
	// 查询日期，格式：YYYY-MM-DD，不传则查询今日
	Date string `json:"date" example:"2025-11-14"`
	// 返回数量限制，默认10，最大100
	Limit int `json:"limit" example:"10"`
}

// DailyRankingResponse 每日排行榜响应
type DailyRankingResponse struct {
	// 响应状态码
	Code int `json:"code" example:"200"`
	// 响应消息
	Message string `json:"message" example:"获取成功"`
	// 响应数据
	Data []*SigninRankingItem `json:"data"`
}

// ContinuousRankingRequest 连续签到排行榜请求
type ContinuousRankingRequest struct {
	// 返回数量限制，默认10，最大100
	Limit int `json:"limit" example:"10"`
}

// ContinuousRankingResponse 连续签到排行榜响应
type ContinuousRankingResponse struct {
	// 响应状态码
	Code int `json:"code" example:"200"`
	// 响应消息
	Message string `json:"message" example:"获取成功"`
	// 响应数据
	Data []*SigninRankingItem `json:"data"`
}

// SigninResult 签到结果
type SigninResult struct {
	// 是否签到成功
	IsSuccess bool `json:"is_success" example:"true"`
	// 获得的积分奖励
	RewardPoints int `json:"reward_points" example:"10"`
	// 获得的经验值奖励
	RewardExperience int `json:"reward_experience" example:"10"`
	// 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"5"`
	// 总签到天数
	TotalDays int `json:"total_days" example:"30"`
	// 提示信息
	Message string `json:"message" example:"签到成功！获得10积分，连续签到5天，继续加油！"`
}

// SigninStatus 签到状态
type SigninStatus struct {
	// 今日是否已签到
	IsTodaySigned bool `json:"is_today_signed" example:"false"`
	// 最近签到日期
	LastSigninDate *time.Time `json:"last_signin_date" example:"2025-11-13T10:30:00Z"`
	// 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"4"`
	// 总签到天数
	TotalDays int `json:"total_days" example:"29"`
	// 明日预计奖励
	TomorrowReward int `json:"tomorrow_reward" example:"10"`
}

// SigninRankingItem 排行榜项目
type SigninRankingItem struct {
	// 用户ID
	UserID int64 `json:"user_id" example:"1001"`
	// 用户名
	Username string `json:"username" example:"张三"`
	// 头像
	Avatar string `json:"avatar" example:"https://example.com/avatar.jpg"`
	// 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"30"`
	// 总签到天数
	TotalDays int `json:"total_days" example:"100"`
	// 奖励积分（仅每日排行榜有此字段）
	RewardPoints int `json:"reward_points" example:"50"`
	// 排名
	Rank int `json:"rank" example:"1"`
}
