package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/labstack/gommon/log"
	"github.com/nlopes/slack"
	"net/http"
	"net/url"
	"strings"
)

const (
	ActionAcOn      = "ac_on"
	ActionHeaterOn  = "heater_on"
	ActionTurnedOff = "ac_off"
	ActionCancel    = "cancel"
)

const (
	OperationCool = "cool"
	OperationWarm = "warm"
	OperationStop = "power-off"
)

type InteractiveMessageUsecase interface {
	MakeSlackResponse(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

type interactiveMessageUsecase struct {
	signingSecrets string
}

func NewInteractionUsecase(signingSecrets string) InteractiveMessageUsecase {
	return &interactiveMessageUsecase{signingSecrets: signingSecrets}
}

func makeResponse(res *events.APIGatewayProxyResponse, original slack.Message, title, value string) (events.APIGatewayProxyResponse, error) {
	if original.Attachments == nil {
		original.Attachments = []slack.Attachment{slack.Attachment{}}
	}

	original.Text = ""
	original.ResponseType = "in_channel"
	original.ReplaceOriginal = true

	original.Attachments[0].Actions = []slack.AttachmentAction{}
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}
	resJson, err := json.Marshal(&original)
	if err != nil {
		return *res, errors.New("fail to marshal of original message")
	}

	res.Body = string(resJson)
	res.Headers = make(map[string]string)
	res.Headers["Content-type"] = "application/json"
	res.StatusCode = http.StatusOK
	return *res, nil
}

func (i *interactiveMessageUsecase) MakeSlackResponse(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res := events.APIGatewayProxyResponse{}

	str, _ := url.QueryUnescape(req.Body)
	str = strings.Replace(str, "payload=", "", 1)
	//log.Infof("str:%v type:%T", str, str)
	var message slack.InteractionCallback
	// Requestをslack.InteractiveCallback型にParseして利用していく
	if err := json.Unmarshal([]byte(str), &message); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			//Headers:           nil,
			//MultiValueHeaders: nil,
			Body: "json error",
			//IsBase64Encoded:   false,
		}, nil
	}

	if req.HTTPMethod != http.MethodPost {
		res.StatusCode = http.StatusMethodNotAllowed
		return res, errors.New("invalid method")
	}

	if err := i.verify(req); err != nil {
		log.Error(err)
		return res, err
	}

	action := message.ActionCallback.AttachmentActions[0]
	switch action.Name {
	case ActionAcOn:
		title := "OK, TURN ON AN AIR-CONDITIONER!!"
		PutAcSettings(OperationCool)
		return makeResponse(&res, message.OriginalMessage, title, "")
	case ActionHeaterOn:
		title := "OK, TURN ON A HEATER!!"
		PutAcSettings(OperationWarm)
		return makeResponse(&res, message.OriginalMessage, title, "")
	case ActionTurnedOff:
		title := "OK, TURN OFF..."
		PutAcSettings(OperationStop)
		return makeResponse(&res, message.OriginalMessage, title, "")
	case ActionCancel:
		title := fmt.Sprintf(":bb8-flame: MAY THE FORCE BE WITH YOU.")
		return makeResponse(&res, message.OriginalMessage, title, "")

	default:
		res.StatusCode = http.StatusInternalServerError
		return res, errors.New("invalid action was submitted")
	}
}

func (i *interactiveMessageUsecase) verify(request events.APIGatewayProxyRequest) error {
	httpHeader := http.Header{}
	for key, value := range request.Headers {
		httpHeader.Set(key, value)
	}
	sv, err := slack.NewSecretsVerifier(httpHeader, i.signingSecrets)
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
