package DomainInterfaceRepository

import (
	"time"

	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
)

type QueueRepository interface {
	SendMessage(messages []*DomainEntity.MessageEntity) error
	ReadMessage(maxNumberOfMessages int, waitTimeSeconds int) (messages []*DomainEntity.MessageEntity, err error)
	ReadMessageReservedBefore(maxNumberOfMessages int, maxDate time.Time) (messages []*DomainEntity.MessageEntity, err error)
	DeleteMessage(messages []*DomainEntity.MessageEntity) error
	DeleteMessageByReceiptHandle(receiptHandles []string) error
}
