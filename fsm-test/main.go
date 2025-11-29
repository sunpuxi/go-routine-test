package main

import (
	"context"
	"fmt"
	"github.com/looplab/fsm"
	"goRedisLock/tool"
)

// 用户状态
const (
	StatePendingReview = "pending_review" // 待审核（初始状态）
	StateApproved      = "approved"       // 审核通过
	StateRejected      = "rejected"       // 审核拒绝
	StateActive        = "active"         // 活跃（审核通过后初始状态）
	StateInactive      = "inactive"       // 不活跃（30天无消耗）
	StateOverdue       = "overdue"        // 逾期（90天无消耗）
)

// 触发状态转换的事件
const (
	EventApprove      = "approve"        // 审批通过操作
	EventReject       = "reject"         // 审批拒绝操作
	EventConsume      = "consume"        // 用户发生消费
	EventNoConsume30d = "no_consume_30d" // 定时任务检测到30天无消耗
	EventNoConsume90d = "no_consume_90d" // 定时任务检测到90天无消耗
)

// UserFsm User 模型，包含状态机实例
type UserFsm struct {
	ID   string
	Name string
	// ... 其他用户字段
	FSM          *fsm.FSM `json:"-"` // 不序列化到JSON
	CurrentState string   // 显式记录当前状态，用于持久化
}

// NewUser 创建新用户实例，初始化状态机为"待审核"
func NewUser(id, name string, currentStatus string) *UserFsm {
	user := &UserFsm{
		ID:           id,
		Name:         name,
		CurrentState: currentStatus,
	}

	user.FSM = fsm.NewFSM(
		user.CurrentState, // 初始状态
		fsm.Events{
			// 审核相关事件
			{Name: EventApprove, Src: []string{StatePendingReview}, Dst: StateApproved},
			{Name: EventReject, Src: []string{StatePendingReview}, Dst: StateRejected},

			// 消费状态流转事件
			// 审核通过后，用户初始为"活跃"状态，可以理解为"已激活"
			{Name: EventNoConsume30d, Src: []string{StateApproved, StateActive}, Dst: StateInactive},
			{Name: EventNoConsume90d, Src: []string{StateInactive}, Dst: StateOverdue},
			// 用户一旦消费，无论是从"Inactive"还是"Overdue"，都回归"Active"状态
			{Name: EventConsume, Src: []string{StateInactive, StateOverdue}, Dst: StateActive},
		},
		fsm.Callbacks{
			// 通用回调：每次成功进入新状态后，更新User结构体的CurrentState并持久化到DB
			"enter_state": func(_ context.Context, e *fsm.Event) {
				// e.Dst 是目标状态
				user.CurrentState = e.Dst
				fmt.Printf("用户 %s 状态变更: %s -> %s\n", user.ID, e.Src, e.Dst)
				// 关键：调用方法将新状态安全更新到数据库
				if err := user.updateStateInDB(e.Dst); err != nil {
					// 处理数据库更新失败的情况，例如记录日志、告警等
					// 注意：此时内存中的状态机状态已变更，但数据库未更新，可能需要重试或人工干预
					fmt.Printf("错误：更新用户 %s 数据库状态失败: %v\n", user.ID, err)
				}
			},
			// 可以根据需要添加更具体的回调，例如审核通过后发送通知
			"enter_approved": func(_ context.Context, e *fsm.Event) {
				fmt.Printf("通知：用户 %s 审核已通过！\n", user.ID)
				// 这里可以调用发送邮件、短信等的逻辑
			},
		},
	)
	return user
}

// Approve 审核操作
func (u *UserFsm) Approve(ctx context.Context) error {
	return u.FSM.Event(ctx, EventApprove)
}

func (u *UserFsm) Reject(ctx context.Context) error {
	return u.FSM.Event(ctx, EventReject)
}

// MarkInactiveAfter30d 定时任务调用的方法
func (u *UserFsm) MarkInactiveAfter30d(ctx context.Context) error {
	return u.FSM.Event(ctx, EventNoConsume30d)
}

func (u *UserFsm) MarkOverdueAfter90d(ctx context.Context) error {
	return u.FSM.Event(ctx, EventNoConsume90d)
}

// OnConsume 用户消费时调用的方法
func (u *UserFsm) OnConsume(ctx context.Context) error {
	return u.FSM.Event(ctx, EventConsume)
}

func (u *UserFsm) updateStateInDB(targetStatus string) error {
	fmt.Printf("status is gonna be target status is %s", targetStatus)
	return nil
}

func main() {
	ctx := context.Background()
	newUserFsm := NewUser("111", "Daniel", StateApproved)
	fmt.Printf("%s\n", tool.JsonEncode(newUserFsm))

	// 模拟当前的操作为审核通过操作
	err := newUserFsm.Approve(ctx)
	if err != nil {
		fmt.Printf("approve failed: %v\n", err)
	}

}
