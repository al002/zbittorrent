package deferlock

import "sync"

type DeferLock struct {
  mu sync.RWMutex
  deferredFuncs []func()
}

func (df *DeferLock) Lock() {
  df.mu.Lock()
}

func (df *DeferLock) Unlock() {
  funcs := df.deferredFuncs

  for _, f := range funcs {
    f()
  }

  df.deferredFuncs = funcs[:0]
  df.mu.Unlock()
}

func (df *DeferLock) RLock() {
  df.mu.RLock()
}

func (df *DeferLock) RUnlock() {
  df.mu.RUnlock()
}

func (df *DeferLock) AddDefer(f func()) {
  df.deferredFuncs = append(df.deferredFuncs, f)
}
