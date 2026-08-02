package simulation

import (
	"context"
	"time"
)

type SimulationInput struct{
	// DemandaMedia é a média de unidades vendidas por dia.
	DemandaMedia float64 `json:"demanda_media"`

	// Variacao representa o desvio padrão aproximado da demanda diária.
	Variacao float64 `json:"variacao"`

	// EstoqueAtual é a quantidade de unidades atualmente mantida em estoque.
	EstoqueAtual int `json:"estoque_atual"`

	// NivelServicoDesejado é a disponibilidade desejada, entre 0 e 1
	// (ex.: 0.95 significa estoque suficiente em ~95% dos cenários).
	NivelServicoDesejado float64 `json:"nivel_servico_desejado"`

	// Simulacoes é a quantidade de cenários de demanda a simular.
	Simulacoes int `json:"simulacoes"`

	// Seed, quando diferente de zero, torna a simulação determinística.
	Seed int64 `json:"-"`
}

type SimulationResult struct{
	// EstoqueRecomendado é o menor nível de estoque que alcança o nível de
	// serviço desejado dentro dos cenários simulados.
	EstoqueRecomendado int `json:"estoque_recomendado"`

	// RiscoFaltaEstoqueAtual é a proporção de cenários em que a demanda
	// simulada supera o estoque atual informado.
	RiscoFaltaEstoqueAtual float64 `json:"risco_falta_estoque_atual"`

	// RiscoFaltaEstoqueRecomendado é o risco de falta remanescente após
	// adotar o estoque recomendado.
	RiscoFaltaEstoqueRecomendado float64 `json: "risco_falta_estoque_recomendado"`

	// DemandaMediaSimulada é a média observada nos cenários gerados
	// (deve se aproximar de DemandaMedia, servindo como checagem de
	// sanidade da simulação).
	DemandaMediaSimulada float64 `json:"demanda_media_simulada"`

	// SimulacoesExecutadas é a quantidade de cenários efetivamente
	// processados.
	SimulacoesExecutadas int `json:"simulacoes_executadas"`

	// Duracao é o tempo total de processamento da simulação.
	Duracao time.Duration `json:"duracao"`

	// Workers indica quantos workers concorrentes foram utilizados
	// (1 significa execução sequencial).
	Workers int `json:"workers"`
}

// Simulator define o contrato do motor de simulações.
type Simulator interface {
	SimularEstoque(ctx context.Context, input SimulationInput)(SimulationResult, error)
}