package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"

	"go-stock-ai/internal/simulation"
)

// DefaultModel é usado quando nem o chamador nem a variável de ambiente
const DefaultModel = "gpt-4o-mini"

// Client integra a aplicação Go à Responses API da OpenAI
type Client struct {
	api   openai.Client
	Model string
}

// NewClient cria um cliente da API da OpenAI. Quando apiKey ou model são
// strings vazias, os valores são lidos de OPENAI_API_KEY e OPENAI_MODEL
func NewClient(apiKey, model string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		model = DefaultModel
	}

	var opts []option.RequestOption
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}

	return &Client{
		api:   openai.NewClient(opts...),
		Model: model,
	}
}

// AnalysisResult consolida tudo que a interface precisa exibir ao final da análise
type AnalysisResult struct {
	Explicacao          string
	Parametros          simulation.SimulationInput
	Resultado           simulation.SimulationResult
	Etapas              []string
	SuposicoesAssumidas []string
	// FerramentaChamada indica se o modelo decidiu acionar a ferramenta.
	FerramentaChamada bool
}


// Analisar executa o fluxo completo: envia a mensagem do usuário ao
// modelo com a ferramenta simular_estoque disponível, valida e executa a
// simulação quando solicitada pelo modelo, devolve o resultado e obtém a explicação final
func (c *Client) Analisar(ctx context.Context, mensagemUsuario string, sim simulation.Simulator) (AnalysisResult, error) {
	etapas := []string{"Interpretando a solicitação..."}

	primeiraResposta, err := c.api.Responses.New(ctx, responses.ResponseNewParams{
		Model:        c.Model,
		Instructions: param.NewOpt(SystemInstructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.NewOpt(mensagemUsuario),
		},
		Tools: []responses.ToolUnionParam{SimularEstoqueTool()},
	})
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("não foi possível interpretar a solicitação neste momento: %w", err)
	}

	chamada, encontrada := extrairChamadaDeFerramenta(primeiraResposta)
	if !encontrada {
		// O modelo não tinha dados suficientes e devolveu uma pergunta
		// complementar em vez de acionar a ferramenta (seção 20 do doc).
		etapas = append(etapas, "Informações insuficientes para simular.")
		return AnalysisResult{
			Explicacao:        primeiraResposta.OutputText(),
			Etapas:            etapas,
			FerramentaChamada: false,
		}, nil
	}

	etapas = append(etapas, "Parâmetros identificados...")

	entrada, err := ParseToolArguments(chamada.Arguments)
	if err != nil {
		return AnalysisResult{}, err
	}

	entrada, suposicoes := simulation.ApplyDefaults(entrada)

	etapas = append(etapas, "Validando os dados...")
	if err := simulation.Validate(entrada); err != nil {
		return AnalysisResult{}, err
	}

	etapas = append(etapas, fmt.Sprintf("Executando %d simulações...", entrada.Simulacoes))
	resultado, err := sim.SimularEstoque(ctx, entrada)
	if err != nil {
		return AnalysisResult{}, err
	}

	etapas = append(etapas, "Calculando os indicadores...", "Preparando a explicação...")

	resultadoJSON, err := json.Marshal(resultado)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("não foi possível serializar o resultado da simulação: %w", err)
	}

	respostaFinal, err := c.api.Responses.New(ctx, responses.ResponseNewParams{
		Model:              c.Model,
		PreviousResponseID: param.NewOpt(primeiraResposta.ID),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfFunctionCallOutput(chamada.CallID, string(resultadoJSON)),
			},
		},
	})
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("não foi possível preparar a explicação neste momento: %w", err)
	}

	return AnalysisResult{
		Explicacao:          respostaFinal.OutputText(),
		Parametros:          entrada,
		Resultado:           resultado,
		Etapas:              etapas,
		SuposicoesAssumidas: suposicoes,
		FerramentaChamada:   true,
	}, nil
}

func extrairChamadaDeFerramenta(resp *responses.Response) (responses.ResponseFunctionToolCall, bool) {
	for _, item := range resp.Output {
		if item.Type == "function_call" && item.Name == ToolNameSimularEstoque {
			return item.AsFunctionCall(), true
		}
	}
	return responses.ResponseFunctionToolCall{}, false
}

