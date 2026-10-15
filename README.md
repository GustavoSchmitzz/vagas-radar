# 🎯 Job Radar — Monitor de Estágios e Processos Seletivos em Tecnologia

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14%2B-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20Interface--Driven-brightgreen)](#-arquitetura)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

O **Job Radar** é uma aplicação de alta performance escrita em **Go (Golang)** que roda como um serviço de terminal (*daemon*) em segundo plano. Ele monitora continuamente a abertura de **vagas de estágio** e **processos seletivos** na área de Engenharia e Desenvolvimento de Software, aplicando filtros geográficos estritos e deduplicação atômica no **PostgreSQL**.

Projetado especialmente para **estudantes universitários**, o radar elimina ruídos de vagas para profissionais experientes (júnior/pleno/sênior/CLT) por padrão, mas conta com uma flag para flexibilizar a busca caso desejado.

---

## ⚡ Principais Características

* **Foco Exclusivo em Início de Carreira:** Estágios regulares e processos seletivos formais (programas corporativos de estágio, programas de talentos e aceleração tech).
* **Filtro Geográfico Inteligente:**
  * **Remoto (Brasil):** Aceita vagas de qualquer estado brasileiro, analisando o texto para barrar restrições ocultas (ex: *"vaga remota exclusiva para residentes em SP"* é descartada).
  * **Presencial ou Híbrido:** Aceita **exclusivamente** oportunidades sediadas em **Cuiabá - MT** ou **Várzea Grande - MT**. Vagas presenciais em outros estados são rejeitadas.
* **Filtro Flexível Sob Demanda:** Através da flag `-include-all`, você pode incluir outras senioridades (Júnior, Pleno, CLT) quando quiser expandir o radar.
* **Performance e Consumo Ínfimo:**
  * Consumo de memória em repouso: **~10 a 15 MB de RAM**.
  * Consumo de CPU: **0%** durante os intervalos de espera.
  * Binário único estático e compilado (sem dependências de interpretadores ou runtimes externos).
* **Concorrência Segura & Tolerância a Falhas:**
  * Varredura paralela usando *goroutines* leves com semáforo de controle de taxa (*worker pool*).
  * Isolamento de falhas: se uma fonte externa estiver temporariamente offline, as demais continuam normalmente.
* **Persistência Confiável no PostgreSQL:**
  * Deduplicação atômica baseada em hash SHA-256 (`ON CONFLICT (hash) DO UPDATE`).
  * Atualização automática de datas sem gerar registros duplicados.
* **Automação no Linux (`systemd --user`):**
  * Inicia automaticamente junto com o sistema operacional sem precisar de permissões de administrador (`sudo`).

---

## 🏗 Arquitetura

O projeto adota uma arquitetura em camadas orientada a interfaces simples e modulares:

```mermaid
flowchart TD
    CLI["CLI / Daemon Runner"] --> Sched["Scheduler / Engine Concorrente"]
    Sched -->|Goroutines com Pool| Src["JobSource (Interface)"]
    
    subgraph Fontes Integradas (23 Provedores)
        Src --> S1["20 Comunidades GitHub BR (Backend, Frontend, Data, QA, Mobile, etc.)"]
        Src --> S2["Remotive API Oficial (Worldwide / LATAM / Brasil)"]
        Src --> S3["Canais RSS de Tecnologia (WeWorkRemotely, WorkingNomads)"]
    end

    Src -->|Vagas Brutas| Pipe["Pipeline de Triagem (Engine)"]

    subgraph Validação de Negócio
        Pipe --> R1{"É Estágio / Processo Seletivo?"}
        R1 -- Não --> Disc["Descartada"]
        R1 -- Sim --> R2{"Localização Compatível?"}
        R2 -- Remoto Brasil s/ Restrição --> Appr["Aprovada"]
        R2 -- Presencial/Híbrido Cuiabá ou VG --> Appr
        R2 -- Outras Cidades --> Disc
    end

    Appr --> Hash["Cálculo de Hash SHA-256"]
    Hash --> Repo["Repositório PostgreSQL"]
    Repo -->|Registro Novo| Notif["Terminal Notifier (Badges Coloridos)"]
```

---

## 📡 Catálogo de Fontes Integradas (23 Fontes Ativas)

O Job Radar monitora ativamente as principais comunidades abertas e plataformas de TI:

| Categoria | Fontes Conectadas | Foco Tecnológico |
| :--- | :--- | :--- |
| **Backend & Sistemas** | `backend-br/vagas`, `soujava/vagas-java`, `dotnetdevbr/vagas`, `nodejsdevbr/vagas`, `clj-br/vagas`, `phpdevbr/vagas` | Go, Java/Spring, .NET, Node, Clojure, PHP |
| **Frontend & Web** | `frontendbr/vagas`, `react-brasil/vagas`, `vuejs-br/vagas` | React, Vue, TypeScript, Next.js, Web |
| **Mobile** | `flutterbr/vagas`, `androiddevbr/vagas` | Flutter/Dart, Android Nativo/Kotlin |
| **Dados & IA** | `datascience-br/vagas` | Ciência de Dados, Engenharia de Dados, Python |
| **Qualidade & DevOps** | `qa-brasil/vagas`, `DevOps-Brasil/Vagas` | Testes automatizados, QA, Cloud, Docker, CI/CD |
| **Design, Produto & Web3** | `uxbrasil/vagas`, `remotejobsbr/design-ux-vagas`, `web3br/vagas` | UI/UX, Design de Produto, Web3 |
| **Comunidades Gerais BR** | `techmagiccube/vagas`, `VagasBR/desenvolvedores`, `developersRJ/vagas` | Hubs amplos com centenas de oportunidades |
| **APIs e Feeds Remotos** | `Remotive API`, `WeWorkRemotely RSS`, `WorkingNomads RSS` | Vagas remotas elegíveis para o Brasil e LATAM |

---

## 📂 Estrutura de Diretórios

```
.
├── cmd/
│   └── job-radar/
│       └── main.go              # Ponto de entrada da CLI
├── internal/
│   ├── config/                  # Carregamento de variáveis e flags
│   ├── domain/                  # Entidades puras (Job, Location, Seniority, etc.)
│   ├── filter/                  # Pipeline de regras de negócio (Estágio & Localização)
│   ├── notifier/                # Emissão de alertas formatados no terminal
│   ├── runner/                  # Orquestrador concorrente de fontes e pipeline
│   ├── source/                  # Implementações de coleta de dados (JobSource)
│   └── storage/                 # Camada de persistência (PostgreSQL com pgxpool)
├── scripts/
│   ├── install-service.sh       # Script de registro no systemd --user
│   ├── job-radar.service        # Definição do serviço Linux de segundo plano
│   └── setup-db.sh              # Script auxiliar para criação do banco
├── .env.example                 # Modelo de configuração de ambiente
├── .gitignore                   # Proteção contra commit de binários e senhas
├── go.mod                       # Módulo Go (github.com/GustavoSchmitzz/vagas-radar)
└── go.sum                       # Checksums de dependências
```

---

## 🚀 Como Executar em Qualquer Máquina

### 1. Pré-requisitos
* **Go (Golang)**: versão `1.22` ou superior ([Instalar Go](https://go.dev/doc/install)).
* **PostgreSQL**: versão `14` ou superior ([Instalar PostgreSQL](https://www.postgresql.org/download/)).

---

### 2. Clonando o Repositório
```bash
git clone https://github.com/GustavoSchmitzz/vagas-radar.git
cd vagas-radar
```

---

### 3. Configurando o Banco de Dados
Certifique-se de que o PostgreSQL está em execução na sua máquina. Crie o banco de dados `vagas_radar`:

**Via script automatizado:**
```bash
./scripts/setup-db.sh
```

**Ou manualmente via terminal SQL (`psql`):**
```sql
CREATE DATABASE vagas_radar;
```

---

### 4. Configurando Variáveis de Ambiente
Copie o modelo de exemplo para criar seu `.env`:
```bash
cp .env.example .env
```

Edite o arquivo `.env` com os dados do seu PostgreSQL local (se necessário):
```ini
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=sua_senha_aqui
POSTGRES_DB=vagas_radar
STRICT_INTERNSHIP_ONLY=true
SCAN_INTERVAL=2h
```

---

### 5. Compilando o Binário
Execute a compilação otimizada:
```bash
go build -ldflags="-s -w" -o job-radar ./cmd/job-radar
```
> O binário estático executável `./job-radar` será gerado na raiz da pasta.

---

## 💻 Modos de Uso

### 1. Varredura Imediata (`-scan`)
Executa uma busca em todas as fontes cadastradas, aplica os filtros, salva as vagas inéditas no PostgreSQL e imprime os resultados:
```bash
./job-radar -scan
```

### 2. Listar Vagas Já Salvas (`-list`)
Consulta e exibe no terminal as últimas vagas salvas no banco de dados:
```bash
./job-radar -list
```
*Para limitar a quantidade exibida (ex: 30 vagas):*
```bash
./job-radar -list -limit 30
```

### 3. Abrir uma Vaga no Navegador (`-open <ID>`)
Abre diretamente a página de inscrição/issue da vaga no navegador:
```bash
./job-radar -open 1
```

### 4. Incluir Vagas de Outras Senioridades / CLT (`-include-all`)
Caso queira relaxar o filtro de estágio para ver também vagas Júnior, Pleno ou CLT:
```bash
./job-radar -scan -include-all
```

### 5. Modo Daemon Contínuo (`-daemon`)
Executa o processo continuamente em segundo plano, rodando uma varredura inicial e repetindo periodicamente:
```bash
./job-radar -daemon -interval 2h
```

---

## ⚙️ Tabela de Flags da CLI

| Flag | Tipo | Padrão | Descrição |
| :--- | :--- | :--- | :--- |
| `-scan` | `bool` | `false` | Dispara uma varredura imediata e encerra. |
| `-list` | `bool` | `false` | Lista as oportunidades salvas no PostgreSQL. |
| `-open` | `int` | `0` | Abre a vaga especificada pelo ID no navegador padrão. |
| `-limit` | `int` | `15` | Quantidade de vagas a exibir com `-list`. |
| `-daemon` | `bool` | `false` | Inicia em modo contínuo (serviço). |
| `-interval`| `string`| `2h` | Intervalo do daemon (ex: `30m`, `1h`, `2h`). |
| `-include-all` | `bool` | `false` | Libera vagas Júnior, Pleno, Sênior e CLT. |
| `-allow-clt` | `bool` | `false` | Permite vagas sob regime CLT. |

---

## 🤖 Automação no Linux: Inicialização no Boot com `systemd --user`

Para que o **Job Radar** inicialize automaticamente sempre que você ligar o computador e fizer login:

1. **Instale o serviço:**
   ```bash
   ./scripts/install-service.sh
   ```

2. **Gerenciamento do Serviço:**
   * **Iniciar o daemon agora:**
     ```bash
     systemctl --user start job-radar.service
     ```
   * **Verificar se está rodando:**
     ```bash
     systemctl --user status job-radar.service
     ```
   * **Acompanhar os logs em tempo real:**
     ```bash
     journalctl --user -u job-radar.service -f
     ```
   * **Pausar o serviço:**
     ```bash
     systemctl --user stop job-radar.service
     ```

---

## 🔌 Como Adicionar uma Nova Fonte de Vagas

Graças à interface aberta `JobSource`, você pode plugar qualquer nova API, feed RSS ou repositório criando apenas um arquivo em `internal/source/`:

```go
package source

import (
    "context"
    "github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

type MinhaNovaFonte struct{}

func (s *MinhaNovaFonte) Name() string {
    return "portal:minha-fonte"
}

func (s *MinhaNovaFonte) Fetch(ctx context.Context) ([]domain.RawJob, error) {
    // 1. Faça sua requisição HTTP
    // 2. Converta os dados para []domain.RawJob
    return []domain.RawJob{
        {
            ExternalID:  "123",
            Title:       "Estágio em Engenharia de Software",
            Company:     "Tech MT",
            URL:         "https://exemplo.com/vaga/123",
            LocationRaw: "Cuiabá, MT",
            WorkModeRaw: "Presencial",
        },
    }, nil
}
```

Em seguida, basta registrá-la em `cmd/job-radar/main.go`:
```go
registry.Register(&source.MinhaNovaFonte{})
```

---

## 🧪 Executando a Bateria de Testes

O projeto conta com 100% de cobertura nos pontos críticos de negócio:

```bash
go test -v ./...
```

---

## 👤 Autor

Desenvolvido por **Gustavo Schmitz**
* GitHub: [@GustavoSchmitzz](https://github.com/GustavoSchmitzz)

---

## 📄 Licença

Este projeto está sob a licença [MIT](LICENSE).
