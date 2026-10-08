package protobufencoder

import (
	"encoding/base64"
	"smsgateway/contract/sms"
	"smsgateway/entity"

	"google.golang.org/protobuf/proto"
)

func EncodeSendSMSRequest(msg entity.Message) string {
	protoMsg := sms.RecievedMessage{
		IdempotencyKey:    msg.IdempotencyKey,
		MessageId:         int64(msg.ID),
		ReceptorNumber: msg.ReceptorNumber,
		Content:           msg.Content,
	}
	payload, err := proto.Marshal(&protoMsg)
	if err != nil {
		// TODO - log error
		return ""
	}

	return base64.StdEncoding.EncodeToString(payload)
}

func DecodeMatchingUsersMatchedEvent(data string) entity.Message {
	payload, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		// TODO - log error
		return entity.Message{}
	}

	pbMsg := sms.RecievedMessage{}
	if err := proto.Unmarshal(payload, &pbMsg); err != nil {
		// TODO - log error
		return entity.Message{}
	}

	return entity.Message{
		IdempotencyKey:    pbMsg.IdempotencyKey,
		ID:                  uint(pbMsg.MessageId),
		ReceptorNumber: pbMsg.ReceptorNumber,
		Content:             pbMsg.Content,
	}
}