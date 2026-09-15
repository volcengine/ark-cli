// Install: go get github.com/volcengine/volcengine-go-sdk@latest
//
// 模板说明:
// - model 使用带版本号的合并 ID (模型名-PrimaryVersion);
//   用 `arkcli models get <name> --format json` 查 PrimaryVersion 后拼合
// - $ARK_API_KEY 运行前由环境变量提供, 不要把真实 Key 写进文件
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

func main() {
	client := arkruntime.NewClientWithApiKey(
		// 通过 os.Getenv 从环境变量中获取 ARK_API_KEY
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
	ctx := context.Background()

	resp, err := client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: "doubao-seed-2-0-pro-260215",
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: &responses.ItemInputMessage{
							Role: responses.MessageRole_user,
							Content: []*responses.ContentItem{{
								Union: &responses.ContentItem_Text{
									Text: &responses.ContentItemText{
										Type: responses.ContentItemType_input_text,
										Text: "Hello!",
									},
								},
							}},
						},
					},
				}}},
			},
		},
	})
	if err != nil {
		fmt.Printf("response error: %v\n", err)
		return
	}
	fmt.Println(resp)
}
