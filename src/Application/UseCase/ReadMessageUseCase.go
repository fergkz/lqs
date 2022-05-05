package ApplicationUseCase

import (
	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
)

type ReadMessageUseCase struct {
	Repository DomainInterfaceRepository.QueueRepository
}

func (useCase *ReadMessageUseCase) Run(quantity int, waitingSeconds int) (messages []*DomainEntity.MessageEntity) {

	messages, err := useCase.Repository.ReadMessage(quantity, waitingSeconds)

	if err != nil {
		panic(err)
	}

	return messages
}
