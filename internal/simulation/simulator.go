package simulation

import(
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// MonteCarloSimulator é a implementação padrão de Simulator. Ele gera
// cenários de demanda diária a partir de uma distribuição normal
// (DemandaMedia, Variacao) e usa esses cenários para estimar o risco de
// falta de estoque e recomendar um nível de estoque que atenda ao nível de
// serviço desejado.
type MonteCarloSimulator struct {}

func NewMonteCarloSimulator() *MonteCarloSimulator {
	return &MonteCarloSimulator{}
}

// SimularEstoque executa a simulação sequencial: gera as amostras de
// demanda uma a uma, em um único goroutine. É a versão de referência usada
// para comparação com a versão concorrente (ver concurrent.go) em
// benchmarks.
func (s *MonteCarloSimulator) SimularEstoque(ctx context.Context, input SimulationInput) (SimulationResult, error) {
	if err := ctx.Err(); err != nil {
		return SimulationResult{}, err
	}
	if err := validarEntradaBasica(input); err != nil {
		return SimulationResult{}, err
	}

	inicio := time.Now()

	rng := newRNG(input.Seed)
	amostras := make([]float64, 0, input.Simulacoes)

	for i := 0; i < input.Simulacoes; i++ {
		// Permite cancelamento cooperativo em simulações grandes.
		if i%10000 == 0 {
			if err := ctx.Err(); err != nil {
				return SimulationResult{}, err
			}
		}
		amostras = append(amostras, gerarDemanda(rng, input.DemandaMedia, input.Variacao))
	}

	return consolidarResultado(amostras, input, time.Since(inicio), 1), nil
}

// consolidarResultado transforma um conjunto de amostras de demanda em
// indicadores estruturados: risco do estoque atual, estoque recomendado e
// risco remanescente após a recomendação.
func consolidarResultado(amostras []float64, input SimulationInput, duracao time.Duration, workers int) SimulationResult {
	n := len(amostras)

	riscoAtual := riscoDeFalta(amostras, float64(input.EstoqueAtual))
	recomendado := estoqueParaNivelServico(amostras, input.NivelServicoDesejado)
	riscoRecomendado := riscoDeFalta(amostras, float64(recomendado))

	var soma float64
	for _, a := range amostras {
		soma += a
	}
	mediaSimulada := 0.0
	if n > 0 {
		mediaSimulada = soma / float64(n)
	}

	return SimulationResult{
		EstoqueRecomendado:           recomendado,
		RiscoFaltaEstoqueAtual:       riscoAtual,
		RiscoFaltaEstoqueRecomendado: riscoRecomendado,
		DemandaMediaSimulada:         mediaSimulada,
		SimulacoesExecutadas:         n,
		Duracao:                      duracao,
		Workers:                      workers,
	}
}

// estoqueParaNivelServico calcula o menor estoque inteiro que alcança o
// nível de serviço desejado, usando o percentil correspondente da
// distribuição de demanda simulada.
func estoqueParaNivelServico(amostras []float64, nivelServico float64) int {
	if len(amostras) == 0 {
		return 0
	}
	ordenadas := make([]float64, len(amostras))
	copy(ordenadas, amostras)
	sort.Float64s(ordenadas)

	indice := int(math.Ceil(nivelServico*float64(len(ordenadas)))) - 1
	if indice < 0 {
		indice = 0
	}
	if indice >= len(ordenadas) {
		indice = len(ordenadas) - 1
	}

	return int(math.Ceil(ordenadas[indice]))
}

// riscoDeFalta calcula a proporção de amostras em que a demanda simulada
// supera o nível de estoque informado.
func riscoDeFalta(amostras []float64, estoque float64) float64 {
	if len(amostras) == 0 {
		return 0
	}
	faltas := 0
	for _, a := range amostras {
		if a > estoque {
			faltas++
		}
	}
	return float64(faltas) / float64(len(amostras))
}

// gerarDemanda produz uma amostra de demanda diária a partir de uma
// distribuição normal truncada em zero (demanda negativa não faz sentido
// no domínio do problema e é ajustada para zero,
func gerarDemanda(rng *rand.Rand, media, variacao float64) float64 {
	valor := rng.NormFloat64()*variacao + media
	if valor < 0 {
		return 0
	}
	return valor
}
// newRNG cria um gerador de números aleatórios. Quando seed é diferente de zero
func newRNG(seed int64) *rand.Rand {
	if seed != 0 {
		return rand.New(rand.NewSource(seed))
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// validarEntradaBasica é uma checagem mínima usada internamente pelo simulador
func validarEntradaBasica(input SimulationInput) error{
	if input.DemandaMedia <= 0 {
		return fmt.Errorf("Demanda média deve ser maior que zero")
	}
	if input.Simulacoes <= 0 {
		return  fmt.Errorf("Quantidade de simulações deve ser maior que zero")
	}
	return nil
	
}