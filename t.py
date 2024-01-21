import requests
import json

def main():
    url = "http://localhost:46230/conversation"

    # 可以修改以下变量来测试不同的输入
    message = "我第一句话说的什么"
    conversation_id = "07710821-ad06-408c-ba60-1a69bf3ca92a"
    parent_message_id = "73b144d2-a41f-4aeb-b3bb-8624f0e54ba6"
    #conversation_id = "bb8cd44b-672f-4cc7-8a3f-9de0bf55e4c0"
    #parent_message_id = "cd0acb94-52d6-4d08-b648-80267e683245"

    data = {
        "message": message,
        "conversationId": conversation_id,
        "parentMessageId": parent_message_id
    }

    headers = {
        "Content-Type": "application/json"
    }

    response = requests.post(url, data=json.dumps(data), headers=headers)
    
    if response.status_code == 200:
        response_data = response.json()
        print("响应: ", response_data["response"])
        print("conversationId: ", response_data["conversationId"])
        print("messageId: ", response_data["messageId"])
    else:
        print("请求失败，状态码: ", response.status_code)
        print("错误信息: ", response.text)

if __name__ == "__main__":
    main()