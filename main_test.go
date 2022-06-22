package lqs

import (
	"fmt"
	"testing"
	"time"

	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
)

var mainTestCurrent struct {
	LastMessagesCount int
}

func init() {
	mainTestCurrent.LastMessagesCount = 0
}

func TestSendMessagesFiles(t *testing.T) {

	messages := []string{
		"Message A",
		"Message B",
	}

	for _, message := range messages {
		mOb := new(DomainEntity.MessageEntity)
		mOb.Body = message
		mOb.CreatedAt = time.Now()
		mOb.DelaySeconds = 0
		mOb.Queue = "TEST"

		service := Service("test", "QTEST", false)
		service.SendMessage(mOb)
	}
}

func TestDropQueueMySQL(t *testing.T) {
	serviceMysql := ServiceMySQL("localhost", "lqsServer", "root", "root", 13307, "QTESTEM", false)
	serviceMysql.DropQueue()
	serviceMysql2 := ServiceMySQL("localhost", "lqsServer", "root", "root", 13307, "QTESTEM2", false)
	serviceMysql2.DropQueue()
}
func TestSendMessagesMySQL(t *testing.T) {
	serviceMysql := ServiceMySQL("localhost", "lqsServer", "root", "root", 13307, "QTESTEM", false)
	serviceMysql2 := ServiceMySQL("localhost", "lqsServer", "root", "root", 13307, "QTESTEM2", false)

	for i := 1; i <= 10; i++ {
		msgs := []*DomainEntity.MessageEntity{}

		for j := 1; j <= 10; j++ {
			mOb := new(DomainEntity.MessageEntity)
			mOb.Body = fmt.Sprintf("TEST MESSAGE %d", i)
			mOb.CreatedAt = time.Now()
			mOb.DelaySeconds = 0
			mOb.Queue = "TEST"

			msgs = append(msgs, mOb)

			mainTestCurrent.LastMessagesCount++
		}

		serviceMysql.SendMessages(msgs)
		serviceMysql2.SendMessages(msgs)

		t.Logf("Messages sended: %d", i)
	}

	if serviceMysql.CountTotalMessages() != mainTestCurrent.LastMessagesCount {
		t.Fatalf(`%d messages in queue. Need %d`, serviceMysql.CountTotalMessages(), mainTestCurrent.LastMessagesCount)
	}

	if serviceMysql2.CountTotalMessages() != mainTestCurrent.LastMessagesCount {
		t.Fatalf(`%d messages in queue. Need %d`, serviceMysql2.CountTotalMessages(), mainTestCurrent.LastMessagesCount)
	}

}
func TestReadAndDeleteMessagesMySQL(t *testing.T) {

	serviceMysql := ServiceMySQL("localhost", "lqsServer", "root", "root", 13307, "QTESTEM", false)

	readed := 0

	for {
		messages := serviceMysql.ReadMessages(25)

		t.Logf("Read %d messages", len(messages))

		if len(messages) == 0 {
			break
		}

		readed += len(messages)

		serviceMysql.RemoveMessages(messages)
		t.Logf("Removed %d messages", len(messages))
	}

	if readed != mainTestCurrent.LastMessagesCount {
		t.Fatalf(`%d messages in queue. Need %d`, readed, mainTestCurrent.LastMessagesCount)
	}
}
