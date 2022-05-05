package ApplicationUseCase

import (
	DomainEntity "github.com/fergkz/lqs/src/Domain/Entity"
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
)

type RemoveMessageUseCase struct {
	Repository DomainInterfaceRepository.QueueRepository
}

func (useCase *RemoveMessageUseCase) Run(messages []*DomainEntity.MessageEntity) []*DomainEntity.MessageEntity {

	err := useCase.Repository.DeleteMessage(messages)

	if err != nil {
		panic(err)
	}

	return messages
}
