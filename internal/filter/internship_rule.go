package filter

import (
	"regexp"
	"strings"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

var (
	// Padrões para detecção de estágio regular
	internshipKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(estagio|estágio|estagiario|estagiário|estagiaria|estagiária)\b`),
		regexp.MustCompile(`(?i)\b(intern|internship)\b`),
	}

	// Padrões para programas formais e processos seletivos de estágio / estudantes
	selectionProcessKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bprograma\s+de\s+(estagio|estágio|talentos|trainee|aceleracao|aceleração)\b`),
		regexp.MustCompile(`(?i)\bprocesso\s+seletivo\s+(?:de\s+)?(estagio|estágio|estudantes|talentos)\b`),
		regexp.MustCompile(`(?i)\b(tech|engineering|software)\s+intern\s+program\b`),
		regexp.MustCompile(`(?i)\bprograma\s+estagi[aá]ri[oa]s?\b`),
		regexp.MustCompile(`(?i)\binscri[cç][oõ]es\s+abertas\s+para\s+est[aá]gio\b`),
	}

	// Padrões de senioridades superiores no título / labels
	seniorKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(senior|sênior|sr\.?|lead|especialista|tech\s+lead|principal)\b`),
	}
	midKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(pleno|pl\.?|mid|mid-level)\b`),
	}
	juniorKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(junior|júnior|jr\.?)\b`),
	}

	// Padrões de regime de contratação
	cltKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(clt|efetivo)\b`),
	}
	pjKeywords = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(pj|pessoa\s+juridica|pessoa\s+jurídica)\b`),
	}
)

// InternshipRule analisa os aspectos de senioridade, tipo de oportunidade e contrato.
type InternshipRule struct {
	cfg config.FilterConfig
}

// NewInternshipRule cria uma nova instância da regra.
func NewInternshipRule(cfg config.FilterConfig) *InternshipRule {
	return &InternshipRule{cfg: cfg}
}

// ClassificationResult reúne as informações extraídas do texto da oportunidade.
type ClassificationResult struct {
	OpportunityType domain.OpportunityType
	Seniority       domain.Seniority
	ContractType    domain.ContractType
	Accepted        bool
	Reason          string
}

// Classify analisa o título e a descrição de uma vaga bruta.
func (r *InternshipRule) Classify(raw domain.RawJob) ClassificationResult {
	textContext := strings.ToLower(raw.Title + " " + raw.LocationRaw + " " + raw.WorkModeRaw)
	fullDesc := strings.ToLower(raw.Description)
	title := strings.ToLower(raw.Title)

	// 1. Detectar senioridade explícita (Senior, Pleno, Junior) prioritariamente
	isSenior := matchesAny(textContext, seniorKeywords)
	isMid := matchesAny(textContext, midKeywords)
	isJunior := matchesAny(textContext, juniorKeywords)

	// 2. Verificar se é Processo Seletivo / Programa de Estágio voltado a estudantes
	isSelectionProcess := false
	if r.cfg.IncludeSelectionPrograms {
		for _, re := range selectionProcessKeywords {
			if re.MatchString(title) || re.MatchString(fullDesc) {
				isSelectionProcess = true
				break
			}
		}
	}

	// 3. Verificar se é Estágio regular
	isInternship := false
	for _, re := range internshipKeywords {
		if re.MatchString(title) || re.MatchString(raw.LocationRaw) || re.MatchString(raw.WorkModeRaw) {
			isInternship = true
			break
		}
	}
	// Se não achou no título/labels, verifica se há menção clara de estágio no corpo,
	// desde que o título não aponte para sênior ou pleno
	if !isInternship && !isSenior && !isMid {
		for _, re := range internshipKeywords {
			if re.MatchString(fullDesc) {
				isInternship = true
				break
			}
		}
	}

	// 4. Se a vaga foi marcada como Sênior ou Pleno no título/labels, ela NÃO é estágio
	if isSenior || isMid {
		isInternship = false
		isSelectionProcess = false
	}

	// 5. Classificar senioridade final
	seniority := domain.SeniorityUnknown
	if isInternship || isSelectionProcess {
		seniority = domain.SeniorityInternship
	} else if isSenior {
		seniority = domain.SenioritySenior
	} else if isMid {
		seniority = domain.SeniorityMid
	} else if isJunior {
		seniority = domain.SeniorityJunior
	}

	// 6. Classificar tipo de oportunidade
	oppType := domain.OpportunityGeneralJob
	if isSelectionProcess {
		oppType = domain.OpportunitySelectionProcess
	} else if isInternship {
		oppType = domain.OpportunityInternship
	}

	// 7. Classificar regime de contratação
	contractType := domain.ContractUnknown
	if isInternship || isSelectionProcess {
		contractType = domain.ContractInternship
	} else if matchesAny(textContext+" "+fullDesc, cltKeywords) {
		contractType = domain.ContractCLT
	} else if matchesAny(textContext+" "+fullDesc, pjKeywords) {
		contractType = domain.ContractPJ
	}

	// 8. Aplicar critérios de aceitação
	if r.cfg.StrictInternshipOnly {
		if oppType == domain.OpportunityInternship || oppType == domain.OpportunitySelectionProcess {
			return ClassificationResult{
				OpportunityType: oppType,
				Seniority:       domain.SeniorityInternship,
				ContractType:    domain.ContractInternship,
				Accepted:        true,
				Reason:          "Oportunidade de estágio / processo seletivo aprovada",
			}
		}
		return ClassificationResult{
			OpportunityType: oppType,
			Seniority:       seniority,
			ContractType:    contractType,
			Accepted:        false,
			Reason:          "Rejeitada: filtro estrito para estágio/processos seletivos ativo",
		}
	}

	// Quando o filtro de estágio é relaxado
	if contractType == domain.ContractCLT && !r.cfg.AllowCLT {
		return ClassificationResult{
			OpportunityType: oppType,
			Seniority:       seniority,
			ContractType:    contractType,
			Accepted:        false,
			Reason:          "Rejeitada: regime CLT não habilitado",
		}
	}

	return ClassificationResult{
		OpportunityType: oppType,
		Seniority:       seniority,
		ContractType:    contractType,
		Accepted:        true,
		Reason:          "Aprovada via filtro abrangente",
	}
}

func matchesAny(text string, regexes []*regexp.Regexp) bool {
	for _, re := range regexes {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
