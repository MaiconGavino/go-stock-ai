package simulation

import "fmt"

const(
	DefaultNivelServico = 0.95
	DefaultSimulacoes = 1000
	
	// MaxSimulacoes é o limite máximo de cenários permitido em uma única requisição
	MaxSimulacoes = 1_000_000
)


// ValidationError representa uma falha de validação com uma mensagem
// amigável, adequada para ser exibida diretamente ao usuário
type ValidationError struct{
	Mensagem string
}

func (e *ValidationError) Error() string{
	return e.Mensagem
}

func validationErrorf(formato string, args ...any) error{
	return &ValidationError{Mensagem: fmt.Sprintf(formato, args...)}
}

// ApplyDefaults preenche valores padrão para campos não informados
// (zerados) e retorna a lista de suposições assumidas em linguagem
// natural, para exibição transparente na interface.
func ApplyDefaults(input SimulationInput) (SimulationInput, []string) {
	var suposicoes []string

	if input.NivelServicoDesejado == 0 {
		input.NivelServicoDesejado = DefaultNivelServico
		suposicoes = append(suposicoes, fmt.Sprintf("nível de serviço padrão: %.0f%%", DefaultNivelServico*100))
	}

	if input.Simulacoes == 0 {
		input.Simulacoes = DefaultSimulacoes
		suposicoes = append(suposicoes, fmt.Sprintf("simulações padrão: %d", DefaultSimulacoes))
	}

	return input, suposicoes
}

// Validate verifica se a entrada é válida para execução da simulação,
// retornando mensagens amigáveis (ValidationError) que podem ser
// apresentadas diretamente ao usuário
func Validate(input SimulationInput) error {
	if input.DemandaMedia <= 0 {
		return validationErrorf("não foi possível executar a simulação porque a demanda média deve ser maior que zero")
	}

	if input.Variacao < 0 {
		return validationErrorf("a variação da demanda não pode ser negativa")
	}

	if input.EstoqueAtual < 0 {
		return validationErrorf("o estoque atual não pode ser negativo")
	}

	if input.NivelServicoDesejado <= 0 || input.NivelServicoDesejado > 1 {
		return validationErrorf("o nível de serviço desejado deve estar entre 0 e 1 (ex.: 0.95 para 95%%)")
	}

	if input.Simulacoes <= 0 {
		return validationErrorf("a quantidade de simulações deve ser maior que zero")
	}

	if input.Simulacoes > MaxSimulacoes {
		return validationErrorf("a quantidade de simulações excede o limite permitido de %d cenários; tente utilizar uma quantidade menor", MaxSimulacoes)
	}

	return nil
}