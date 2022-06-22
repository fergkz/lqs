package lqs

import (
	"time"

	ApplicationUseCase "github.com/fergkz/lqs/src/Application/UseCase"
	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
	InfrastructureRepository "github.com/fergkz/lqs/src/Infrastructure/Repository"
)

var BuildTagServer bool = false
var BuildTagImport bool = false

type Queue struct {
	repository DomainInterfaceRepository.QueueRepository
}

func Service(StoragePath string, QueueName string, Fifo bool) *Queue {
	app := new(Queue)
	app.repository = InfrastructureRepository.NewQueueRepositoryFiles(StoragePath, QueueName, Fifo)
	return app
}

func ServiceMySQL(Hostname string, Database string, Username string, Password string, Port int, QueueName string, Fifo bool) *Queue {
	app := new(Queue)
	app.repository = InfrastructureRepository.NewQueueRepositoryMySQL(Hostname, Database, Username, Password, Port, QueueName, Fifo)
	return app
}

func (app *Queue) DropQueue() {
	app.repository.DropQueue()
}

func (app *Queue) NewAttributeCollection() []*DomainEntity.AttributeEntity {
	var collection []*DomainEntity.AttributeEntity
	return collection
}

func (app *Queue) NewAttribute(Key string, Value interface{}) *DomainEntity.AttributeEntity {
	attribute := new(DomainEntity.AttributeEntity)
	attribute.Key = Key
	attribute.Value = Value
	return attribute
}

func (app *Queue) NewMessageCollection() []*DomainEntity.MessageEntity {
	var collection []*DomainEntity.MessageEntity
	return collection
}

func (app Queue) NewMessage(
	Body string,
	Attributes []*DomainEntity.AttributeEntity,
	DelaySeconds int,
) *DomainEntity.MessageEntity {
	message := new(DomainEntity.MessageEntity)
	message.Queue = app.repository.GetQueueName()
	message.DelaySeconds = DelaySeconds
	message.Attributes = Attributes
	message.Body = Body
	return message
}

func (app *Queue) SendMessages(messages []*DomainEntity.MessageEntity) []*DomainEntity.MessageEntity {
	(&ApplicationUseCase.SendMessageUseCase{
		Repository: app.repository,
	}).Run(messages)
	return messages
}

func (app *Queue) SendMessage(message *DomainEntity.MessageEntity) *DomainEntity.MessageEntity {
	messages := app.NewMessageCollection()
	messages = append(messages, message)
	(&ApplicationUseCase.SendMessageUseCase{
		Repository: app.repository,
	}).Run(messages)
	return messages[0]
}

func (app *Queue) ReadMessages(quantity int) (messages []*DomainEntity.MessageEntity) {
	messages = (&ApplicationUseCase.ReadMessageUseCase{
		Repository: app.repository,
	}).Run(quantity, 0)
	return messages
}

func (app *Queue) ReadMessagesReservedBefore(quantity int, maxDate time.Time) (messages []*DomainEntity.MessageEntity) {
	messages = (&ApplicationUseCase.ReadMessageReservedBeforeUseCase{
		Repository: app.repository,
	}).Run(quantity, maxDate)
	return messages
}

func (app *Queue) RemoveMessagesByReceiptHandle(receiptHandles []string) bool {
	return (&ApplicationUseCase.RemoveMessageByReceiptHandleUseCase{
		Repository: app.repository,
	}).Run(receiptHandles)
}

func (app *Queue) RemoveMessages(messages []*DomainEntity.MessageEntity) []*DomainEntity.MessageEntity {
	(&ApplicationUseCase.RemoveMessageUseCase{
		Repository: app.repository,
	}).Run(messages)
	return messages
}

func (app *Queue) CountTotalMessages() int {
	return app.repository.CountTotalMessages()
}

func (app *Queue) ExportAllToFile(filepath string) {
	app.repository.ExportAllToFile(filepath)
}
