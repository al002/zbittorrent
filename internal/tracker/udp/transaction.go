package udpTracker

import "math/rand/v2"

func RandomTransactionId() TransactionId {
  return TransactionId(rand.Uint32())
}

type TransactionResponseHandler func(res DispatchedResponse)

type Transaction struct {
  id TransactionId
  tm *TransactionManager
  h TransactionResponseHandler
}

func (t *Transaction) Id() TransactionId {
  return t.id
}

func (t *Transaction) End() {
  t.tm.forgetTransaction(t.id)
}
