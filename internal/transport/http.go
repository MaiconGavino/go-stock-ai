package transport

import(
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go-stock-ai/internal/simulation"
	"go-stock-ai/internal/service"
)


// Servicer é o contrato que o serviço de análise precisa satisfazer para ser usado pelo handler HTTP.
type Servicer interface{
	Analisar(ctx context.Context, mensagem string)(service.AnalysisResponse, error)
}

// Handler agrupa as rotas HTTP da aplicação.
type Handler struct {
	svc    Servicer
	webDir string
}

// NewHandler cria um novo handler HTTP. webDir é o diretório contendo os
// arquivos estáticos da interface (index.html, app.js, styles.css).
func NewHandler(svc Servicer, webDir string) *Handler {
	return &Handler{svc: svc, webDir: webDir}
}

// Routes monta o roteador HTTP completo: a API de análise e os arquivos
// estáticos da interface web.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analisar", h.analisar)
	mux.Handle("/", http.FileServer(http.Dir(h.webDir)))
	return mux
}

// analisarRequest é o corpo esperado por POST /api/analisar (seção 11.2
// do documento).
type analisarRequest struct {
	Mensagem string `json:"mensagem"`
}

// erroResponse é o formato de erro devolvido pela API.
type erroResponse struct {
	Erro string `json:"erro"`
}

func (h *Handler) analisar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		escreverErro(w, http.StatusMethodNotAllowed, "Método não permitido; utilize POST.")
		return
	}

	var req analisarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		escreverErro(w, http.StatusBadRequest, "Não foi possível interpretar a requisição enviada.")
		return
	}

	resp, err := h.svc.Analisar(r.Context(), req.Mensagem)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ehValidacao := err.(*simulation.ValidationError); ehValidacao {
			status = http.StatusUnprocessableEntity
		}
		escreverErro(w, status, err.Error())
		return
	}

	escreverJSON(w, http.StatusOK, resp)
}

func escreverJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		log.Printf("erro ao escrever resposta JSON: %v", err)
	}
}

func escreverErro(w http.ResponseWriter, status int, mensagem string) {
	escreverJSON(w, status, erroResponse{Erro: mensagem})
}