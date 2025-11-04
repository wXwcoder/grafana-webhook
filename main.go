package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/webhooks/v6/gitlab"
)

// Grafana Alert Data Structures

type GrafanaAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type GrafanaWebhookPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []GrafanaAlert    `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	OrgID             int               `json:"orgId"`
	Title             string            `json:"title"`
	State             string            `json:"state"`
	Message           string            `json:"message"`
}

type TargetType string

const (
	Target_WeiXin TargetType = "weixin"
	Target_FeiShu TargetType = "feishu"
)

var (
	hookPath   = "/webhook"
	hookPort   = 8081
	targetType = Target_FeiShu //weixin, feishu
	robotUrl   = "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
)

// 微信书消息结构
type WechatMsgContent struct {
	//Content gitlab.MergeRequestEventPayload `json:"content"`
	Content string `json:"content"`
}

type WechatMsg_PushEvent struct {
	MsgType string           `json:"msgtype"`
	Text    WechatMsgContent `json:"markdown"`
}

// 飞书消息结构
type FeishuTitleContent struct {
	//Content gitlab.MergeRequestEventPayload `json:"content"`
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

type FeishuTitle struct {
	//Content gitlab.MergeRequestEventPayload `json:"content"`
	Title   string                 `json:"title"`
	Content [][]FeishuTitleContent `json:"content"`
}

type FeishuZh_cn struct {
	Zh_cn FeishuTitle `json:"zh_cn"`
}

type FeishuContent struct {
	Post FeishuZh_cn `json:"post"`
}

type FeishuMsg_MergeRequestEvent struct {
	MsgType string        `json:"msg_type"`
	Content FeishuContent `json:"content"`
}

func gracefulPushMsg(event *gitlab.PushEventPayload) string {
	var msg string
	msg += fmt.Sprintf("项目名称: %v\n", event.Project.Name)

	msg += fmt.Sprintf("仓库地址: [%s]\n", event.Repository.URL, event.Repository.URL)
	msg += fmt.Sprintf("commit: %v\n>", event.CheckoutSHA)
	for _, v := range event.Commits {
		msg += fmt.Sprintf("%s\n", "----------------------")
		msg += fmt.Sprintf("Title: %v\n", v.Title)
		msg += fmt.Sprintf("ID: %v\n", v.ID)
		msg += fmt.Sprintf("Message: %v\n", v.Message)
		msg += fmt.Sprintf("Timestamp: %v\n", v.Timestamp.String())
		msg += fmt.Sprintf("URL: [%s](%s)\n", v.URL, v.URL)
		msg += fmt.Sprintf("Author: %v\n", v.Author.Name)
		for _, str := range v.Added {
			msg += "Added：" + str + "\n"
		}
		for _, str := range v.Modified {
			msg += fmt.Sprintf("Modified: %v\n", str)
		}
		for _, str := range v.Removed {
			msg += fmt.Sprintf("Removed: %v\n", str)
		}
	}
	return msg
}

func gracefulMergeMsg(event *gitlab.MergeRequestEventPayload) string {
	var msg string
	msg += fmt.Sprintf("项目名称: %v\n", event.Project.Name)
	msg += fmt.Sprintf("仓库地址: [%s](%s)\n", event.Repository.URL, event.Repository.URL)
	msg += fmt.Sprintf("Kind: %v\n", event.ObjectKind)
	msg += fmt.Sprintf("User: %v\n", event.User.UserName)
	msg += fmt.Sprintf("%s\n", "----------------------")
	for _, v := range event.Changes.LabelChanges.Previous {
		msg += fmt.Sprintf("Previous ID: %v\n", v.ID)
		msg += fmt.Sprintf("Previous Title: %v\n", v.Title)
		msg += fmt.Sprintf("Previous ProjectID: %v\n", v.ProjectID)
		msg += fmt.Sprintf("Previous CreatedAt: %v\n", v.CreatedAt)
		msg += fmt.Sprintf("Previous UpdatedAt: %v\n", v.UpdatedAt)
		msg += fmt.Sprintf("Previous Description: %v\n", v.Description)
		msg += fmt.Sprintf("Previous Type: %v\n", v.Type)
	}
	msg += fmt.Sprintf("%s\n", "--")
	for _, v := range event.Changes.LabelChanges.Current {
		msg += fmt.Sprintf("Current ID: %v\n", v.ID)
		msg += fmt.Sprintf("Current Title: %v\n", v.Title)
		msg += fmt.Sprintf("Current ProjectID: %v\n", v.ProjectID)
		msg += fmt.Sprintf("Current CreatedAt: %v\n", v.CreatedAt)
		msg += fmt.Sprintf("Current UpdatedAt: %v\n", v.UpdatedAt)
		msg += fmt.Sprintf("Current Description: %v\n", v.Description)
		msg += fmt.Sprintf("Current Type: %v\n", v.Type)
	}
	return msg
}

func main() {

	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println("v1.0.3")
		return
	}

	if len(os.Args) == 1 || (len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "help")) {
		fmt.Println("webhook:")
		fmt.Printf("    %s\n", "hook [push/merge] events of gitlab, and forwarded events to [weixin/feishu]'s robot")
		fmt.Printf("    %s\n", "config webhook url： http://[webhook deploy ip:prot]/webhook")
		fmt.Printf("    %s\n", "config addr： gitlab-> project-> Settings-> Webhooks")
		fmt.Printf("\n")
		fmt.Println("run cmd： ")
		fmt.Printf("    %s\n", "./webhook [port] [\"weixin\"/\"feishu\"] [robotUrl]")
		return
	}

	if len(os.Args) != 4 || os.Args[1] == "" || os.Args[2] == "" || os.Args[3] == "" {
		fmt.Println("please add port and webhook.")
		return
	}
	var err error
	hookPort, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}

	targetType = TargetType(os.Args[2])

	robotUrl = os.Args[3]

	gateIP, err := ExternalIP()
	if err != nil {
		fmt.Println(err)
	}
	//http://192.168.1.47:8081/webhook
	listenAddr := fmt.Sprintf("http://%s:%d%s", gateIP, hookPort, hookPath)

	fmt.Println("hookIP: ", gateIP)
	fmt.Println("hookPort: ", hookPort)
	fmt.Println("hookPath: ", hookPath)
	fmt.Println("target: ", targetType)
	fmt.Println("robotUrl: ", robotUrl)
	fmt.Println("webhook listenAddr: ", listenAddr)

	//hook, _ := gitlab.New()
	http.HandleFunc(hookPath, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic:", r)
			}
		}()
		err, payload := Parse(r)
		if err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		if payload != "" {
			fmt.Println("Grafana Alert payload:", payload)
			payload = strings.Replace(payload, "localhost:3000", "14.103.191.33:13000", -1)
			// Check if it's a Grafana alert by looking at the User-Agent header
			//userAgent := r.Header.Get("User-Agent")
			if true /*strings.Contains(userAgent, "Grafana")*/ {
				var grafanaPayload GrafanaWebhookPayload
				err := json.Unmarshal([]byte(payload), &grafanaPayload)
				if err == nil {
					// It's a Grafana alert, process accordingly
					processGrafanaAlert(grafanaPayload)
				} else {
					fmt.Println("Failed to parse Grafana alert payload:", err)
				}
			} else {
				// Assume it's a GitLab event
				var msg interface{}
				switch targetType {
				case Target_WeiXin:
					msg = &WechatMsg_PushEvent{MsgType: "markdown", Text: WechatMsgContent{Content: payload}}
				case Target_FeiShu:
					msg = &FeishuMsg_MergeRequestEvent{MsgType: "post", Content: FeishuContent{Post: FeishuZh_cn{Zh_cn: FeishuTitle{Title: "", Content: [][]FeishuTitleContent{[]FeishuTitleContent{FeishuTitleContent{Tag: "text", Text: payload}}}}}}}
				}
				var result interface{}
				err = HttpPost(robotUrl, msg, &result, nil)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("result: %+v\n", result)
			}
		}

	})
	http.ListenAndServe(fmt.Sprintf(":%d", hookPort), nil)
}

// New function to process Grafana alerts
func processGrafanaAlert(payload GrafanaWebhookPayload) {
	var msg interface{}
	switch targetType {
	case Target_WeiXin:
		msg = &WechatMsg_PushEvent{MsgType: "markdown", Text: WechatMsgContent{Content: payload.Message}}
	case Target_FeiShu:
		// 格式化飞书告警消息，使其更美观和信息丰富
		var contentLines []FeishuTitleContent

		// 添加标题
		contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: "## " + payload.Title})

		// 添加状态信息
		statusText := "状态: "
		if payload.Status == "firing" {
			statusText += "🚨 告警中"
		} else {
			statusText += "✅ 已恢复"
		}
		contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: statusText})

		// 添加消息内容
		if payload.Message != "" {
			contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: "\n" + payload.Message})
		}

		// 添加告警详情
		if len(payload.Alerts) > 0 {
			contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: "\n### 告警详情:"})
			for i, alert := range payload.Alerts {
				contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: fmt.Sprintf("\n---\n**告警 #%d**", i+1)})

				// 告警状态
				alertStatus := "状态: "
				if alert.Status == "firing" {
					alertStatus += "🚨 触发"
				} else {
					alertStatus += "✅ 解决"
				}
				contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: alertStatus})

				// 告警标签
				if len(alert.Labels) > 0 {
					contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: "标签:"})
					for k, v := range alert.Labels {
						contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: fmt.Sprintf("  - %s: %s", k, v)})
					}
				}

				// 告警注解
				if len(alert.Annotations) > 0 {
					contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: "注解:"})
					for k, v := range alert.Annotations {
						contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: fmt.Sprintf("  - %s: %s", k, v)})
					}
				}

				// 告警时间
				contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: fmt.Sprintf("开始时间: %s", alert.StartsAt.Format("2006-01-02 15:04:05"))})
				if !alert.EndsAt.IsZero() {
					contentLines = append(contentLines, FeishuTitleContent{Tag: "text", Text: fmt.Sprintf("结束时间: %s", alert.EndsAt.Format("2006-01-02 15:04:05"))})
				}
			}
		}

		// 构造飞书消息
		msg = &FeishuMsg_MergeRequestEvent{
			MsgType: "post",
			Content: FeishuContent{
				Post: FeishuZh_cn{
					Zh_cn: FeishuTitle{
						Title:   payload.Title,
						Content: [][]FeishuTitleContent{contentLines},
					},
				},
			},
		}
	}
	var result interface{}
	err := HttpPost(robotUrl, msg, &result, nil)
	if err != nil {
		fmt.Println(err)
	}
	msgStr, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Grafana alert Title: %s, msg: %s, result: %v\n", payload.Title, string(msgStr), result)
}

func Parse(r *http.Request) (error, string) {
	defer func() {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}()

	if r.Method != http.MethodPost {
		return errors.New("Method error"), ""
	}

	/*secret := r.Header.Get("WEBHOOK-SECRET-KEY")
	if secret != "9F25BE4E-4E66-4C4A-AA12-1543F4B14CAE" {
		return errors.New("WEBHOOK-SECRET-TOKEN error"), ""
	}*/

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return err, ""
	}

	return nil, string(payload)
}
