package DomainInterfaceRepository

import DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"

type QueueRepository interface {
	SendMessage(messages []*DomainEntity.MessageEntity) error
	ReadMessage(maxNumberOfMessages int, waitTimeSeconds int) (messages []*DomainEntity.MessageEntity, err error)
	DeleteMessage(messages []*DomainEntity.MessageEntity) error
}
