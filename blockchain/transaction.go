package blockchain

type Transaction struct {
	ID     []byte
	Input  []TransactionInput
	Output []TransactionOutput
}

type TransactionInput struct {
	ID        []byte
	Output    int
	Signiture string
}

type TransactionOutput struct {
	Value     int
	PublicKey string
}
