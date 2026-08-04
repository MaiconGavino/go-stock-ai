// Package service orquestra a análise de uma solicitação em linguagem
// natural: aplica limites de segurança, chama a IA (que por sua vez aciona
// o motor de simulação)
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"go-stock-ai/internal/ai"
	"go-stock-ai/internal/simulation"
)

// MaxMensagemLength é o tamanho máximo aceito para a mensagem do usuário
const MaxMensagemLength = 2000

// TimeoutAnalise é a duração máxima permitida para uma análise completa
const TimeoutAnalise = 30 * time.Second

type Analisador interface {
	Analisar(ctx context.Context, mensagemUsuario string, sim simulation.Simulator) (ai.AnalysisResult, error)
}

// AnalysisResponse é o formato de saída da API HTTP
type AnalysisResponse struct {
	Resposta            string                       `json:"resposta"`
	Parametros          *simulation.SimulationInput  `json:"parametros,omitempty"`
	Resultado           *simulation.SimulationResult `json:"resultado,omitempty"`
	Etapas              []string                     `json:"etapas"`
	SuposicoesAssumidas []string                     `json:"suposicoes_assumidas,omitempty"`
	FerramentaChamada   bool                         `json:"ferramenta_chamada"`
}

// Service coordena a análise de uma solicitação, combinando o cliente de
// IA com um motor de simulação
type Service struct {
	ai Analisador
	simulador simulation.Simulator
}

// NewService cria um novo serviço de análise.
func NewService(ai Analisador, simulador simulation.Simulator) *Service {
	return &Service{ai: ai, simulador: simulador}
}

// Analisar valida a mensagem do usuário, aplica um tempo limite e delega a
// interpretação ao cliente de IA, traduzindo qualquer falha
func (s *Service) Analisar(ctx context.Context, mensagem string) (AnalysisResponse, error) {
	mensagem = strings.TrimSpace(mensagem)

	if mensagem == "" {
		return AnalysisResponse{}, &simulation.ValidationError{
			Mensagem: "Descreva o cenário do seu estoque para que eu possa analisar.",
		}
	}
	if len(mensagem) > MaxMensagemLength {
		return AnalysisResponse{}, &simulation.ValidationError{
			Mensagem: "A descrição enviada é muito longa. Tente resumir o cenário.",
		}
	}

	ctx, cancel := context.WithTimeout(ctx, TimeoutAnalise)
	defer cancel()

	resultado, err := s.ai.Analisar(ctx, mensagem, s.simulador)
	if err != nil {
		return AnalysisResponse{}, traduzirErro(err)
	}

	resposta := AnalysisResponse{
		Resposta:            resultado.Explicacao,
		Etapas:              resultado.Etapas,
		SuposicoesAssumidas: resultado.SuposicoesAssumidas,
		FerramentaChamada:   resultado.FerramentaChamada,
	}
	if resultado.FerramentaChamada {
		resposta.Parametros = &resultado.Parametros
		resposta.Resultado = &resultado.Resultado
	}

	return resposta, nil
}

// traduzirErro converte erros técnicos (timeout de contexto, falhas de
// rede da API) em mensagens amigáveis. Erros de validação
func traduzirErro(err error) error {
	var ve *simulation.ValidationError
	if errors.As(err, &ve) {
		return ve
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &simulation.ValidationError{
			Mensagem: "A simulação excedeu o tempo permitido. Tente utilizar uma quantidade menor de cenários.",
		}
	}

	return &simulation.ValidationError{
		Mensagem: "Não foi possível interpretar a solicitação neste momento. Você pode preencher os dados manualmente.",
	}
}