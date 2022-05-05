package DomainEntity

import (
	"time"

	DomainTool "github.com/fergkz/lqs/src/Domain/Tool"
)

type MessageEntity struct {
	Queue         string
	DelaySeconds  int
	Attributes    []*AttributeEntity
	Body          string
	ReceiptHandle string
	CreatedAt     time.Time
}

func NewMessage(
	Queue string,
	DelaySeconds int,
	Attributes []*AttributeEntity,
	Body string,
	ReceiptHandle string,
	CreatedAt time.Time,
) (message *MessageEntity) {
	message.Queue = Queue
	message.DelaySeconds = DelaySeconds
	message.Attributes = Attributes
	message.Body = Body
	message.ReceiptHandle = ReceiptHandle
	message.CreatedAt = CreatedAt
	return message
}

func (entity MessageEntity) String() string {
	return DomainTool.Pretty.Prepare(entity)
}

func (entity MessageEntity) Save(file string) {
	DomainTool.Pretty.Save(entity, file)
}

func (entity MessageEntity) Print() {
	DomainTool.Pretty.Println(entity)
}
