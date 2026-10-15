package notifier

import (
	"fmt"
	"strings"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// Notifier define o contrato de envio de notificações de novas oportunidades.
type Notifier interface {
	NotifyNewJob(job *domain.Job)
	PrintSummary(stats SummaryStats)
}

// SummaryStats consolida os números de uma rodada de varredura.
type SummaryStats struct {
	TotalSources   int
	FailedSources  int
	TotalEvaluated int
	TotalAccepted  int
	TotalNewSaved  int
}

// TerminalNotifier imprime as oportunidades de forma destacada no terminal com cores ANSI.
type TerminalNotifier struct {
	useColor bool
}

// NewTerminalNotifier cria uma nova instância do notificador de terminal.
func NewTerminalNotifier(useColor bool) *TerminalNotifier {
	return &TerminalNotifier{useColor: useColor}
}

// ANSI Escape Codes
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorGreen   = "\033[32m"
	colorCyan    = "\033[36m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorBlue    = "\033[34m"
	colorGray    = "\033[90m"
)

// NotifyNewJob exibe a oportunidade formatada no terminal.
func (t *TerminalNotifier) NotifyNewJob(job *domain.Job) {
	badge := t.formatBadge(job)

	locInfo := job.Location.Raw
	if job.Location.City != "" {
		locInfo = fmt.Sprintf("%s - %s", job.Location.City, job.Location.State)
	}
	if locInfo == "" {
		locInfo = string(job.WorkMode)
	}

	fmt.Printf("\n%s %s%s%s\n", badge, colorBold, job.Title, colorReset)
	fmt.Printf("   %sEmpresa:%s %s | %sModalidade:%s %s (%s)\n",
		colorGray, colorReset, job.Company,
		colorGray, colorReset, job.WorkMode, locInfo)
	fmt.Printf("   %sOrigem:%s %s | %sData:%s %s\n",
		colorGray, colorReset, job.Source,
		colorGray, colorReset, job.PublishedAt.Format("02/01/2006 15:04"))
	fmt.Printf("   %sLink:%s %s%s%s\n",
		colorGray, colorReset, colorCyan, job.URL, colorReset)
	fmt.Println(strings.Repeat("─", 65))
}

// PrintSummary exibe o relatório de encerramento da varredura.
func (t *TerminalNotifier) PrintSummary(stats SummaryStats) {
	fmt.Printf("\n%s════════════════════ RELATÓRIO DA VARREDURA ════════════════════%s\n", colorBold, colorReset)
	fmt.Printf(" Fontes consultadas : %d (%d falhas)\n", stats.TotalSources, stats.FailedSources)
	fmt.Printf(" Vagas analisadas   : %d\n", stats.TotalEvaluated)
	fmt.Printf(" Vagas aprovadas    : %d\n", stats.TotalAccepted)
	fmt.Printf(" %sNovas vagas salvas  : %d%s\n", colorGreen+colorBold, stats.TotalNewSaved, colorReset)
	fmt.Printf("%s═════════════════════════════════════════════════════════════════%s\n\n", colorBold, colorReset)
}

func (t *TerminalNotifier) formatBadge(job *domain.Job) string {
	switch {
	case job.OpportunityType == domain.OpportunitySelectionProcess:
		return colorYellow + colorBold + "[PROCESSO SELETIVO]" + colorReset

	case job.OpportunityType == domain.OpportunityInternship:
		if job.WorkMode == domain.WorkModeRemote {
			return colorGreen + colorBold + "[ESTÁGIO REMOTO - BRASIL]" + colorReset
		}
		return colorCyan + colorBold + "[ESTÁGIO MT - CUIABÁ/VG]" + colorReset

	default:
		// Vagas aceitas quando o filtro de senioridade/CLT é relaxado
		seniorityLabel := string(job.Seniority)
		if seniorityLabel == "" || seniorityLabel == string(domain.SeniorityUnknown) {
			seniorityLabel = "VAGA GERAL"
		}
		return colorMagenta + colorBold + fmt.Sprintf("[%s - %s]", seniorityLabel, job.WorkMode) + colorReset
	}
}
