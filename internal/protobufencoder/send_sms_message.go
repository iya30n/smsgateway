package protobufencoder

import (
	"smsgateway/contract/sms"
	"smsgateway/entity"

	"google.golang.org/protobuf/proto"
)

func EncodeSendSMSRequest(msg entity.Message) []byte {
	protoMsg := sms.SmsRequest{
		IdempotencyKey: msg.IdempotencyKey,
		MessageId:      int64(msg.ID),
		ReceptorNumber: msg.ReceptorNumber,
		Content:        msg.Content,
	}
	payload, err := proto.Marshal(&protoMsg)
	if err != nil {
		// TODO - log error
		return []byte{}
	}

	return payload
}

func DecodeMatchingUsersMatchedEvent(data []byte) (entity.Message, error) {
	pbMsg := sms.SmsRequest{}
	if err := proto.Unmarshal(data, &pbMsg); err != nil {
		return entity.Message{}, err
	}

	return entity.Message{
		IdempotencyKey: pbMsg.IdempotencyKey,
		ID:             uint(pbMsg.MessageId),
		ReceptorNumber: pbMsg.ReceptorNumber,
		Content:        pbMsg.Content,
	}, nil
}
