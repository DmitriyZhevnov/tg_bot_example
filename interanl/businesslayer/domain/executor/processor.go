package executor

import "fmt"

type Processor struct {
	// collection
}

func NewProcessor() *Processor {
	return &Processor{
		// collection: collection
	}
}

func (p *Processor) SaveValues(a, b int) error {
	fmt.Println("-------------------")
	fmt.Println("a: ", a, ", b: ", b)
	fmt.Println("-------------------")
	//return p.collection.Save(a, b)

	return nil
}
