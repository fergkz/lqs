package ApplicationUseCase

import (
	DomainInterfaceRepository "github.com/fergkz/lqs/src/Domain/Interface/Repository"
)

type RemoveMessageByReceiptHandleUseCase struct {
	Repository DomainInterfaceRepository.QueueRepository
}

func (useCase *RemoveMessageByReceiptHandleUseCase) Run(receiptHandles []string) bool {

	err := useCase.Repository.DeleteMessageByReceiptHandle(receiptHandles)

	if err != nil {
		panic(err)
	}

	return true
}
