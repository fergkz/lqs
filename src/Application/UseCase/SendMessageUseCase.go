package ApplicationUseCase

import (
	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
)

type SendMessageUseCase struct {
	Repository DomainInterfaceRepository.QueueRepository
}

func (useCase *SendMessageUseCase) Run(messages []*DomainEntity.MessageEntity) []*DomainEntity.MessageEntity {

	err := useCase.Repository.SendMessage(messages)

	if err != nil {
		panic(err)
	}

	return messages
}
