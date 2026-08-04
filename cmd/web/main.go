package mcp
// Comando web sobe a aplicação HTTP completa: interpretação em
// linguagem natural (OpenAI), simulação computacional (Go, concorrente)
// e a interface web simples que consome tudo isso.

import (
	"log"
	"net/http"
	"os"
	"time"

	"go-stock-ai/internal/ai"
	"go-stock-ai/internal/service"
	"go-stock-ai/internal/simulation"
	"go-stock-ai/internal/transport"
)

func main() {
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Println("aviso: OPENAI_API_KEY não definida — as requisições à IA falharão até que seja configurada")
	}

	clienteIA := ai.NewClient("", "")
	simulador := simulation.NewConcurrentSimulator(0) // 0 = runtime.NumCPU()
	svc := service.NewService(clienteIA, simulador)
	handler := transport.NewHandler(svc, webDir)

	servidor := &http.Server{
		Addr:              ":" + porta,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("go-stock-ai ouvindo em http://localhost:%s (modelo: %s)", porta, clienteIA.Model)
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatalf("erro ao iniciar o servidor: %v", err)
	}
}