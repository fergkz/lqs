package InfrastructureRepository

import (
	"database/sql"
	"encoding/json"
	"time"

	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"

	// uuid "github.com/satori/go.uuid"
	// uuid "github.com/satori/go.uuid"

	"fmt"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type QueueRepositoryMySQL struct {
	hostname                           string
	database                           string
	username                           string
	password                           string
	port                               int
	queueName                          string
	fifo                               bool
	queueMessageAlreadyDatabaseCreated bool
}

type queueMessageDTOQueueRepositoryMySQL struct {
	Id         string
	CreatedAt  time.Time
	ReservedAt time.Time
	ReadAfter  time.Time
	Message    DomainEntity.MessageEntity
}

func NewQueueRepositoryMySQL(Hostname string, Database string, Username string, Password string, Port int, QueueName string, Fifo bool) QueueRepositoryMySQL {
	repository := new(QueueRepositoryMySQL)
	repository.hostname = Hostname
	repository.database = Database
	repository.username = Username
	repository.password = Password
	repository.port = Port
	repository.queueName = QueueName
	repository.fifo = Fifo
	repository.queueMessageAlreadyDatabaseCreated = false
	return *repository
}

func (repository QueueRepositoryMySQL) GetQueueName() string {
	return repository.queueName
}

func (repository QueueRepositoryMySQL) getTablename() string {
	return `EVT_` + repository.queueName
}

func (repository *QueueRepositoryMySQL) connect() *sql.DB {

	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowNativePasswords=true&multiStatements=true", repository.username, repository.password, repository.hostname, repository.port, repository.database)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		panic(err)
	}

	if !repository.queueMessageAlreadyDatabaseCreated {
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS ` + repository.getTablename() + ` (
				id VARCHAR(64) PRIMARY KEY,
				created_at DATETIME NOT NULL,
				reserved_at DATETIME NULL,
				reserved_key VARCHAR(64) NULL,
				read_after DATETIME NOT NULL,
				message LONGTEXT
			);
		`)

		if err != nil {
			panic(err)
		}

		repository.queueMessageAlreadyDatabaseCreated = true
	}

	return db
}

func (repository QueueRepositoryMySQL) DropQueue() {
	db := repository.connect()

	sqlStatement, err := db.Prepare(`DROP TABLE IF EXISTS ` + repository.getTablename())

	if err != nil {
		panic(err)
	}

	_, err = sqlStatement.Exec()

	if err != nil {
		panic(err)
	}

	db.Close()

	repository.queueMessageAlreadyDatabaseCreated = false
}

func (repository QueueRepositoryMySQL) SendMessage(messages []*DomainEntity.MessageEntity) error {
	var messagesDTO []queueMessageDTOQueueRepositoryMySQL

	for index := range messages {
		message := (messages)[index]

		message.Queue = repository.queueName

		if message.CreatedAt.IsZero() {
			message.CreatedAt = time.Now()
		}

		message.ReceiptHandle = uuid.New().String()

		messageDTO := queueMessageDTOQueueRepositoryMySQL{
			Id:        message.ReceiptHandle,
			Message:   *message,
			ReadAfter: message.CreatedAt.Add(time.Duration(message.DelaySeconds) * time.Second),
			CreatedAt: message.CreatedAt,
		}

		messagesDTO = append(messagesDTO, messageDTO)
	}

	sqlStr := `INSERT INTO ` + repository.getTablename() + ` (id, created_at, reserved_at, read_after, message) VALUES`
	sqlVals := []interface{}{}

	for _, messageDTO := range messagesDTO {

		sqlStr += `(?, ?, ?, ?, ?),`

		b, _ := json.Marshal(messageDTO.Message)

		sqlVals = append(
			sqlVals,
			messageDTO.Id,
			messageDTO.CreatedAt.Format("2006-01-02 15:04:05"),
			nil,
			messageDTO.ReadAfter.Format("2006-01-02 15:04:05"),
			string(b),
		)

	}

	sqlStr = sqlStr[0 : len(sqlStr)-1]

	db := repository.connect()
	stmt, err := db.Prepare(sqlStr)

	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec(sqlVals...)

	if err != nil {
		panic(err)
	}

	db.Close()

	return nil
}

func (repository QueueRepositoryMySQL) ReadMessage(maxNumberOfMessages int, waitTimeSeconds int) (messages []*DomainEntity.MessageEntity, err error) {
	return repository.readMessageGeneral(maxNumberOfMessages, time.Time{})
}

func (repository QueueRepositoryMySQL) ReadMessageReservedBefore(maxNumberOfMessages int, maxDate time.Time) (messages []*DomainEntity.MessageEntity, err error) {
	return repository.readMessageGeneral(maxNumberOfMessages, maxDate)
}

func (repository QueueRepositoryMySQL) readMessageGeneral(maxNumberOfMessages int, maxDate time.Time) (messages []*DomainEntity.MessageEntity, err error) {
	db := repository.connect()

	strFifo := ""
	if repository.fifo {
		strFifo = "ORDER BY created_at ASC, reserved_at ASC"
	}

	whereReserved := "reserved_at IS NULL"
	if !maxDate.IsZero() {
		whereReserved = "reserved_at IS NOT NULL AND reserved_at < '" + maxDate.Format("2006-01-02 15:04:05") + "'"
	}

	reservedKey := uuid.New().String()

	sqlStatement, err := db.Prepare(`
		UPDATE ` + repository.getTablename() + `
		   SET reserved_at = NOW(),
		       reserved_key = ?
		 WHERE read_after <= NOW()
		   AND ` + whereReserved + `
		` + strFifo + `
		 LIMIT ?
	`)

	if err != nil {
		panic(err)
	}

	_, err = sqlStatement.Exec(reservedKey, maxNumberOfMessages)

	db.Close()

	if err != nil {
		panic(err)
	}

	db = repository.connect()

	res, err := db.Query(`
		SELECT id,
		       created_at,
			   reserved_at,
			   read_after,
			   message
		  FROM `+repository.getTablename()+`
		 WHERE reserved_key = ?
	`, reservedKey)

	if err != nil {
		panic(err)
	}

	db.Close()

	dtos := []queueMessageDTOQueueRepositoryMySQL{}

	for res.Next() {
		dto := queueMessageDTOQueueRepositoryMySQL{}

		var createdAt mysql.NullTime
		var reservedAt mysql.NullTime
		var readAfter mysql.NullTime
		var message sql.NullString

		err := res.Scan(
			&dto.Id,
			&createdAt,
			&reservedAt,
			&readAfter,
			&message,
		)

		if createdAt.Valid {
			dto.CreatedAt = createdAt.Time
		}

		if reservedAt.Valid {
			dto.ReservedAt = reservedAt.Time
		}

		if readAfter.Valid {
			dto.ReadAfter = readAfter.Time
		}

		if message.Valid {
			Message := new(DomainEntity.MessageEntity)
			json.Unmarshal([]byte(message.String), Message)
			dto.Message = *Message
		}

		if err != nil {
			panic(err)
		}

		dtos = append(dtos, dto)
		messages = append(messages, &dto.Message)
	}

	return messages, nil
}

func (repository QueueRepositoryMySQL) DeleteMessage(messages []*DomainEntity.MessageEntity) error {
	for _, message := range messages {
		repository.DeleteMessageByReceiptHandle([]string{message.ReceiptHandle})
	}

	return nil
}

func (repository QueueRepositoryMySQL) DeleteMessageByReceiptHandle(receiptHandles []string) error {
	db := repository.connect()

	sqlStatement, err := db.Prepare(`DELETE FROM ` + repository.getTablename() + ` WHERE id = ?`)

	if err != nil {
		panic(err)
	}

	for _, id := range receiptHandles {
		_, err = sqlStatement.Exec(id)

		if err != nil {
			panic(err)
		}
	}

	db.Close()

	return nil
}

func (repository QueueRepositoryMySQL) CountTotalMessages() int {
	db := repository.connect()

	res, err := db.Query(`SELECT count(0) as cont FROM ` + repository.getTablename())

	if err != nil {
		panic(err)
	}

	db.Close()

	res.Next()

	count := -1

	err = res.Scan(&count)

	if err != nil {
		panic(err)
	}

	return count
}

func (repository QueueRepositoryMySQL) ExportAllToFile(filepath string) {}
