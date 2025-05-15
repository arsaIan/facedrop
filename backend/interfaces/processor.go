package interfaces

type Processor interface {
	Process(eventID uint) (interface{}, error)
}

