const btn = document.getElementById("btn-simular");
const mensagemEl = document.getElementById("mensagem");
const etapasEl = document.getElementById("etapas");
const resultadoEl = document.getElementById("resultado");
const explicacaoEl = document.getElementById("explicacao");
const indicadoresEl = document.getElementById("indicadores");
const tecnicoJsonEl = document.getElementById("tecnico-json");
const erroEl = document.getElementById("erro");

const ROTULOS_INDICADORES = {
	estoque_recomendado: "Estoque recomendado",
	risco_falta_estoque_atual: "Risco atual de falta",
	risco_falta_estoque_recomendado: "Risco estimado após ajuste",
	demanda_media_simulada: "Demanda média simulada",
	simulacoes_executadas: "Cenários analisados",
	workers: "Workers utilizados",
};

function formatarValor(chave, valor) {
	if (chave.startsWith("risco_")) {
		return (valor * 100).toFixed(1) + "%";
	}
	if (typeof valor === "number" && !Number.isInteger(valor)) {
		return valor.toFixed(2);
	}
	return valor;
}

function limparInterface() {
	etapasEl.hidden = true;
	etapasEl.innerHTML = "";
	resultadoEl.hidden = true;
	erroEl.hidden = true;
	erroEl.textContent = "";
}

function exibirEtapas(etapas) {
	if (!etapas || etapas.length === 0) return;
	etapasEl.hidden = false;
	etapasEl.innerHTML = etapas.map((e) => `<div class="etapa">${e}</div>`).join("");
}

function exibirResultado(resposta) {
	resultadoEl.hidden = false;
	explicacaoEl.textContent = resposta.resposta;

	indicadoresEl.innerHTML = "";
	if (resposta.resultado) {
		for (const [chave, rotulo] of Object.entries(ROTULOS_INDICADORES)) {
			if (!(chave in resposta.resultado)) continue;
			const dt = document.createElement("dt");
			dt.textContent = rotulo;
			const dd = document.createElement("dd");
			dd.textContent = formatarValor(chave, resposta.resultado[chave]);
			indicadoresEl.append(dt, dd);
		}
	}

	tecnicoJsonEl.textContent = JSON.stringify(resposta, null, 2);
}

function exibirErro(mensagem) {
	erroEl.hidden = false;
	erroEl.textContent = mensagem;
}

async function simular() {
	const mensagem = mensagemEl.value.trim();
	if (!mensagem) {
		exibirErro("Descreva o cenário do seu estoque antes de simular.");
		return;
	}

	limparInterface();
	btn.disabled = true;
	btn.textContent = "Simulando...";
	exibirEtapas(["Interpretando a solicitação..."]);

	try {
		const resposta = await fetch("/api/analisar", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ mensagem }),
		});

		const dados = await resposta.json();

		if (!resposta.ok) {
			exibirErro(dados.erro || "Não foi possível concluir a análise.");
			return;
		}

		exibirEtapas(dados.etapas);
		exibirResultado(dados);
	} catch (err) {
		exibirErro("Não foi possível se conectar ao servidor. Tente novamente.");
	} finally {
		btn.disabled = false;
		btn.textContent = "Simular cenário";
	}
}

btn.addEventListener("click", simular);