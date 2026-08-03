package simulation

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ConcurrentMonteCarloSimulator é uma implementação de Simulator que
// distribui as simulações entre múltiplos workers (goroutines),
// consolidando os resultados parciais recebidos por um channel.

type ConcurrentMonteCarloSimulator struct{
	//workers define quantas gorutines processarão as simulações em paralelo. 
	Workers int 
}

// NewConcurrentSimulator cria um simulador concorrente com a quantidade de
// workers informada. Use 0 para deixar a implementação decidir com base na
// quantidade de CPUs disponíveis (runtime.NumCPU()).
func NewConcurrentSimulator(workers int) *ConcurrentMonteCarloSimulator {
	return &ConcurrentMonteCarloSimulator{Workers: workers}
}

// workerResult carrega o resultado parcial de um worker: as amostras de
// demanda que ele gerou, ou um erro (entrada inválida ou cancelamento).
type workerResult struct{
	amostras []float64
	err error
}

// SimularEstoque divide a quantidade de simulações solicitada entre os
// workers configurados, executa cada bloco em uma goroutine própria e
// consolida os resultados parciais recebidos por channel.
func (s *ConcurrentMonteCarloSimulator) SimularEstoque(ctx context.Context, input SimulationInput) (SimulationResult, error) {
	if err := ctx.Err(); err != nil {
		return SimulationResult{}, err
	}
	if err := Validate(input); err != nil {
		return SimulationResult{}, err
	}

	inicio := time.Now()

	workers := s.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > input.Simulacoes{
		workers = input.Simulacoes
	}
	if workers < 1 {
		workers = 1 
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	blocos := dividirTrabalho(input.Simulacoes, workers)
	resultados := make(chan workerResult, workers)

	seedBase := input.Seed
	if seedBase == 0{
		seedBase = time.Now().UnixNano()
	}

	var wg sync.WaitGroup
	for w, qtd := range blocos {
		wg.Add(1)
		// Cada worker recebe uma seed derivada, mas distinta, para não
		// repetir exatamente a mesma sequência de números aleatórios e,
		// ao mesmo tempo, manter a execução determinística em testes
		// quando input.Seed é diferente de zero.
		go func(qtd int, seedWorker int64){
			defer wg.Done()
			resultados <- executarBloco(ctx, qtd, input, seedWorker)
		}(qtd, seedBase+int64(w))
	}
	go func ()  {
		wg.Wait()
		close(resultados)
	}()

	todasAmostras := make([]float64, 0, input.Simulacoes)
	for r := range resultados {
		if r.err != nil {
			cancel()
			return SimulationResult{}, r.err
		}
		todasAmostras = append(todasAmostras, r.amostras...)
	}
	return consolidarResultado(todasAmostras, input, time.Since(inicio), workers), nil
}
// dividirTrabalho distribui `total` itens entre `workers` blocos da forma
// mais equilibrada possível, alocando o resto entre os primeiros blocos.
func dividirTrabalho(total, workers int) []int {
	blocos := make([]int, workers)
	base := total/workers
	resto := total % workers

	for i := range blocos {
		blocos[i] = base
		if i < resto {
			blocos[i]++
		}
	}
	return blocos
}
// executarBloco gera a fatia de amostras de demanda correspondente a um
// worker, verificando o cancelamento do contexto periodicamente.
func executarBloco(ctx context.Context, qtd int, input SimulationInput, seedWorker int64) workerResult {
	rng := newRNG(seedWorker)
	amostras := make([]float64, 0, qtd)

	for i := 0; i < qtd; i++{
		if i%10000 == 0 {
			if err := ctx.Err(); err != nil{
				return workerResult{err: err}
			}
		}
		amostras = append(amostras, gerarDemanda(rng, input.DemandaMedia, input.Variacao))
	}
	return workerResult{amostras: amostras}

}