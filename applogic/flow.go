package applogic

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// ApplyPromptChoiceQ 应用promptchoiceQ的逻辑，动态修改requestmsg
func (app *App) ApplyPromptChoiceQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// 检查是否启用了EnhancedQA
	if config.GetEnhancedQA(promptstr) {
		promptChoices := config.GetPromptChoicesQ(promptstr)
		if len(promptChoices) == 0 {
			// 获取系统历史，但不包括系统消息
			systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
			if err != nil {
				fmt.Printf("Error getting system history ApplyPromptChoiceQ: %v\n", err)
				return
			}

			// 如果有系统历史并且有至少一个消息
			if len(systemHistory) > 0 {
				lastSystemMessage := systemHistory[len(systemHistory)-2] // 获取最后一个系统消息
				// 将最后一个系统历史消息附加到用户消息后
				*requestmsg += " (" + lastSystemMessage.Text + ")"
			}
		} else {
			var ischange bool
			// 获取用户剧情存档中的当前状态
			CustomRecord, err := app.FetchCustomRecord(userid)
			if err != nil {
				fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
				return
			}

			// 获取当前场景的总对话长度
			PromptMarksLength := config.GetPromptMarksLength(promptstr)

			// 计算当前对话轮次
			currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
			fmt.Printf("故事模式:当前对话轮次Q %v\n", currentRound)

			// 遍历所有的 promptChoices 配置项
			var randomChoices []structs.PromptChoice
			for _, choice := range promptChoices {
				if choice.Round == currentRound {
					if len(choice.Keywords) == 0 {
						// Keywords 为空，收集所有符合条件的 ReplaceText
						randomChoices = append(randomChoices, choice)
					} else {
						bestMatchCount := 0
						bestText := ""
						// 遍历每组触发词设置
						for _, keyword := range choice.Keywords {
							matchCount := 0
							if strings.Contains(*requestmsg, keyword) {
								matchCount++
							}
							// 如果当前组的匹配数量最高，更新最佳文本
							if matchCount > bestMatchCount {
								bestMatchCount = matchCount
								bestText = choice.ReplaceText[rand.Intn(len(choice.ReplaceText))]
							}
						}
						// 如果找到了有效的触发词组合，附加最佳文本到消息中
						if bestMatchCount > 0 {
							*requestmsg += " (" + bestText + ")"

						}
					}
				}
			}

			// 处理 randomChoices 中的随机选择 从符合轮次的里选一个 如果选出的包含多个 就再随机选一个
			if len(randomChoices) > 0 {
				selectedChoice := randomChoices[rand.Intn(len(randomChoices))]
				*requestmsg += " (" + selectedChoice.ReplaceText[rand.Intn(len(selectedChoice.ReplaceText))] + ")"
			}

			// 如果内容没有改变,回滚到用最后一个Q来加入对话中
			if !ischange {
				// 获取系统历史，但不包括系统消息
				systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
				if err != nil {
					fmt.Printf("Error getting system history GetMessagesExcludingSystem: %v\n", err)
					return
				}

				// 如果有系统历史并且有至少一个消息
				if len(systemHistory) > 0 {
					lastSystemMessage := systemHistory[len(systemHistory)-2] // 获取最后一个系统消息
					// 将最后一个系统历史消息附加到用户消息后
					*requestmsg += " (" + lastSystemMessage.Text + ")"
				}
			}
		}
	}
}

// ApplyPromptCoverQ 应用promptCoverQ的逻辑，动态覆盖requestmsg
func (app *App) ApplyPromptCoverQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// 检查是否启用了EnhancedQA
	if config.GetEnhancedQA(promptstr) {
		promptCover := config.GetPromptCoverQ(promptstr)
		if len(promptCover) == 0 {
			// 直接返回
			return
		} else {
			var ischange bool
			// 获取用户剧情存档中的当前状态
			CustomRecord, err := app.FetchCustomRecord(userid)
			if err != nil {
				fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
				return
			}

			// 获取当前场景的总对话长度
			PromptMarksLength := config.GetPromptMarksLength(promptstr)

			// 计算当前对话轮次
			currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
			fmt.Printf("故事模式覆盖Q:当前对话轮次Q %v\n", currentRound)

			// 遍历所有的 promptChoices 配置项
			var randomChoices []structs.PromptChoice
			for _, choice := range promptCover {
				if choice.Round == currentRound {
					if len(choice.Keywords) == 0 {
						// Keywords 为空，收集所有符合条件的 ReplaceText
						randomChoices = append(randomChoices, choice)
					} else {
						bestMatchCount := 0
						bestText := ""
						// 遍历每组触发词设置
						for _, keyword := range choice.Keywords {
							matchCount := 0
							if strings.Contains(*requestmsg, keyword) {
								matchCount++
							}
							// 如果当前组的匹配数量最高，更新最佳文本
							if matchCount > bestMatchCount {
								bestMatchCount = matchCount
								bestText = choice.ReplaceText[rand.Intn(len(choice.ReplaceText))]
							}
						}
						// 如果找到了有效的触发词组合，附加最佳文本到消息中
						if bestMatchCount > 0 {
							*requestmsg = bestText
							ischange = true
						}
					}
				}
			}

			// 处理 randomChoices 中的随机选择 从符合轮次的里选一个 如果选出的包含多个 就再随机选一个
			if len(randomChoices) > 0 {
				selectedChoice := randomChoices[rand.Intn(len(randomChoices))]
				*requestmsg = selectedChoice.ReplaceText[rand.Intn(len(selectedChoice.ReplaceText))]
				ischange = true
			}

			// 如果内容没有改变,回滚到用最后一个Q来覆盖对话中
			if !ischange {
				// 获取系统历史，但不包括系统消息
				systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
				if err != nil {
					fmt.Printf("Error getting system history GetMessagesExcludingSystem: %v\n", err)
					return
				}

				// 如果有系统历史并且有至少一个消息
				if len(systemHistory) > 0 {
					lastSystemMessage := systemHistory[len(systemHistory)-2] // 获取最后一个系统消息
					// 将最后一个系统历史消息覆盖用户消息
					*requestmsg = lastSystemMessage.Text
				}
			}
		}
	}
}

// ApplySwitchOnQ 应用switchOnQ的逻辑，动态修改promptstr
func (app *App) ApplySwitchOnQ(promptstr *string, requestmsg *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// promptstr 随 switchOnQ 变化
	promptstrChoices := config.GetSwitchOnQ(*promptstr)
	if len(promptstrChoices) != 0 {
		// 获取用户剧情存档中的当前状态
		CustomRecord, err := app.FetchCustomRecord(userid)
		if err != nil {
			fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
			return
		}

		// 获取当前场景的总对话长度
		PromptMarksLength := config.GetPromptMarksLength(*promptstr)

		// 计算当前对话轮次
		currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
		fmt.Printf("关键词切换分支状态:当前对话轮次Q %v,当前promptstr:%s\n", currentRound, *promptstr)

		// 遍历所有的PromptSwitch配置项
		var randomChoices []structs.PromptSwitch
		for _, choice := range promptstrChoices {
			if choice.Round == currentRound {
				if len(choice.Keywords) == 0 {
					// Keywords为空，收集所有符合条件的Switch
					randomChoices = append(randomChoices, choice)
				} else {
					bestMatchCount := 0
					var bestText string
					// 遍历每组关键词设置
					for _, keyword := range choice.Keywords {
						matchCount := 0
						if strings.Contains(*requestmsg, keyword) {
							matchCount++
						}
						// 如果当前组的匹配数量最高，更新最佳文本
						if matchCount > bestMatchCount {
							bestMatchCount = matchCount
							bestText = choice.Switch[rand.Intn(len(choice.Switch))]
						}
					}
					// 如果找到了有效的触发词组合，修改分支
					if bestMatchCount > 0 {
						*promptstr = bestText
						// 获取新的信号长度 刷新持久化数据库
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("根据关键词切换prompt为: %s, newPromptStrStat: %d\n", *promptstr, PromptMarksLength)
						// 应用 PromptChoiceQ
						app.ApplyPromptChoiceQ(*promptstr, requestmsg, message)
					}
				}
			}
		}

		// 处理 randomChoices 中的随机选择 从符合轮次的里选一个 如果选出的包含多个 就再随机选一个
		if len(randomChoices) > 0 {
			selectedChoice := randomChoices[rand.Intn(len(randomChoices))]
			selectedText := selectedChoice.Switch[rand.Intn(len(selectedChoice.Switch))]
			*promptstr = selectedText
			// 获取新的信号长度 刷新持久化数据库
			PromptMarksLength := config.GetPromptMarksLength(*promptstr)
			app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
			fmt.Printf("随机选择prompt为: %s, newPromptStrStat: %d\n", *promptstr, PromptMarksLength)
			// 应用 PromptChoiceQ
			app.ApplyPromptChoiceQ(*promptstr, requestmsg, message)
		}

	}
}

// ProcessExitChoicesQ 处理配置中的退出选择逻辑，根据特定触发词决定是否触发退出行为。
func (app *App) ProcessExitChoicesQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage, selfid string) {
	exitChoices := config.GetExitOnQ(promptstr)
	if len(exitChoices) == 0 {
		return
	}

	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// 获取用户剧情存档中的当前状态
	CustomRecord, err := app.FetchCustomRecord(userid)
	if err != nil {
		fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
		return
	}

	// 获取当前场景的总对话长度
	PromptMarksLength := config.GetPromptMarksLength(promptstr)

	// 计算当前对话轮次
	currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
	fmt.Printf("关键词判断退出分支:当前对话轮次Q %v\n", currentRound)

	// 遍历所有的 promptChoices 配置项
	for _, choice := range exitChoices {
		if choice.Round == currentRound {
			bestMatchCount := 0
			bestText := ""
			// 遍历每组触发词设置
			for _, keyword := range choice.Keywords {
				matchCount := 0
				if strings.Contains(*requestmsg, keyword) {
					matchCount++
				}
				// 如果当前组的匹配数量最高，更新最佳文本
				if matchCount > bestMatchCount {
					bestMatchCount = matchCount
					bestText = keyword
				}
			}
			// 如果找到了有效的触发词组合，就退出分支
			if bestMatchCount > 0 {
				app.HandleExit(bestText, message, selfid, promptstr)
				return
			}
		}
	}
}

// HandleExit 处理用户退出逻辑，包括发送消息和重置用户状态。
func (app *App) HandleExit(exitText string, message *structs.OnebotGroupMessage, selfid string, promptstr string) {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	fmt.Printf("处理重置操作on:%v", exitText)
	app.migrateUserToNewContext(userid)
	RestoreResponse := config.GetRandomRestoreResponses()
	if message.RealMessageType == "group_private" || message.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			utils.SendPrivateMessage(message.UserID, RestoreResponse, selfid, promptstr)
		} else {
			utils.SendSSEPrivateRestoreMessage(message.UserID, RestoreResponse, promptstr, selfid)
		}
	} else {
		utils.SendGroupMessage(message.GroupID, message.UserID, RestoreResponse, selfid, promptstr)
	}
	app.deleteCustomRecord(userid)
}

// ProcessExitChoicesA 处理基于关键词的退出逻辑。
func (app *App) ProcessExitChoicesA(promptstr string, response *string, message *structs.OnebotGroupMessage, selfid string) {
	exitChoices := config.GetExitOnA(promptstr)
	if len(exitChoices) == 0 {
		return
	}

	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// 获取用户剧情存档中的当前状态
	CustomRecord, err := app.FetchCustomRecord(userid)
	if err != nil {
		fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
		return
	}

	// 获取当前场景的总对话长度
	PromptMarksLength := config.GetPromptMarksLength(promptstr)

	// 计算当前对话轮次
	currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
	fmt.Printf("关键词判断退出分支:当前对话轮次A %v\n", currentRound)

	// 遍历所有的 promptChoices 配置项
	for _, choice := range exitChoices {
		if choice.Round == currentRound {
			bestMatchCount := 0
			bestText := ""
			// 遍历每组触发词设置
			for _, keyword := range choice.Keywords {
				matchCount := 0
				if strings.Contains(*response, keyword) {
					matchCount++
				}
				// 如果当前组的匹配数量最高，更新最佳文本
				if matchCount > bestMatchCount {
					bestMatchCount = matchCount
					bestText = keyword
				}
			}
			// 如果找到了有效的触发词组合，就退出分支
			if bestMatchCount > 0 {
				app.HandleExit(bestText, message, selfid, promptstr)
				return
			}
		}
	}
}

// ApplySwitchOnA 应用switchOnA的逻辑，动态修改promptstr
func (app *App) ApplySwitchOnA(promptstr *string, response *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	// 获取与 switchOnA 相关的选择
	promptstrChoices := config.GetSwitchOnA(*promptstr)
	if len(promptstrChoices) != 0 {
		// 获取用户剧情存档中的当前状态
		CustomRecord, err := app.FetchCustomRecord(userid)
		if err != nil {
			fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
			return
		}

		// 获取当前场景的总对话长度
		PromptMarksLength := config.GetPromptMarksLength(*promptstr)

		// 计算当前对话轮次
		currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
		fmt.Printf("关键词[%v]切换分支状态:当前对话轮次A %v", *response, currentRound)

		// 遍历所有的PromptSwitch配置项
		var randomChoices []structs.PromptSwitch
		for _, choice := range promptstrChoices {
			if choice.Round == currentRound {
				if len(choice.Keywords) == 0 {
					// Keywords为空，收集所有符合条件的Switch
					randomChoices = append(randomChoices, choice)
				} else {
					bestMatchCount := 0
					var bestText string
					// 遍历每组关键词设置
					for _, keyword := range choice.Keywords {
						matchCount := 0
						if strings.Contains(*response, keyword) {
							matchCount++
						}
						// 如果当前组的匹配数量最高，更新最佳文本
						if matchCount > bestMatchCount {
							bestMatchCount = matchCount
							bestText = choice.Switch[rand.Intn(len(choice.Switch))]
						}
					}
					// 如果找到了有效的触发词组合，修改分支
					if bestMatchCount > 0 {
						*promptstr = bestText
						// 获取新的信号长度 刷新持久化数据库
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("根据关键词切换prompt为: %s, newPromptStrStat: %d\n", *promptstr, PromptMarksLength)
					}
				}
			}
		}

		// 处理 randomChoices 中的随机选择 从符合轮次的里选一个 如果选出的包含多个 就再随机选一个
		if len(randomChoices) > 0 {
			selectedChoice := randomChoices[rand.Intn(len(randomChoices))]
			selectedText := selectedChoice.Switch[rand.Intn(len(selectedChoice.Switch))]
			*promptstr = selectedText
			// 获取新的信号长度 刷新持久化数据库
			PromptMarksLength := config.GetPromptMarksLength(*promptstr)
			app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
			fmt.Printf("随机选择prompt为: %s, newPromptStrStat: %d\n", *promptstr, PromptMarksLength)
		}
	}
}

// ApplyPromptChoiceA 应用故事模式的情绪增强逻辑，并返回增强内容。
func (app *App) ApplyPromptChoiceA(promptstr string, response string, message *structs.OnebotGroupMessage) string {
	userid := message.UserID + message.SelfID
	if config.GetGroupContext() == 2 && message.MessageType != "private" {
		userid = message.GroupID + message.SelfID
	}

	promptChoices := config.GetPromptChoicesA(promptstr)
	if len(promptChoices) == 0 {
		// 获取系统历史，但不包括系统消息
		systemHistory, err := prompt.GetMessagesExcludingSystem(promptstr)
		if err != nil {
			fmt.Printf("Error getting system history GetMessagesExcludingSystem: %v\n", err)
			return ""
		}

		// 如果有系统历史并且有至少一个消息
		if len(systemHistory) > 0 {
			lastSystemMessage := systemHistory[len(systemHistory)-1] // 获取最后一个消息 角色是assistant
			// 将最后一个系统历史消息附加到用户消息后
			return " (" + lastSystemMessage.Text + ")"
		}
		// 如果systemHistory没有内容 且 promptChoices 长度是0
		return ""
	}

	// 获取用户剧情存档中的当前状态
	CustomRecord, err := app.FetchCustomRecord(userid)
	if err != nil {
		fmt.Printf("app.FetchCustomRecord 出错: %s\n", err)
		return ""
	}

	// 获取当前场景的总对话长度
	PromptMarksLength := config.GetPromptMarksLength(promptstr)

	// 计算当前对话轮次
	currentRound := PromptMarksLength - CustomRecord.PromptStrStat + 1
	fmt.Printf("故事模式:当前对话轮次A %v\n", currentRound)

	// 遍历所有的 promptChoices 配置项
	var randomChoices []structs.PromptChoice
	for _, choice := range promptChoices {
		if choice.Round == currentRound {
			if len(choice.Keywords) == 0 {
				// Keywords 为空，收集所有符合条件的 ReplaceText
				randomChoices = append(randomChoices, choice)
			} else {
				bestMatchCount := 0
				bestText := ""
				// 遍历每组触发词设置
				for _, keyword := range choice.Keywords {
					matchCount := 0
					if strings.Contains(response, keyword) {
						matchCount++
					}
					// 如果当前组的匹配数量最高，更新最佳文本
					if matchCount > bestMatchCount {
						bestMatchCount = matchCount
						bestText = choice.ReplaceText[rand.Intn(len(choice.ReplaceText))]
					}
				}
				// 如果找到了有效的触发词组合，返回最佳文本，会附加到当前的llm回复后方
				if bestMatchCount > 0 {
					return "(" + bestText + ")"

				}
			}
		}
	}

	// 处理 randomChoices 中的随机选择 从符合轮次的里选一个 如果选出的包含多个 就再随机选一个
	if len(randomChoices) > 0 {
		selectedChoice := randomChoices[rand.Intn(len(randomChoices))]
		return " (" + selectedChoice.ReplaceText[rand.Intn(len(selectedChoice.ReplaceText))] + ")"
	}

	// 默认 没有匹配到任何内容时
	return ""
}

// ApplyPromptChanceQ 应用promptChanceQ的逻辑，动态修改requestmsg
func (app *App) ApplyPromptChanceQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage) {
	// 获取PromptChance数组
	promptChances := config.GetPromptChanceQ(promptstr)

	// 遍历所有的 promptChances 配置项
	for _, chance := range promptChances {
		// 基于概率进行计算
		if rand.Intn(100) < chance.Probability {
			*requestmsg += " (" + chance.Text + ")"
		}
	}
}
