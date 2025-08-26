
# Security Policy

Nós levamos segurança a sério. Se você encontrou uma vulnerabilidade, **não** abra uma issue pública.

## Como reportar
- Use **GitHub Security Advisory** (Security > Advisories > Report a vulnerability) para reportar de forma privada, **ou**
- Envie e-mail para: **security@exemplo.com** (substitua pelo seu contato) com:
  - Descrição do problema e impacto
  - Passos para reprodução (PoC, logs, versões)
  - Escopo afetado (endpoints, componentes)

Você receberá uma resposta inicial em até **72 horas**.

## Escopo
- Código fonte deste repositório
- Configurações de CI/CD (GitHub Actions, Dependabot)
- Dependências diretas (Go modules)

## Fora do escopo
- Serviços de terceiros (Stripe, S3 etc.) fora do código deste repo
- Vulnerabilidades causadas por configurações incorretas no ambiente do usuário

## Processo de correção
1. Análise e confirmação da vulnerabilidade.
2. Definição de severidade (CVSS aproximado) e plano de correção.
3. Patch preparado e validado (testes automatizados).
4. **Divulgação responsável**: publicação do advisory e release contendo o fix.

## Boas práticas recomendadas ao usar este projeto
- Use **tokens e segredos** somente via **GitHub Secrets** / variáveis de ambiente.
- Habilite **branch protection** e **CodeQL** no GitHub.
- Rode `go test ./...` e scanners de dependência (Dependabot) regularmente.
