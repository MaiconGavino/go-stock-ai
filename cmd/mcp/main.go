package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go-stock-ai/internal/simulation"
)


// simularEstoqueInput espelha o mesmo schema usado pela ferramenta exposta
// ao modelo GPT (internal/ai/tools.go)
type simularEstoqueInput struct {
	DemandaMedia         float64 `json:"demanda_media" jsonschema:"média de unidades vendidas por dia"`
	Variacao             float64 `json:"variacao,omitempty" jsonschema:"variação (desvio padrão) aproximada da demanda diária"`
	EstoqueAtual         int     `json:"estoque_atual" jsonschema:"quantidade de unidades atualmente mantida em estoque"`
	NivelServicoDesejado float64 `json:"nivel_servico_desejado,omitempty" jsonschema:"disponibilidade desejada, entre 0 e 1 (ex.: 0.95 para 95%)"`
	Simulacoes           int     `json:"simulacoes,omitempty" jsonschema:"quantidade de cenários a simular"`
}


// novoServidorMCP monta o servidor MCP e registra a ferramenta
// simular_estoque, apoiada no mesmo simulation.Simulator usado pela aplicação web
func novoServidorMCP(sim simulation.Simulator) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "go-stock-ai",
		Version: "v0.1.0",
	}, nil)

	tool := &mcp.Tool{
		Name: "simular_estoque",
		Description: "Simula cenários de demanda diária (Monte Carlo) para " +
			"estimar o risco de falta de um produto e recomendar um nível " +
			"de estoque que atenda a um nível de serviço desejado.",
	}

	mcp.AddTool(server, tool, func(ctx context.Context, _ *mcp.CallToolRequest, in simularEstoqueInput) (*mcp.CallToolResult, simulation.SimulationResult, error) {
		inicio := time.Now()

		entrada := simulation.SimulationInput{
			DemandaMedia:         in.DemandaMedia,
			Variacao:             in.Variacao,
			EstoqueAtual:         in.EstoqueAtual,
			NivelServicoDesejado: in.NivelServicoDesejado,
			Simulacoes:           in.Simulacoes,
		}
		entrada, _ = simulation.ApplyDefaults(entrada)

		if err := simulation.Validate(entrada); err != nil {
			// Erro tratado (ValidationError, mensagem amigável): o SDK
			// converte automaticamente em um resultado de ferramenta com
			// IsError=true, em vez de um erro de protocolo.
			slog.Warn("chamada de simular_estoque rejeitada na validação",
				"duracao_ms", time.Since(inicio).Milliseconds(),
				"erro", err.Error(),
			)
			return nil, simulation.SimulationResult{}, err
		}

		resultado, err := sim.SimularEstoque(ctx, entrada)
		if err != nil {
			slog.Warn("chamada de simular_estoque falhou",
				"duracao_ms", time.Since(inicio).Milliseconds(),
				"erro", err.Error(),
			)
			return nil, simulation.SimulationResult{}, err
		}

		slog.Info("chamada de simular_estoque concluída",
			"duracao_ms", time.Since(inicio).Milliseconds(),
			"simulacoes_executadas", resultado.SimulacoesExecutadas,
			"workers", resultado.Workers,
		)

		return nil, resultado, nil
	})

	return server
}

func main() {
	log.SetOutput(os.Stderr) // stdout é reservado ao protocolo MCP (stdio)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	simulador := simulation.NewConcurrentSimulator(0) // 0 = runtime.NumCPU()
	server := novoServidorMCP(simulador)

	log.Println("go-stock-ai MCP server pronto — aguardando conexão via stdio")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("erro ao executar o servidor MCP: %v", err)
	}
}