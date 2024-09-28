package army_of_the_hell

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/exp/slices"
)

var groupWhiteList = []int64{
	541065568,  // 测试群
	730561889,  // 测试群
	757553235,  // 一费奶五战队
	1006990128, // 单挑联赛Week8
}


var enableWhiteList = false
var sendGroupWelcome = true
var sendGroupMsg = false
var sendPrivateMsg = true
var botActivationRequired = true
var watchingPlayerLimit int = 5

func Handle() {
	var currentGroupId int64
	var currentGroupName string
	var currentPlayerIds []int64
	var currentPlayerNames []string
	var currentWatchingPlayerIds []int64
	var currentWatchingPlayerNames []string
	var gameStarted bool = false
	var gameLock sync.Mutex
	var game *Game

	var groupTicket = make(chan struct{})
	var groupMsgQueue = make(chan string, 102400)
	go func() {
		for {
			groupTicket <- struct{}{}
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	var privateTicket = make(chan struct{})
	var privateMsgQueue = make(chan string, 102400)
	go func() {
		for {
			privateTicket <- struct{}{}
			time.Sleep(2800 * time.Millisecond)
		}
	}()

	zero.OnCommand("地狱大军").
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID == 0 {
				ctx.Send("地狱大军只能在群聊中行动！")
				return
			}
			if !slices.Contains(groupWhiteList, ctx.Event.GroupID) && enableWhiteList {
				fmt.Println("group not in white list")
				return
			}
			gameLock.Lock()
			defer gameLock.Unlock()
			if currentGroupId != 0 {
				if gameStarted {
					go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n地狱大军正在行动中，无法同时出动！"))
				} else {
					go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n地狱大军已经在准备中，无法同时出动！所在群：" + currentGroupName))
				}
				return
			}
			if sendGroupWelcome {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n地狱大军即将出动！\n其他玩家可输入 /加入 /退出 参与游戏。"))
				// groupmsghelper.SendText("地狱大军即将出动！\n其他玩家可输入：\n /加入  加入游戏\n /退出  退出游戏")
			}
			// if sendPrivateMsg {
            // 	go ctx.SendPrivateMessage(ctx.Event.UserID, message.Message{
            // 		message.Text("地狱大军即将出动！"),
            // 	})
            // }
			currentGroupId = ctx.Event.GroupID
			currentGroupName = ctx.GetGroupInfo(ctx.Event.GroupID, false).Name
			currentPlayerIds = []int64{ctx.Event.UserID}
			currentPlayerNames = []string{ctx.Event.Sender.Name()}
			currentWatchingPlayerIds = nil
			currentWatchingPlayerNames = nil
		})
	zero.OnCommand("加入").
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID == 0 {
				ctx.Send("请在群聊中执行此指令")
				return
			}
			gameLock.Lock()
			defer gameLock.Unlock()
			if currentGroupId == 0 {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n游戏未创建，可以在群聊中发送 /地狱大军 开启一场新的地狱大军游戏。"))
				return
			}
			if gameStarted {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n加入失败：地狱大军已经开始行动！"))
				return
			}
			if currentGroupId != ctx.Event.GroupID {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n加入失败：地狱大军正在别处进行准备，所在群：" + currentGroupName))
				return
			}
			if slices.Contains(currentPlayerIds, ctx.Event.UserID) {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n您已经加入了游戏"))
				return
			}
			if slices.Contains(currentWatchingPlayerIds, ctx.Event.UserID) {
				currentWatchingPlayerIds = append(currentWatchingPlayerIds[:slices.Index(currentWatchingPlayerIds, ctx.Event.UserID)], currentWatchingPlayerIds[slices.Index(currentWatchingPlayerIds, ctx.Event.UserID)+1:]...)
			}
			if sendGroupWelcome {
				// groupmsghelper.SendText("地狱大军即将出动！")
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n加入成功——地狱大军即将出动！"))
			}
			// if sendPrivateMsg {
			// 	go ctx.SendPrivateMessage(ctx.Event.UserID, message.Message{
			// 		message.Text("地狱大军即将出动！"),
			// 	})
			// }
			currentPlayerIds = append(currentPlayerIds, ctx.Event.UserID)
			currentPlayerNames = append(currentPlayerNames, ctx.Event.Sender.Name())
		})
	zero.OnCommand("重置").
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID == 0 {
				ctx.Send("请在群聊中执行此指令")
				return
			}
			gameLock.Lock()
			defer gameLock.Unlock()
			if currentGroupId == 0 {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n重置失败：没有等待中的游戏"))
				return
			}
			if gameStarted && !zero.AdminPermission(ctx) {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n地狱大军正在行动中，当前无法重置！"))
				return
			}
			if gameStarted {
				fmt.Println("AdminPermission force reset")
				for _, id := range currentPlayerIds {
					ctx.SendPrivateMessage(id, "当前游戏进程被中断")
				}
				for _, id := range currentWatchingPlayerIds {
					ctx.SendPrivateMessage(id, "当前游戏进程被中断")
				}
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n中断当前游戏成功"))
			} else {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n重置房间成功"))
			}
			gameStarted = false
			currentGroupId = 0
			currentGroupName = ""
			currentPlayerIds = nil
			currentPlayerNames = nil
			currentWatchingPlayerIds = nil
			currentWatchingPlayerNames = nil
		})
	zero.OnCommand("围观").Handle(func(ctx *zero.Ctx) {
		e := If(ctx.Event.GroupID != 0, "\n", "").(string)
		gameLock.Lock()
		defer gameLock.Unlock()
		if currentGroupId == 0 {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "游戏未创建，可以在群聊中发送 /地狱大军 开启一场新的地狱大军游戏。"))
			return
		}
		// if currentGroupId != ctx.Event.GroupID {
		// 	go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "地狱大军正在别处行动，请到指定群聊发送指令进行观战，所在群：", currentGroupName))
		// 	return
		// }
		if slices.Contains(currentPlayerIds, ctx.Event.UserID) {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "您正在游戏中，无法进行围观"))
			return
		}
		if slices.Contains(currentWatchingPlayerIds, ctx.Event.UserID) {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "您已经加入围观了。如果您没有收到任何消息，可能是机器人私信消息未激活/已过期，需私信机器人任意消息来激活"))
			return
		}
		if len(currentWatchingPlayerIds) >= watchingPlayerLimit {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "围观失败：观众人数已达上限", watchingPlayerLimit, "人"))
			return
		}
		currentWatchingPlayerIds = append(currentWatchingPlayerIds, ctx.Event.UserID)
		currentWatchingPlayerNames = append(currentWatchingPlayerNames, ctx.Event.Sender.Name())
		go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(e, "您开始围观进行中的游戏，请给机器人*发送任意私信*来接收观战消息\n（被动消息仅有5分钟有效期，超时需再次私信激活）"))
	})
	quit := func(ctx *zero.Ctx, playerId int64) {
		if currentGroupId == 0 {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("退出失败：没有等待中的游戏"))
			return
		}
		if slices.Contains(currentWatchingPlayerIds, playerId) {
			index := slices.Index(currentWatchingPlayerIds, playerId)
			currentWatchingPlayerIds = append(currentWatchingPlayerIds[:index], currentWatchingPlayerIds[index+1:]...)
			currentWatchingPlayerNames = append(currentWatchingPlayerNames[:index], currentWatchingPlayerNames[index+1:]...)
			go ctx.SendChain(message.At(playerId), message.Text("退出围观成功！"))
			return
		}
		if gameStarted {
			if slices.Contains(currentPlayerIds, playerId) {
				index := slices.Index(currentPlayerIds, playerId)
				go ctx.SendChain(message.At(playerId), message.Text(playerId, "中途离开了游戏"))
				if game.PlayerLeave(index) {
					gameStarted = false
					currentGroupId = 0
					go ctx.SendChain(message.At(playerId), message.Text("游戏结束"))
				}
				return
			}
		}
		if slices.Contains(currentPlayerIds, playerId) {
			index := slices.Index(currentPlayerIds, playerId)
			currentPlayerIds = append(currentPlayerIds[:index], currentPlayerIds[index+1:]...)
			currentPlayerNames = append(currentPlayerNames[:index], currentPlayerNames[index+1:]...)
			if sendGroupWelcome {
				// groupmsghelper.SendText("已退出地狱大军！")
				go ctx.SendChain(message.At(playerId), message.Text("已退出地狱大军！"))
			}
			// if sendPrivateMsg {
			// 	go ctx.SendPrivateMessage(playerId, message.Message{
			// 		message.Text("已退出地狱大军！"),
			// 	})
			// }
			if len(currentPlayerIds) == 0 {
				gameStarted = false
				currentGroupId = 0
			}
		} else {
			go ctx.SendChain(message.At(playerId), message.Text("退出失败：您未加入游戏"))
		}
	}
	zero.OnCommand("退出").
		Handle(func(ctx *zero.Ctx) {
			gameLock.Lock()
			defer gameLock.Unlock()
			quit(ctx, ctx.Event.UserID)
		})
	zero.OnCommand("赛况").
		Handle(func(ctx *zero.Ctx) {
			if !gameStarted {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(If(ctx.Event.GroupID != 0, "\n", "").(string) + "游戏未开始，可以在群聊中发送 /地狱大军 开启一场新的地狱大军游戏。"))
			} else {
				// gameLock.Lock()
				// defer gameLock.Unlock()

				status := fmt.Sprintf(If(ctx.Event.GroupID != 0, "\n", "").(string) + "游戏进行中，当前第%d回合，所在群：%s。\n\n", game.Turn, currentGroupName)

				if game.WaitResponsePlayerId != -1 {
					status += "正在等待 " + game.CurrentPlayerNickname[game.WaitResponsePlayerId] + " 作出回应。"
				} else {
					for i := 0; i < game.PlayerNum; i++ {
						if game.CurrentPlayerReady[i] {
							status += game.CurrentPlayerNickname[i] + "：已行动\n"
						} else {
							status += game.CurrentPlayerNickname[i] + "：未行动\n"
						}
					}
				}

				if len(currentWatchingPlayerIds) > 0 {
					status += "\n围观观众列表："
					for i := 0; i < len(currentWatchingPlayerIds); i++ {
						status += currentWatchingPlayerNames[i] + " "
					}
				}
				ctx.Send(status)
			}
		})
	zero.OnCommand("开始").
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID == 0 {
				ctx.Send("请在群聊中执行此指令")
				return
			}
			gameLock.Lock()
			defer gameLock.Unlock()
			if currentGroupId == 0 {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n游戏未创建，可以在群聊中发送 /地狱大军 开启一场新的地狱大军游戏。"))
				return
			}
			if gameStarted {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n开始失败：地狱大军已经开始行动！"))
				return
			}
			if currentGroupId != ctx.Event.GroupID {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n开始失败：地狱大军正在别处进行准备"))
				return
			}
			if !slices.Contains(currentPlayerIds, ctx.Event.UserID) {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n开始失败：您尚未加入游戏"))
				return
			}
			gameStarted = true
			game = New(len(currentPlayerIds))
			game.PrintFunc = func(msg string) {
				fmt.Println(msg)
				msg = strings.TrimSpace(msg)
				if msg == "" {
					return
				}
				if sendGroupMsg {
					groupMsgQueue <- msg
					go func() {
						<-groupTicket
						ctx.Send(<-groupMsgQueue)
						// groupmsghelper.SendText(<-groupMsgQueue)
					}()
				}
				if sendPrivateMsg {
					privateMsgQueue <- msg
					go func() {
						<-privateTicket
						msg := <-privateMsgQueue
						for _, id := range currentPlayerIds {
							ctx.SendPrivateMessage(id, msg)
							time.Sleep(300 * time.Millisecond)
						}
						for _, id := range currentWatchingPlayerIds {
							ctx.SendPrivateMessage(id, msg)
							time.Sleep(300 * time.Millisecond)
						}
					}()
				}
			}
			for index, name := range currentPlayerNames {
				game.SetName(index, name)
			}
			if botActivationRequired {
				go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("【准备阶段】\n请所有玩家私信机器人任意消息来进行准备（用于激活机器人私信）"))
			} else {
				if !sendGroupMsg {
					go ctx.SendChain(message.At(ctx.Event.UserID), message.Text("【游戏开始】\n请玩家私信机器人进行游戏"))
				}
				game.Start()
			}
		})
	zero.OnCommand("帮助").
		Handle(func(ctx *zero.Ctx) {
			go ctx.SendChain(message.At(ctx.Event.UserID), message.Text(If(ctx.Event.GroupID != 0, "\n", "").(string) +
			"地狱大军游戏帮助：\n游戏目标是通过拍卖地狱随从获取能力，并逐渐解锁更强大的地狱生物。首先招募三个BOSS的玩家取胜。\n详细帮助请查看群文件。\n项目作者：dva\n\n相关指令：\n/加入  /退出  /围观  /开始"))
			// ctx.SendGroupForwardMessage(ctx.Event.GroupID, message.Message{
			// 	message.CustomNode("天下缟素", 2700582117, []message.MessageSegment{
			// 		message.Text("地狱大军游戏帮助："),
			// 	}),
			// 	message.CustomNode("dva", 2446629225, []message.MessageSegment{
			// 		message.Text("游戏目标是通过拍卖地狱随从获取能力，并逐渐解锁更强大的地狱生物。首先招募三个BOSS的玩家取胜。"),
			// 	}),
			// 	message.CustomNode("睦月mutsuki", 3182618911, []message.MessageSegment{
			// 		message.Text("详细游戏规则请在《单挑联赛》群文件查看。"),
			// 	}),
			// 	message.CustomNode("含墨", 2154799006, []message.MessageSegment{
			// 		message.Text("可用操作：\n/加入\n/退出\n/开始\n/帮助"),
			// 	}),
			// })
		})
	zero.OnMessage().
		Handle(func(ctx *zero.Ctx) {
			if strings.HasPrefix(ctx.Event.RawMessage, "/帮他退出") {
				if !gameStarted || !zero.AdminPermission(ctx) {
					return
				}
				fmt.Printf(ctx.Event.RawMessage)
				var cmd, qq string
				_, err1 := fmt.Sscanf(ctx.Event.RawMessage, "%s %s", &cmd, &qq)
				userId, err2 := strconv.ParseInt(qq, 10, 64)
				if err1 != nil || err2 != nil {
					ctx.Send(message.Text("帮他退出失败: 无法解析", qq))
					return
				}
				quit(ctx, userId)
				// for _, message := range ctx.Event.Message {
				// 	if message.Type == "at" {
				// 		qq, ok := message.Data["qq"]
				// 		if !ok {
				// 			continue
				// 		}
				// 		userId, err := strconv.ParseInt(qq, 10, 64)
				// 		if err != nil {
				// 			fmt.Printf("帮他退出失败: 无法解析 %s\n", qq)
				// 			continue
				// 		}
				// 		quit(ctx, userId)
				// 	}
				// }
				return
			}

			if !gameStarted {
				return
			}
			if ctx.Event.GroupID != currentGroupId {
				return
			}
			id := slices.Index(currentPlayerIds, ctx.Event.UserID)
			if id == -1 {
				return
			}
			fmt.Printf("id: %d group_msg: %v\n", id, ctx.Event.Message)

			// handle public message here.
			gameLock.Lock()
			defer gameLock.Unlock()
		})
	zero.OnMessage().
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID != 0 {
				return
			}
			if strings.HasPrefix(ctx.Event.RawMessage, "/") {
				return
			}
			if !gameStarted {
				ctx.Send("游戏未开始，请在群聊中执行指令")
				return
			}
			id := slices.Index(currentPlayerIds, ctx.Event.UserID)
			if id == -1 {
				ctx.Send("游戏已经开始，您不在游戏中，无法执行游戏请求")
				return
			}
			fmt.Printf("id: %d private_msg: %v\n", id, ctx.Event.Message)

			// handle private message here.
			if ctx.Event.RawMessage == "接受试炼" || ctx.Event.RawMessage == "通过试炼" {
				if err := game.AcceptTrial(id); err != nil {
					ctx.Send(err.Error())
				}
				return
			}

			if game.WaitResponsePlayerId == id {
				game.GiveResponse(id, ctx.Event.RawMessage)
				return
			}

			gameLock.Lock()
			defer gameLock.Unlock()
			if botActivationRequired && game.Turn == 0 {
				if err := game.Ready(id); err != nil {
					ctx.Send(err.Error())
				} else if game.CurrentPlayerReady[id] {
					ctx.Send("准备成功！")
				}
				return
			}
			if game.SingleMode || game.CurrentBiddingEntity2.Name == "" {
				price, err := strconv.Atoi(ctx.Event.RawMessage)
				if err != nil {
					fmt.Println("waiting for number, got ", ctx.Event.RawMessage)
					ctx.Send("当前只有一个角色拍卖，仅需提交一个数字")
					return
				}
				if err := game.GivePrice(id, price); err != nil {
					ctx.Send(err.Error())
				} else if game.CurrentPlayerReady[id] {
					ctx.Send("出价成功。")
				}
			} else {
				var price1, price2 int
				_, err := fmt.Sscanf(ctx.Event.RawMessage, "%d %d", &price1, &price2)
				if err != nil {
					fmt.Println("waiting for 2 number, got ", ctx.Event.RawMessage)
					ctx.Send("当前为双拍模式，需要提交两个数字，以空格分隔")
					return
				}
				if err := game.GivePrices(id, price1, price2); err != nil {
					ctx.Send(err.Error())
				} else if game.CurrentPlayerReady[id] {
					ctx.Send("出价成功。")
				}
			}

			if scores := game.GetScores(); scores != nil {
				gameStarted = false
				currentGroupId = 0
			}
		})
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}