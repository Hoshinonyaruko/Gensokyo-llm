package applogic

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/prompt"
	"github.com/hoshinonyaruko/gensokyo-llm/structs"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

// ApplyPromptChoiceQ 应用promptchoiceQ的逻辑，动态修改requestmsg
func (app *App) ApplyPromptChoiceQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

			enhancedChoices := config.GetEnhancedPromptChoices(promptstr)
			if enhancedChoices {
				// 遍历所有的promptChoices配置项
				for _, choice := range promptChoices {
					parts := strings.Split(choice, ":")
					roundNumberStr, triggerGroups := parts[0], parts[1]
					// 将字符串轮次号转换为整数
					roundNumber, err := strconv.Atoi(roundNumberStr)
					if err != nil {
						fmt.Printf("Error converting round number: %v\n", err)
						continue // 如果转换出错，跳过当前循环
					}
					// 检查当前轮次是否符合条件
					if roundNumber == currentRound {
						triggerSets := strings.Split(triggerGroups, "-")
						bestMatchCount := 0
						bestText := ""
						// 遍历每组触发词设置
						for _, triggerSet := range triggerSets {
							triggerDetails := strings.Split(triggerSet, "/")
							addedText := triggerDetails[0] // 提取附加文本
							triggers := triggerDetails[1:] // 提取所有触发词
							matchCount := 0
							// 计算当前消息中包含的触发词数量
							for _, trigger := range triggers {
								if strings.Contains(*requestmsg, trigger) {
									matchCount++
								}
							}
							// 如果当前组的匹配数量最高，更新最佳文本
							if matchCount > bestMatchCount {
								bestMatchCount = matchCount
								bestText = addedText
							}
						}
						// 如果找到了有效的触发词组合，附加最佳文本到消息中
						if bestMatchCount > 0 {
							*requestmsg += " (" + bestText + ")"
							ischange = true
						}
					}
				}
			} else {
				// 当enhancedChoices为false时的处理逻辑
				for _, choice := range promptChoices {
					parts := strings.Split(choice, ":")
					roundNumberStr, addedTexts := parts[0], parts[1]
					roundNumber, err := strconv.Atoi(roundNumberStr)
					if err != nil {
						fmt.Printf("Error converting round number: %v\n", err)
						continue // 如果轮次号转换出错，跳过当前循环
					}
					// 如果当前轮次符合条件，随机选择一个文本添加
					if roundNumber == currentRound {
						texts := strings.Split(addedTexts, "-")
						if len(texts) > 0 {
							selectedText := texts[rand.Intn(len(texts))] // 随机选择一个文本
							*requestmsg += " (" + selectedText + ")"
							ischange = true
						}
					}
				}
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
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

			enhancedChoices := config.GetEnhancedPromptChoices(promptstr)
			if enhancedChoices {
				// 遍历所有的promptChoices配置项
				for _, choice := range promptCover {
					parts := strings.Split(choice, ":")
					roundNumberStr, triggerGroups := parts[0], parts[1]
					// 将字符串轮次号转换为整数
					roundNumber, err := strconv.Atoi(roundNumberStr)
					if err != nil {
						fmt.Printf("Error converting round number: %v\n", err)
						continue // 如果转换出错，跳过当前循环
					}
					// 检查当前轮次是否符合条件
					if roundNumber == currentRound {
						triggerSets := strings.Split(triggerGroups, "-")
						bestMatchCount := 0
						bestText := ""
						// 遍历每组触发词设置
						for _, triggerSet := range triggerSets {
							triggerDetails := strings.Split(triggerSet, "/")
							addedText := triggerDetails[0] // 提取附加文本
							triggers := triggerDetails[1:] // 提取所有触发词
							matchCount := 0
							// 计算当前消息中包含的触发词数量
							for _, trigger := range triggers {
								if strings.Contains(*requestmsg, trigger) {
									matchCount++
								}
							}
							// 如果当前组的匹配数量最高，更新最佳文本
							if matchCount > bestMatchCount {
								bestMatchCount = matchCount
								bestText = addedText
							}
						}
						// 如果找到了有效的触发词组合，覆盖最佳文本到消息中
						if bestMatchCount > 0 {
							*requestmsg = bestText
							ischange = true
						}
					}
				}
			} else {
				// 当enhancedChoices为false时的处理逻辑
				for _, choice := range promptCover {
					parts := strings.Split(choice, ":")
					roundNumberStr, addedTexts := parts[0], parts[1]
					roundNumber, err := strconv.Atoi(roundNumberStr)
					if err != nil {
						fmt.Printf("Error converting round number: %v\n", err)
						continue // 如果轮次号转换出错，跳过当前循环
					}
					// 如果当前轮次符合条件，随机选择一个文本覆盖
					if roundNumber == currentRound {
						texts := strings.Split(addedTexts, "-")
						if len(texts) > 0 {
							selectedText := texts[rand.Intn(len(texts))] // 随机选择一个文本
							*requestmsg = selectedText
							ischange = true
						}
					}
				}
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
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

		enhancedChoices := config.GetEnhancedPromptChoices(*promptstr)
		fmt.Printf("关键词切换分支状态:%v\n", enhancedChoices)
		if enhancedChoices {
			// 遍历所有的promptChoices配置项
			for _, choice := range promptstrChoices {
				parts := strings.Split(choice, ":")
				roundNumberStr, triggerGroups := parts[0], parts[1]
				// 将字符串轮次号转换为整数
				roundNumber, err := strconv.Atoi(roundNumberStr)
				if err != nil {
					fmt.Printf("Error converting round number: %v\n", err)
					continue // 如果转换出错，跳过当前循环
				}
				// 检查当前轮次是否符合条件
				if roundNumber == currentRound {
					triggerSets := strings.Split(triggerGroups, "-")
					bestMatchCount := 0
					bestText := ""
					// 遍历每组触发词设置
					for _, triggerSet := range triggerSets {
						triggerDetails := strings.Split(triggerSet, "/")
						addedText := triggerDetails[0] // 提取附加文本
						triggers := triggerDetails[1:] // 提取所有触发词
						matchCount := 0
						// 计算当前消息中包含的触发词数量
						for _, trigger := range triggers {
							if strings.Contains(*requestmsg, trigger) {
								matchCount++
							}
						}
						// 如果当前组的匹配数量最高，更新最佳文本
						if matchCount > bestMatchCount {
							bestMatchCount = matchCount
							bestText = addedText
						}
					}
					// 如果找到了有效的触发词组合，修改分支
					if bestMatchCount > 0 {
						*promptstr = bestText
						// 获取新的信号长度 刷新持久化数据库
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("enhancedChoices=true,根据关键词切换prompt为: %s,newPromptStrStat:%d\n", *promptstr, PromptMarksLength)
						// 故事模式规则 应用 PromptChoiceQ 这一次是为了,替换了分支后,再次用新的分支的promptstr处理一次,因为原先的promptstr是跳转前,要用跳转后的再替换一次
						app.ApplyPromptChoiceQ(*promptstr, requestmsg, message)
					}
				}
			}
		} else {
			// 当enhancedChoices为false时的处理逻辑
			for _, choice := range promptstrChoices {
				parts := strings.Split(choice, ":")
				roundNumberStr, addedTexts := parts[0], parts[1]
				roundNumber, err := strconv.Atoi(roundNumberStr)
				if err != nil {
					fmt.Printf("Error converting round number: %v\n", err)
					continue // 如果轮次号转换出错，跳过当前循环
				}
				// 如果当前轮次符合条件，随机选择一个分支跳转
				if roundNumber == currentRound {
					texts := strings.Split(addedTexts, "-")
					if len(texts) > 0 {
						selectedText := texts[rand.Intn(len(texts))] // 随机选择一个文本
						*promptstr = selectedText
						// 获取新的信号长度 刷新持久化数据库
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("enhancedChoices=false,根据关键词切换prompt为: %s,newPromptStrStat:%d\n", *promptstr, PromptMarksLength)
						// 故事模式规则 应用 PromptChoiceQ 这一次是为了,替换了分支后,再次用新的分支的promptstr处理一次,因为原先的promptstr是跳转前,要用跳转后的再替换一次
						app.ApplyPromptChoiceQ(*promptstr, requestmsg, message)
					}
				}
			}
		}
	}
}

// ProcessExitChoicesQ 处理配置中的退出选择逻辑，根据特定触发词决定是否触发退出行为。
func (app *App) ProcessExitChoicesQ(promptstr string, requestmsg *string, message *structs.OnebotGroupMessage, selfid string) {
	exitChoices := config.GetExitOnQ(promptstr)
	if len(exitChoices) == 0 {
		return
	}

	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

	enhancedChoices := config.GetEnhancedPromptChoices(promptstr)
	if enhancedChoices {
		for _, choice := range exitChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, triggerGroups := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue // 如果转换出错，跳过当前循环
			}
			if roundNumber == currentRound {
				triggerSets := strings.Split(triggerGroups, "-")
				bestMatchCount := 0
				bestText := ""
				for _, triggerSet := range triggerSets {
					triggerDetails := strings.Split(triggerSet, "/")
					addedText := triggerDetails[0]
					triggers := triggerDetails[1:]
					matchCount := 0
					for _, trigger := range triggers {
						if strings.Contains(*requestmsg, trigger) {
							matchCount++
						}
					}
					if matchCount > bestMatchCount {
						bestMatchCount = matchCount
						bestText = addedText
					}
				}
				if bestMatchCount > 0 {
					app.HandleExit(bestText, message, selfid, promptstr)
					return
				}
			}
		}
	} else {
		for _, choice := range exitChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, addedTexts := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue
			}
			if roundNumber == currentRound {
				texts := strings.Split(addedTexts, "-")
				if len(texts) > 0 {
					selectedText := texts[rand.Intn(len(texts))]
					app.HandleExit(selectedText, message, selfid, promptstr)
					return
				}
			}
		}
	}
}

// HandleExit 处理用户退出逻辑，包括发送消息和重置用户状态。
func (app *App) HandleExit(exitText string, message *structs.OnebotGroupMessage, selfid string, promptstr string) {
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
	}

	fmt.Printf("处理重置操作on:%v", exitText)
	app.migrateUserToNewContext(userid)
	RestoreResponse := config.GetRandomRestoreResponses()
	if message.RealMessageType == "group_private" || message.MessageType == "private" {
		if !config.GetUsePrivateSSE() {
			utils.SendPrivateMessage(message.UserID, RestoreResponse, selfid, promptstr)
		} else {
			utils.SendSSEPrivateRestoreMessage(message.UserID, RestoreResponse, promptstr)
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

	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

	enhancedChoices := config.GetEnhancedPromptChoices(promptstr)
	if enhancedChoices {
		for _, choice := range exitChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, triggerGroups := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue
			}
			if roundNumber == currentRound {
				triggerSets := strings.Split(triggerGroups, "-")
				bestMatchCount := 0
				bestText := ""
				for _, triggerSet := range triggerSets {
					triggerDetails := strings.Split(triggerSet, "/")
					addedText := triggerDetails[0]
					triggers := triggerDetails[1:]
					matchCount := 0
					for _, trigger := range triggers {
						if strings.Contains(*response, trigger) {
							matchCount++
						}
					}
					if matchCount > bestMatchCount {
						bestMatchCount = matchCount
						bestText = addedText
					}
				}
				if bestMatchCount > 0 {
					app.HandleExit(bestText, message, selfid, promptstr)
					return
				}
			}
		}
	} else {
		for _, choice := range exitChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, addedTexts := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue
			}
			if roundNumber == currentRound {
				texts := strings.Split(addedTexts, "-")
				if len(texts) > 0 {
					selectedText := texts[rand.Intn(len(texts))]
					app.HandleExit(selectedText, message, selfid, promptstr)
					return
				}
			}
		}
	}
}

// ApplySwitchOnA 应用switchOnA的逻辑，动态修改promptstr
func (app *App) ApplySwitchOnA(promptstr *string, response *string, message *structs.OnebotGroupMessage) {
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

		enhancedChoices := config.GetEnhancedPromptChoices(*promptstr)
		fmt.Printf("关键词切换分支状态:%v\n", enhancedChoices)
		if enhancedChoices {
			for _, choice := range promptstrChoices {
				parts := strings.Split(choice, ":")
				roundNumberStr, triggerGroups := parts[0], parts[1]
				roundNumber, err := strconv.Atoi(roundNumberStr)
				if err != nil {
					fmt.Printf("Error converting round number: %v\n", err)
					continue // 如果转换出错，跳过当前循环
				}
				if roundNumber == currentRound {
					triggerSets := strings.Split(triggerGroups, "-")
					bestMatchCount := 0
					bestText := ""
					for _, triggerSet := range triggerSets {
						triggerDetails := strings.Split(triggerSet, "/")
						addedText := triggerDetails[0] // 提取附加文本
						triggers := triggerDetails[1:] // 提取所有触发词
						matchCount := 0
						for _, trigger := range triggers {
							if strings.Contains(*response, trigger) {
								matchCount++
							}
						}
						if matchCount > bestMatchCount {
							bestMatchCount = matchCount
							bestText = addedText
						}
					}
					if bestMatchCount > 0 {
						*promptstr = bestText
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("enhancedChoices=true,根据关键词A切换prompt为: %s,newPromptStrStat:%d\n", *promptstr, PromptMarksLength)
					}
				}
			}
		} else {
			for _, choice := range promptstrChoices {
				parts := strings.Split(choice, ":")
				roundNumberStr, addedTexts := parts[0], parts[1]
				roundNumber, err := strconv.Atoi(roundNumberStr)
				if err != nil {
					fmt.Printf("Error converting round number: %v\n", err)
					continue
				}
				if roundNumber == currentRound {
					texts := strings.Split(addedTexts, "-")
					if len(texts) > 0 {
						selectedText := texts[rand.Intn(len(texts))] // 随机选择一个文本
						*promptstr = selectedText
						PromptMarksLength := config.GetPromptMarksLength(*promptstr)
						app.InsertCustomTableRecord(userid, *promptstr, PromptMarksLength)
						fmt.Printf("enhancedChoices=false,根据关键词A切换prompt为: %s,newPromptStrStat:%d\n", *promptstr, PromptMarksLength)
					}
				}
			}
		}
	}
}

// ApplyPromptChoiceA 应用故事模式的情绪增强逻辑，并返回增强内容。
func (app *App) ApplyPromptChoiceA(promptstr string, response string, message *structs.OnebotGroupMessage) string {
	userid := message.UserID
	if config.GetGroupContext() && message.MessageType != "private" {
		userid = message.GroupID
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

	enhancedChoices := config.GetEnhancedPromptChoices(promptstr)
	if enhancedChoices {
		for _, choice := range promptChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, triggerGroups := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue // 如果转换出错，跳过当前循环
			}
			if roundNumber == currentRound {
				triggerSets := strings.Split(triggerGroups, "-")
				bestMatchCount := 0
				bestText := ""
				for _, triggerSet := range triggerSets {
					triggerDetails := strings.Split(triggerSet, "/")
					addedText := triggerDetails[0] // 提取附加文本
					triggers := triggerDetails[1:] // 提取所有触发词
					matchCount := 0
					for _, trigger := range triggers {
						if strings.Contains(response, trigger) {
							matchCount++
						}
					}
					if matchCount > bestMatchCount {
						bestMatchCount = matchCount
						bestText = addedText
					}
				}
				if bestMatchCount > 0 {
					return "(" + bestText + ")"
				}
			}
		}
	} else {
		for _, choice := range promptChoices {
			parts := strings.Split(choice, ":")
			roundNumberStr, addedTexts := parts[0], parts[1]
			roundNumber, err := strconv.Atoi(roundNumberStr)
			if err != nil {
				fmt.Printf("Error converting round number: %v\n", err)
				continue // 如果轮次号转换出错，跳过当前循环
			}
			if roundNumber == currentRound {
				texts := strings.Split(addedTexts, "-")
				if len(texts) > 0 {
					selectedText := texts[rand.Intn(len(texts))] // 随机选择一个文本
					return "(" + selectedText + ")"
				}
			}
		}
	}
	return ""
}
