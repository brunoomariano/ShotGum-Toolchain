#!/usr/bin/env python3
"""quiz — Quiz de trivia sobre tecnologia e programação."""

import sys
import random
import time

# ─── ANSI helpers ──────────────────────────────────────────────────────────────

def _c(code, text): return f"\033[{code}m{text}\033[0m"

def red(t):    return _c("91", t)
def green(t):  return _c("92", t)
def yellow(t): return _c("93", t)
def cyan(t):   return _c("96", t)
def dim(t):    return _c("2",  t)
def bold(t):   return _c("1",  t)

def clear():
    sys.stdout.write("\033[2J\033[H")
    sys.stdout.flush()

# ─── Banco de perguntas ────────────────────────────────────────────────────────

PERGUNTAS = [
    {
        "q": "Quem criou a linguagem Python?",
        "opts": [
            "Guido van Rossum",
            "Linus Torvalds",
            "Dennis Ritchie",
            "James Gosling",
        ],
        "ans": 0,
        "exp": "Guido van Rossum criou Python no CWI, Países Baixos, lançando a v1 em 1991.",
    },
    {
        "q": "O que significa a sigla HTTP?",
        "opts": [
            "HyperText Transfer Protocol",
            "High-Tech Transmission Process",
            "HyperText Typing Protocol",
            "Hyperlink Transfer Technology",
        ],
        "ans": 0,
        "exp": "HTTP é o protocolo base da Web para troca de dados entre cliente e servidor.",
    },
    {
        "q": "Qual estrutura de dados segue o princípio LIFO?",
        "opts": [
            "Pilha (Stack)",
            "Fila (Queue)",
            "Lista Ligada",
            "Árvore Binária",
        ],
        "ans": 0,
        "exp": "LIFO = Last In, First Out. A pilha insere e remove sempre pelo mesmo extremo.",
    },
    {
        "q": "Em qual ano Linus Torvalds anunciou o kernel Linux?",
        "opts": ["1991", "1985", "1998", "2001"],
        "ans": 0,
        "exp": "Linus anunciou o Linux em agosto de 1991 em uma mensagem para a comp.os.minix.",
    },
    {
        "q": "O que é um algoritmo de ordenação estável?",
        "opts": [
            "Mantém a ordem relativa de elementos iguais",
            "Sempre ordena em O(n log n)",
            "Nunca usa mais que O(1) de memória extra",
            "É determinístico para qualquer entrada",
        ],
        "ans": 0,
        "exp": "Estabilidade: elementos com chaves iguais preservam a ordem original entre si.",
    },
    {
        "q": "Qual protocolo garante entrega confiável de pacotes na internet?",
        "opts": ["TCP", "UDP", "ICMP", "ARP"],
        "ans": 0,
        "exp": "TCP verifica entrega e retransmite pacotes perdidos. UDP não oferece essa garantia.",
    },
    {
        "q": "O que significa 'imutabilidade' em programação funcional?",
        "opts": [
            "Dados não podem ser alterados após a criação",
            "Funções não podem receber argumentos",
            "Variáveis devem ser globais",
            "Loops são proibidos",
        ],
        "ans": 0,
        "exp": "Imutabilidade evita efeitos colaterais: ao invés de alterar, cria-se um novo valor.",
    },
    {
        "q": "Qual das opções NÃO é uma linguagem que compila para binário nativo?",
        "opts": ["Python", "C", "Go", "Rust"],
        "ans": 0,
        "exp": "Python gera bytecode (.pyc) executado pela VM CPython — não binário nativo diretamente.",
    },
    {
        "q": "O que é 'Big O notation'?",
        "opts": [
            "Forma de descrever a complexidade de um algoritmo",
            "Uma linguagem dos anos 70",
            "Um padrão de design de software",
            "Tipo de banco de dados orientado a objetos",
        ],
        "ans": 0,
        "exp": "Big O descreve como tempo/espaço crescem conforme a entrada aumenta.",
    },
    {
        "q": "Em Git, o que faz 'git stash'?",
        "opts": [
            "Salva mudanças temporariamente sem criar commit",
            "Apaga o último commit",
            "Mescla duas branches",
            "Clona um repositório remoto",
        ],
        "ans": 0,
        "exp": "git stash armazena as mudanças em uma pilha temporária, limpando o working tree.",
    },
    {
        "q": "O que é uma API REST?",
        "opts": [
            "Interface baseada em HTTP sem estado entre cliente e servidor",
            "Banco de dados relacional via HTTP",
            "Protocolo de streaming em tempo real",
            "Sistema de arquivos distribuído",
        ],
        "ans": 0,
        "exp": "REST usa HTTP, é stateless e manipula recursos via métodos (GET, POST, PUT, DELETE).",
    },
    {
        "q": "Qual é a complexidade de busca em uma tabela hash (caso médio)?",
        "opts": ["O(1)", "O(log n)", "O(n)", "O(n²)"],
        "ans": 0,
        "exp": "Com boa função de hash e baixo fator de carga, o acesso é O(1) em média.",
    },
]

# ─── Renderização da pergunta ───────────────────────────────────────────────────

LABELS = ["1", "2", "3", "4"]

def show_question(idx: int, total: int, p: dict, answered: bool = False,
                  chosen: int = -1) -> None:
    clear()
    print(bold(cyan("  ╔══════════════════════════════════════╗")))
    print(bold(cyan("  ║   🧠  Q U I Z  T E C H N O L O G Y  ║")))
    print(bold(cyan("  ╚══════════════════════════════════════╝")))
    print()
    print(f"  {dim(f'Pergunta {idx} de {total}')}")
    print()
    print(f"  {bold(p['q'])}")
    print()

    for i, opt in enumerate(p["opts"]):
        label = LABELS[i]
        if answered:
            if i == p["ans"]:
                print(f"  {bold(green(label + '.'))} {green(opt)}")
            elif i == chosen:
                print(f"  {bold(red(label + '.'))} {red(opt)}")
            else:
                print(f"  {dim(label + '.')} {dim(opt)}")
        else:
            print(f"  {bold(cyan(label + '.'))} {opt}")
    print()


def get_choice(total_opts: int) -> int:
    """Lê a opção do usuário; retorna índice 0-based ou -1 se inválido."""
    try:
        raw = input(f"  » Sua resposta (1-{total_opts}): ").strip()
    except (EOFError, KeyboardInterrupt):
        print()
        print(red("\n  Quiz interrompido."))
        sys.exit(0)
    try:
        n = int(raw) - 1
        if 0 <= n < total_opts:
            return n
    except ValueError:
        pass
    return -1

# ─── Loop do quiz ──────────────────────────────────────────────────────────────

def run_quiz() -> None:
    questions = PERGUNTAS[:]
    random.shuffle(questions)
    total = len(questions)
    score = 0
    results: list[tuple[dict, int, bool]] = []

    for idx, p in enumerate(questions, 1):
        show_question(idx, total, p)

        choice = get_choice(len(p["opts"]))

        if choice == -1:
            print(yellow("  ⚠  Opção inválida — contando como erro."))
            time.sleep(1.2)
        else:
            correct = choice == p["ans"]
            if correct:
                score += 1
                show_question(idx, total, p, answered=True, chosen=choice)
                print(bold(green("  ✓ Correto!")))
            else:
                show_question(idx, total, p, answered=True, chosen=choice)
                print(bold(red("  ✗ Errado!")))
                print(f"  {dim(p['exp'])}")

            results.append((p, choice, correct))
            time.sleep(1.5)

    # ─── Tela de resultados ────────────────────────────────────────────────────
    clear()
    print(bold(cyan("  ╔══════════════════════════════════════╗")))
    print(bold(cyan("  ║           R E S U L T A D O          ║")))
    print(bold(cyan("  ╚══════════════════════════════════════╝")))
    print()

    pct = (score / total * 100) if total else 0
    cor = green if pct >= 70 else (yellow if pct >= 40 else red)

    print(f"  Pontuação: {bold(cor(f'{score}/{total}'))}  "
          f"({cor(f'{pct:.0f}%')})")
    print()

    if pct == 100:
        print(bold(green("  🏆 Perfeito! Conhecimento impecável!")))
    elif pct >= 70:
        print(bold(green("  🎉 Muito bem! Continue assim!")))
    elif pct >= 40:
        print(bold(yellow("  📚 Não foi mal, mas dá para melhorar!")))
    else:
        print(bold(red("  😅 Que tal revisar o material?")))
    print()

    # Resumo das respostas
    print(f"  {bold('Revisão:')}")
    print()
    for p, choice, correct in results:
        icon = green("✓") if correct else red("✗")
        q_short = (p["q"][:55] + "…") if len(p["q"]) > 56 else p["q"]
        print(f"  {icon} {dim(q_short)}")
        if not correct:
            correct_opt = p["opts"][p["ans"]]
            chosen_opt  = p["opts"][choice] if 0 <= choice < 4 else "—"
            print(f"     {green('Correta:')} {correct_opt}")
            print(f"     {red('Sua resp:')} {chosen_opt}")
            print(f"     {dim(p['exp'])}")
            print()

    print()
    try:
        resp = input("  Jogar novamente? [s/N]: ").strip().lower()
    except (EOFError, KeyboardInterrupt):
        resp = ""
    if resp == "s":
        run_quiz()

# ─── Ajuda ─────────────────────────────────────────────────────────────────────

HELP = f"""
{bold(cyan("quiz"))} — Quiz de trivia sobre tecnologia e programação

{bold("Uso:")}
  ./quiz.py
  ./quiz.py --help

{bold("Como jogar:")}
  {len(PERGUNTAS)} perguntas de múltipla escolha embaralhadas.
  Digite o número (1-4) da opção correta e pressione Enter.
  Ao final, veja sua pontuação e a revisão das respostas.

{bold("Atalhos:")}
  Ctrl+C   Encerra o quiz

{bold("Requisito:")}
  Terminal interativo — use a tecla {bold(yellow("i"))} no stg para rodar interativamente.
"""

# ─── Entry point ───────────────────────────────────────────────────────────────

def main():
    if "--help" in sys.argv or "-h" in sys.argv:
        print(HELP)
        sys.exit(0)

    if not sys.stdin.isatty():
        print("quiz: requer terminal interativo.")
        print("  No stg, pressione 'i' para rodar de forma interativa.")
        sys.exit(1)

    run_quiz()


if __name__ == "__main__":
    main()
