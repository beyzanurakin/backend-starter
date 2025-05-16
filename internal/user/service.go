package transaction

var (
	transactions         = map[int]Transaction{}
	transactionIDCounter = 1
	balances             = map[int]float64{}
)

func AddTransaction(t Transaction) Transaction {
	t.ID = transactionIDCounter
	t.CreatedAt = time.Now()
	transactions[t.ID] = t
	transactionIDCounter++
	return t
}

func Credit(userID int, amount float64) Transaction {
	t := Transaction{
		UserID: userID,
		Amount: amount,
		Type:   "credit",
		Status: "completed",
	}
	balances[userID] += amount
	return AddTransaction(t)
}

func Debit(userID int, amount float64) (Transaction, bool) {
	if balances[userID] < amount {
		return Transaction{}, false
	}
	t := Transaction{
		UserID: userID,
		Amount: amount,
		Type:   "debit",
		Status: "completed",
	}
	balances[userID] -= amount
	return AddTransaction(t), true
}

func Transfer(fromUserID, toUserID int, amount float64) (Transaction, bool) {
	if balances[fromUserID] < amount {
		return Transaction{}, false
	}
	balances[fromUserID] -= amount
	balances[toUserID] += amount
	t := Transaction{
		UserID: fromUserID,
		Amount: amount,
		Type:   "transfer",
		Status: "completed",
	}
	return AddTransaction(t), true
}

func GetBalance(userID int) float64 {
	return balances[userID]
}

func GetAllTransactions() []Transaction {
	var list []Transaction
	for _, t := range transactions {
		list = append(list, t)
	}
	return list
}