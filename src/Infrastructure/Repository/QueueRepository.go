package InfrastructureRepository

import (
	"encoding/json"
	"log"
	"os"
	"time"

	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainTool "github.com/fergkz/lqs/src/Domain/Tool"
	"github.com/ostafen/clover"
)

type QueueRepository struct {
	StoragePath string
	QueueName   string
	Fifo        bool
}

type queueMessageDTO struct {
	Id         string `json:"_id"`
	ReservedAt time.Time
	ReadAfter  time.Time
	Message    DomainEntity.MessageEntity
}

func (repository *QueueRepository) badgerOpen() (*clover.DB, bool, error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	if err := os.Mkdir(repository.StoragePath, 0777); err != nil && !os.IsExist(err) {
		return nil, false, err
	}

	connect := false
	db, err := clover.Open(repository.StoragePath + "/" + repository.QueueName)

	if err == nil {
		connect = true
	}

	return db, connect, err
}

func (repository *QueueRepository) badgerOpenLoop() (*clover.DB, error) {
	var db *clover.DB
	var err error
	connected := false

	for i := 0; i < 3000; i++ {
		db, connected, err = repository.badgerOpen()
		if connected {
			break
		}
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return db, err
}

func (repository *QueueRepository) connect() *clover.DB {
	db, _ := repository.badgerOpenLoop()

	exists, _ := db.HasCollection(repository.QueueName)
	if !exists {
		db.CreateCollection(repository.QueueName)
	}

	return db
}

func (repository *QueueRepository) collectionToDocuments(collection *[]queueMessageDTO) []*clover.Document {
	documents := make([]*clover.Document, 0, len(*collection))
	for _, item := range *collection {
		var inInterface map[string]interface{}
		inrec, _ := json.Marshal(item)
		json.Unmarshal(inrec, &inInterface)
		documents = append(documents, clover.NewDocumentOf(inInterface))
	}
	return documents
}

func (repository *QueueRepository) SendMessage(messages []*DomainEntity.MessageEntity) error {
	var messagesDTO []queueMessageDTO

	for index := range messages {
		message := (messages)[index]

		message.Queue = repository.QueueName

		if message.CreatedAt.IsZero() {
			message.CreatedAt = time.Now()
		}

		message.ReceiptHandle = clover.NewObjectId()

		messageDTO := queueMessageDTO{
			Message:   *message,
			ReadAfter: message.CreatedAt.Add(time.Duration(message.DelaySeconds) * time.Second),
			Id:        message.ReceiptHandle,
		}

		messagesDTO = append(messagesDTO, messageDTO)
	}

	docs := repository.collectionToDocuments(&messagesDTO)

	if len(docs) > 0 {
		db := repository.connect()
		defer db.Close()
		err := db.Insert(repository.QueueName, docs...)
		db.Close()
		if err != nil {
			log.Fatalln("ERROR:", err)
		}
	}

	return nil
}

func (repository *QueueRepository) ReadMessage(maxNumberOfMessages int, waitTimeSeconds int) (messages []*DomainEntity.MessageEntity, err error) {
	db := repository.connect()
	defer db.Close()

	query := db.Query(repository.QueueName).Where(
		clover.Field("ReservedAt").IsNilOrNotExists().Or(
			clover.Field("ReservedAt").Eq(time.Time{}),
		),
	).Where(
		clover.Field("ReadAfter").LtEq(time.Now()),
	)

	if repository.Fifo {
		query = query.Sort(clover.SortOption{
			Field:     "CreatedAt",
			Direction: 1,
		})
	}

	docs, err := query.Limit(maxNumberOfMessages).FindAll()

	if err != nil {
		DomainTool.Pretty.Fatalln("ERROR ON QueueRepository.ReadMessage.01", err)
	}

	currentTime := time.Now()

	updates := make(map[string]interface{})
	updates["ReservedAt"] = currentTime

	for _, doc := range docs {
		messageDTO := &queueMessageDTO{}
		doc.Unmarshal(messageDTO)
		messages = append(messages, &messageDTO.Message)
		err := db.Query(repository.QueueName).UpdateById(messageDTO.Message.ReceiptHandle, updates)
		if err != nil {
			DomainTool.Pretty.Fatalln("ERROR ON QueueRepository.ReadMessage.02", err)
		}
	}

	db.Close()

	return messages, nil
}

func (repository *QueueRepository) DeleteMessage(messages []*DomainEntity.MessageEntity) error {
	db := repository.connect()
	defer db.Close()
	for _, message := range messages {
		err := db.Query(repository.QueueName).DeleteById(message.ReceiptHandle)
		if err != nil {
			DomainTool.Pretty.Fatalln("ERROR ON QueueRepository.DeleteMessage.01", err)
		}
	}
	db.Close()

	return nil
}

func (repository *QueueRepository) CountTotalMessages() int {
	db := repository.connect()
	defer db.Close()

	count, _ := db.Query(repository.QueueName).Count()

	db.Close()

	return count
}

func (repository *QueueRepository) ExportAllToFile(filepath string) {
	db, _ := repository.badgerOpenLoop()
	defer db.Close()
	docs, _ := db.Query(repository.QueueName).FindAll()
	db.Close()

	var messages []queueMessageDTO
	for _, doc := range docs {
		messageDTO := &queueMessageDTO{}
		doc.Unmarshal(messageDTO)

		messages = append(messages, *messageDTO)
	}

	DomainTool.Pretty.Save(messages, filepath)
}
