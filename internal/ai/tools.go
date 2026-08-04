package ai

import (
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go/responses"

	"go-stock-ai/internal/simulation"
)

const ToolNameSimularEstoque = "simular_estoque"

//SimularEstoqueTool descreve, em JSON Schema, os parâmetros que o modelo
// deve extrair da solicitação do usuário para acionar a simulação de estoque

func SimularEstoqueTool() responses.ToolUnionParam {
	parametros := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"demanda_media": map[string]any{
				"type":        "number",
				"description": "Média de unidades vendidas por dia.",
			},
			"variacao": map[string]any{
				"type":        "number",
				"description": "Variação (desvio padrão) aproximada da demanda diária. Se o usuário não mencionar, omita este campo.",
			},
			"estoque_atual": map[string]any{
				"type":        "integer",
				"description": "Quantidade de unidades atualmente mantida em estoque.",
			},
			"nivel_servico_desejado": map[string]any{
				"type":        "number",
				"description": "Disponibilidade desejada, entre 0 e 1 (ex.: 0.95 para 95%). Se o usuário não mencionar, omita este campo.",
			},
			"simulacoes": map[string]any{
				"type":        "integer",
				"description": "Quantidade de cenários a simular. Se o usuário não mencionar, omita este campo.",
			},
		},
		"required":             []string{"demanda_media", "estoque_atual"},
		"additionalProperties": false,
	}

	return responses.ToolParamOfFunction(ToolNameSimularEstoque, parametros, false)
}

// toolArguments espelha o schema aceito pela ferramenta simular_estoque.
type toolArguments struct{
	DemandaMedia         float64 `json:"demanda_media"`
	Variacao             float64 `json:"variacao"`
	EstoqueAtual         int     `json:"estoque_atual"`
	NivelServicoDesejado float64 `json:"nivel_servico_desejado"`
	Simulacoes           int     `json:"simulacoes"`

}

// ParseToolArguments converte o JSON de argumentos devolvido pelo modelo
// (ResponseFunctionToolCall.Arguments) em um simulation.SimulationInput.
func ParseToolArguments(argumentsJSON string) (simulation.SimulationInput, error) {
	var args toolArguments
	if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
		return simulation.SimulationInput{}, fmt.Errorf("não foi possível interpretar os parâmetros gerados pelo modelo: %w", err)
	}

	return simulation.SimulationInput{
		DemandaMedia:         args.DemandaMedia,
		Variacao:             args.Variacao,
		EstoqueAtual:         args.EstoqueAtual,
		NivelServicoDesejado: args.NivelServicoDesejado,
		Simulacoes:           args.Simulacoes,
	}, nil
}