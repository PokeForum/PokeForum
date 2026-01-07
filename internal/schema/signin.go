package schema

import (
	"time"
)

// SigninRequest Sign-in request | 签到请求
type SigninRequest struct {
	// User ID (obtained from JWT token, used for documentation only) | 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int64 `json:"user_id" example:"1001"`
}

// SigninResponse Sign-in response | 签到响应
type SigninResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"签到成功"`
	// Response data | 响应数据
	Data *SigninResult `json:"data"`
}

// SigninStatusRequest Get sign-in status request | 获取签到状态请求
type SigninStatusRequest struct {
	// User ID (obtained from JWT token, used for documentation only) | 用户ID（从JWT token中获取，这里仅用于文档）
	UserID int64 `json:"user_id" example:"1001"`
}

// SigninStatusResponse Get sign-in status response | 获取签到状态响应
type SigninStatusResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"获取成功"`
	// Response data | 响应数据
	Data *SigninStatus `json:"data"`
}

// DailyRankingRequest Daily ranking request | 每日排行榜请求
type DailyRankingRequest struct {
	// Query date, format: YYYY-MM-DD, query today if not provided | 查询日期，格式：YYYY-MM-DD，不传则查询今日
	Date string `json:"date" example:"2025-11-14"`
	// Return count limit, default 10, max 100 | 返回数量限制，默认10，最大100
	Limit int `json:"limit" example:"10"`
}

// DailyRankingResponse Daily ranking response | 每日排行榜响应
type DailyRankingResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"获取成功"`
	// Response data | 响应数据
	Data []*SigninRankingItem `json:"data"`
}

// ContinuousRankingRequest Continuous sign-in ranking request | 连续签到排行榜请求
type ContinuousRankingRequest struct {
	// Return count limit, default 10, max 100 | 返回数量限制，默认10，最大100
	Limit int `json:"limit" example:"10"`
}

// ContinuousRankingResponse Continuous sign-in ranking response | 连续签到排行榜响应
type ContinuousRankingResponse struct {
	// Response status code | 响应状态码
	Code int `json:"code" example:"200"`
	// Response message | 响应消息
	Message string `json:"message" example:"获取成功"`
	// Response data | 响应数据
	Data []*SigninRankingItem `json:"data"`
}

// SigninResult Sign-in result | 签到结果
type SigninResult struct {
	// Whether sign-in successful | 是否签到成功
	IsSuccess bool `json:"is_success" example:"true"`
	// Points reward earned | 获得的积分奖励
	RewardPoints int `json:"reward_points" example:"10"`
	// Experience points reward earned | 获得的经验值奖励
	RewardExperience int `json:"reward_experience" example:"10"`
	// Continuous sign-in days | 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"5"`
	// Total sign-in days | 总签到天数
	TotalDays int `json:"total_days" example:"30"`
	// Notification message | 提示信息
	Message string `json:"message" example:"签到成功！获得10积分，连续签到5天，继续加油！"`
	// Current ranking (starting from 1, 0 means not ranked) | 当前排名（从1开始，0表示未上榜）
	Rank int `json:"rank" example:"1"`
}

// SigninStatus Sign-in status | 签到状态
type SigninStatus struct {
	// Whether signed in today | 今日是否已签到
	IsTodaySigned bool `json:"is_today_signed" example:"false"`
	// Last sign-in date | 最近签到日期
	LastSigninDate *time.Time `json:"last_signin_date" example:"2025-11-13T10:30:00Z"`
	// Continuous sign-in days | 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"4"`
	// Total sign-in days | 总签到天数
	TotalDays int `json:"total_days" example:"29"`
}

// SigninRankingItem Ranking item | 排行榜项目
type SigninRankingItem struct {
	// User ID | 用户ID
	UserID int64 `json:"user_id" example:"1001"`
	// Username | 用户名
	Username string `json:"username" example:"张三"`
	// Avatar | 头像
	Avatar string `json:"avatar" example:"https://example.com/avatar.jpg"`
	// Continuous sign-in days | 连续签到天数
	ContinuousDays int `json:"continuous_days" example:"30"`
	// Total sign-in days | 总签到天数
	TotalDays int `json:"total_days" example:"100"`
	// Reward points (only for daily ranking) | 奖励积分（仅每日排行榜有此字段）
	RewardPoints int `json:"reward_points" example:"50"`
	// Ranking | 排名
	Rank int `json:"rank" example:"1"`
}

// SigninRankingResponse Sign-in ranking response | 签到排行榜响应
type SigninRankingResponse struct {
	// Ranking list | 排行榜列表
	List []*SigninRankingItem `json:"list"`
	// Current user ranking (starting from 1, 0 means not ranked) | 当前用户排名（从1开始，0表示未上榜）
	MyRank int `json:"my_rank" example:"5"`
}
