package simulation

import(
	"context"
	"testing"
)

func TestSimularEstoque_Deterministico(t *testing.T){
	sim := NewMonteCarloSimulator()

	input := SimulationInput{
		DemandaMedia: 100,
		Variacao: 20,
		EstoqueAtual: 110,
		NivelServicoDesejado: 0.95,
		Simulacoes: 10000,
		Seed: 42,
	}
	resultado, err := sim.SimularEstoque(context.Background(), input)
	if err != nil{
		t.Fatalf("Erro Inesperado: %v", err)
	}

	if resultado.SimulacoesExecutadas != input.Simulacoes{
		t.Errorf("esperanva %d simulações execultadas, obteve %d", input.Simulacoes, resultado.SimulacoesExecutadas)
	}
	if resultado.EstoqueRecomendado <= input.EstoqueAtual {
		t.Errorf("esperava estoque recomendado maior que o atual (nível de serviço 95%% > risco atual), obteve %d", resultado.EstoqueRecomendado)
	}

	if resultado.RiscoFaltaEstoqueRecomendado > 0.06 {
		t.Errorf("risco após recomendação muito acima do nível de serviço desejado: %.4f", resultado.RiscoFaltaEstoqueRecomendado)
	}

	if resultado.DemandaMediaSimulada < 95 || resultado.DemandaMediaSimulada > 105 {
		t.Errorf("média simulada fora do esperado: %.2f", resultado.DemandaMediaSimulada)
	}
}

func TestSimularEstoque_MesmaSeedMesmoResultado(t *testing.T){
	sim := NewMonteCarloSimulator()

	input := SimulationInput{
		DemandaMedia: 100,
		Variacao: 20,
		EstoqueAtual: 110,
		NivelServicoDesejado: 0.95,
		Simulacoes: 5000,
		Seed: 7,
	}
	r1, err := sim.SimularEstoque(context.Background(), input)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	r2, err := sim.SimularEstoque(context.Background(), input)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if r1.EstoqueRecomendado != r2.EstoqueRecomendado {
		t.Errorf("mesma seed deveria produzir o mesmo resultado: %d != %d", r1.EstoqueRecomendado, r2.EstoqueRecomendado)
	}
	if r1.DemandaMediaSimulada != r2.DemandaMediaSimulada {
		t.Errorf("mesma seed deveria produzir a mesma média simulada: %f != %f", r1.DemandaMediaSimulada, r2.DemandaMediaSimulada)
	}
}

func TestSimularEstoque_DemandaInvalida(t *testing.T){
	sim := NewMonteCarloSimulator()
	input:= SimulationInput{
		DemandaMedia: 0,
		Simulacoes: 100,
	}
	_, err := sim.SimularEstoque(context.Background(), input)
	if err == nil{
		t.Fatal("Esperava erro para demanda média igual zero")
	}
}

func TestSimularEstoque_ContextoCancelador(t *testing.T){
	sim := NewMonteCarloSimulator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := SimulationInput{
		DemandaMedia: 100,
		Variacao: 20,
		Simulacoes: 1000,
	}
	_, err := sim.SimularEstoque(ctx, input)
	if err == nil{
		t.Fatal("Esperaca erro por contexto já cancelado.")
	}
}
