package ApplicationUseCase

import (
	"time"

	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
)

type ReadMessageReservedBeforeUseCase struct {
	Repository DomainInterfaceRepository.QueueRepository
}

func (useCase *ReadMessageReservedBeforeUseCase) Run(quantity int, maxDate time.Time) (messages []*DomainEntity.MessageEntity) {

	messages, err := useCase.Repository.ReadMessageReservedBefore(quantity, maxDate)

	if err != nil {
		panic(err)
	}

	return messages
}
