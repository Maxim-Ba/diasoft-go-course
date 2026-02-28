package hw06pipelineexecution

import (
	"context"
	"sync"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	mergedIn := fanInWithFirstDone(in, done)
	out := mergedIn
	for _, stage := range stages {
		out = stage(out)
	}

	final := make(Bi)
	go func() {
		defer close(final)
		for {
			select {
			case <-done:
				// Сигнал остановки – завершаем пересылку.
				return
			case v, ok := <-out:
				if !ok {
					return
				}
				// Если done закрыт во время отправки, выходим.
				select {
				case <-done:
					return
				case final <- v:
				}
			}
		}
	}()

	return final
}

// Суммирует каналы, закрывается когда один из каналов закрывается.
func fanInWithFirstDone(chs ...In) Out {
	sumCh := make(Bi)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// once := &sync.Once{}

	// Подсчитываем только ненулевые каналы для ожидания.
	nonNilChs := 0
	for _, ch := range chs {
		if ch != nil {
			nonNilChs++
		}
	}
	wg.Add(nonNilChs)

	for _, ch := range chs {
		if ch == nil {
			continue // когда done == nil
		}
		go func(ch In) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					// Пытаемся отправить значение в sumCh, но если контекст уже отменён – выходим.
					select {
					case <-ctx.Done():
						return
					case sumCh <- v:
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(sumCh)
		cancel()
	}()

	return sumCh
}
