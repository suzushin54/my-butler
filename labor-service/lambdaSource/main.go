package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/labstack/gommon/log"
	"github.com/nlopes/slack"
)

const (
	ActionAcOn         = "ac_on"
	ActionHeaterOn     = "heater_on"
	ActionBathHeaterOn = "bath_heater_on"
	ActionTurnedOff    = "ac_off"
	ActionCancel       = "cancel"
)

const (
	UrlVerificationEvent = "url_verification"
	EventCallbackEvent   = "event_callback"
)

const (
	SlackIcon = ":r2d2:"
	SlackName = "R2-D2"
)

// ApiEvent for request parse
type ApiEvent struct {
	Type       string     `json:"type"`
	Text       string     `json:"text"`
	Challenge  string     `json:"challenge"`
	Token      string     `json:"token"`
	SlackEvent SlackEvent `json:"event"`
}

type SlackEvent struct {
	User    string `json:"user"`
	Type    string `json:"type"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

func main() {
	lambda.Start(laborServiceHandler)
}

func laborServiceHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	values, _ := url.ParseQuery(request.Body)
	log.Infof("values:%v type:%T", values, values)

	response := events.APIGatewayProxyResponse{}

	// get environment variables
	channelID := os.Getenv("CHANNEL_ID")
	botOAuth := os.Getenv("BOT_OAUTH")
	botID := os.Getenv("BOT_ID")
	signingSecrets := os.Getenv("SIGNING_SECRETS")

	apiEvent := &ApiEvent{}
	for key, _ := range values {
		// Parse request to apiEvent
		err := json.Unmarshal([]byte(key), apiEvent)
		if err != nil {
			return response, err
		}
	}

	switch apiEvent.Type {
	case UrlVerificationEvent:
		response.Headers = make(map[string]string)
		response.Headers["Content-Type"] = "text/plain"
		response.Body = apiEvent.Challenge
		response.StatusCode = http.StatusOK
		return response, nil
	case EventCallbackEvent:
		slackClient := slack.New(botOAuth)
		slackEvent := apiEvent.SlackEvent

		// input validation
		if slackEvent.Type != "app_mention" {
			return response, errors.New("eventTypeがapp_mentionではありません")
		}

		if !strings.HasPrefix(slackEvent.Text, fmt.Sprintf("<@%s> ", botID)) {
			return response, errors.New("botIDが一致しません")
		}

		if slackEvent.Channel != channelID {
			return response, errors.New("channelIDが一致しません")
		}

		messages := strings.Split(strings.TrimSpace(slackEvent.Text), " ")[1:]
		if len(messages) == 0 || (messages[0] != "hey" && messages[0] != "hi") {
			return response, fmt.Errorf("対応外のメッセージです")
		}

		// 最近まではVerification Tokenによる簡易チェックだったが、deprecatedとなった
		// Signing SecretとRequest Bodyとtimestampを組み合わせてHMAC-SHA256 Hashした署名がRequest Headerに含まれているので計算する
		if err := verify(signingSecrets, request); err != nil {
			log.Error(err)
			return response, err
		}

		// 気温を取得
		temperature := GetTemperature()

		attachment := slack.Attachment{
			Color:      "#f9a41b",
			CallbackID: "server",
			Text:       "BEEP! ROOM TEMPERATURE: " + strconv.FormatFloat(temperature, 'f', 1, 64) + "℃ ..MAY I HELP YOU?",
			Actions: []slack.AttachmentAction{
				{
					Name:  ActionAcOn,
					Text:  "冷房をつけて",
					Type:  "button",
					Style: "primary",
				},
				{
					Name:  ActionHeaterOn,
					Text:  "暖房をつけて",
					Type:  "button",
					Style: "danger",
				},
				{
					Name:  ActionBathHeaterOn,
					Text:  "浴室の暖房をつけて",
					Type:  "button",
					Style: "danger",
				},
				{
					Name:  ActionTurnedOff,
					Text:  "エアコンを消して",
					Type:  "button",
					Style: "default",
				},
				{
					Name:  ActionCancel,
					Text:  "やっぱり大丈夫",
					Type:  "button",
					Style: "default",
				},
			},
		}

		params := slack.PostMessageParameters{
			Username: SlackName,
			//AsUser:          false,
			//Parse:           "",
			//ThreadTimestamp: "",
			//ReplyBroadcast:  false,
			//LinkNames:       0,
			//UnfurlLinks:     false,
			//UnfurlMedia:     false,
			//IconURL:         "",
			IconEmoji: SlackIcon,
			//Markdown:        false,
			//EscapeText:      false,
			//Channel:         "",
			//User:            "",
		}

		msgOptText := slack.MsgOptionText("", true)
		msgOptParams := slack.MsgOptionPostMessageParameters(params)
		msgOptAttachment := slack.MsgOptionAttachments(attachment)

		if _, _, err := slackClient.PostMessage(channelID, msgOptText, msgOptParams, msgOptAttachment); err != nil {
			return response, fmt.Errorf("メッセージ送信に失敗: %s", err)
		}

		response.StatusCode = http.StatusOK
		return response, nil
	default:
		response.StatusCode = http.StatusOK
		return response, nil
	}
}

func verify(signingSecrets string, request events.APIGatewayProxyRequest) error {
	httpHeader := http.Header{}
	for key, value := range request.Headers {
		httpHeader.Set(key, value)
	}
	sv, err := slack.NewSecretsVerifier(httpHeader, signingSecrets)
	if err != nil {
		log.Error(err)
		return err
	}

	if _, err := sv.Write([]byte(request.Body)); err != nil {
		log.Error(err)
		return err
	}

	if err := sv.Ensure(); err != nil {
		log.Error("Invalid SIGNING_SECRETS")
		return err
	}
	return nil
}
