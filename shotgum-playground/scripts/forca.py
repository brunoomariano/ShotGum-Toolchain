#!/usr/bin/env python3
"""forca — Jogo da Forca interativo (Python puro, sem dependências externas)."""

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

# ─── ASCII art da forca (7 estágios, 0 a 6 erros) ──────────────────────────────

STAGES = [
    # 0 erros — estrutura vazia
    "  +---+\n  |   |\n      |\n      |\n      |\n      |\n=========",
    # 1 erro — cabeça
    "  +---+\n  |   |\n  O   |\n      |\n      |\n      |\n=========",
    # 2 erros — cabeça + tronco
    "  +---+\n  |   |\n  O   |\n  |   |\n      |\n      |\n=========",
    # 3 erros — braço esquerdo
    "  +---+\n  |   |\n  O   |\n /|   |\n      |\n      |\n=========",
    # 4 erros — ambos os braços
    "  +---+\n  |   |\n  O   |\n /|\\  |\n      |\n      |\n=========",
    # 5 erros — perna esquerda
    "  +---+\n  |   |\n  O   |\n /|\\  |\n /    |\n      |\n=========",
    # 6 erros — completo (game over)
    "  +---+\n  |   |\n  O   |\n /|\\  |\n / \\  |\n      |\n=========",
]

MAX_ERROS = len(STAGES) - 1  # 6

# ─── Lista de palavras (termo, dica) ───────────────────────────────────────────

PALAVRAS = [
    ("PYTHON",         "linguagem de programação criada nos anos 90"),
    ("TERMINAL",       "onde digitamos comandos no computador"),
    ("ALGORITMO",      "sequência de passos para resolver um problema"),
    ("VARIAVEL",       "espaço na memória que guarda um valor"),
    ("FUNCAO",         "bloco de código reutilizável com nome"),
    ("COMPILADOR",     "transforma código-fonte em executável"),
    ("DEPURACAO",      "processo de encontrar e corrigir erros"),
    ("FRAMEWORK",      "estrutura base para desenvolver aplicações"),
    ("PROTOCOLO",      "conjunto de regras para comunicação em rede"),
    ("BINARIO",        "sistema numérico de base dois"),
    ("SERVIDOR",       "máquina que fornece serviços na rede"),
    ("RECURSIVIDADE",  "quando uma função chama a si mesma"),
    ("ABSTRAÇÃO",      "simplificação de conceitos complexos"),
    ("ITERACAO",       "repetição de um bloco de código"),
    ("SINTAXE",        "conjunto de regras de escrita de uma linguagem"),
    ("REPOSITORIO",    "onde o código-fonte fica armazenado"),
    ("INTERFACE",      "ponto de contato entre dois sistemas"),
    ("ESTRUTURA",      "organização de dados na memória"),
]

# ─── Renderização ──────────────────────────────────────────────────────────────

def render(erros, palavra, acertos, erradas, dica):
    clear()
    print(bold(cyan("  ╔══════════════════════════════════╗")))
    print(bold(cyan("  ║    🪤  J O G O  D A  F O R C A   ║")))
    print(bold(cyan("  ╚══════════════════════════════════╝")))
    print()

    # Galhos: vermelho quando game over, amarelo nos outros casos
    for line in STAGES[erros].splitlines():
        if erros == MAX_ERROS:
            print("  " + red(line))
        else:
            print("  " + (yellow(line) if any(c in line for c in "O|/\\") else line))
    print()

    # Palavra com letras reveladas
    display = "  " + "  ".join(
        bold(green(c)) if c in acertos else dim("_")
        for c in palavra
    )
    print(display)
    print()

    # Dica
    print(f"  {dim('Dica:')} {dim(dica)}")
    print()

    # Contadores
    rest = MAX_ERROS - erros
    cor_rest = green if rest > 3 else (yellow if rest > 1 else red)
    print(f"  Erros: {red(str(erros))}/{MAX_ERROS}   "
          f"Restam: {cor_rest(str(rest))} tentativa(s)")
    print()

    # Letras erradas
    if erradas:
        wrong_str = "  ".join(red(l) for l in sorted(erradas))
        print(f"  Erradas: {wrong_str}")
    else:
        print(f"  {dim('(nenhuma letra errada ainda)')}")
    print()

# ─── Loop principal ────────────────────────────────────────────────────────────

def jogar():
    palavra, dica = random.choice(PALAVRAS)
    acertos: set[str] = set()
    erradas: set[str] = set()
    erros = 0

    while erros < MAX_ERROS:
        render(erros, palavra, acertos, erradas, dica)

        if all(c in acertos for c in palavra):
            break  # vitória antecipada

        try:
            entrada = input("  » Digite uma letra: ").strip().upper()
        except (EOFError, KeyboardInterrupt):
            print()
            print(red("\n  Jogo interrompido."))
            sys.exit(0)

        # Validação
        if len(entrada) != 1 or not entrada.isalpha():
            print(yellow("  ⚠  Digite apenas uma única letra."))
            time.sleep(0.8)
            continue

        if entrada in acertos or entrada in erradas:
            print(yellow("  ⚠  Você já tentou essa letra!"))
            time.sleep(0.8)
            continue

        if entrada in palavra:
            acertos.add(entrada)
        else:
            erradas.add(entrada)
            erros += 1

    # Tela final
    render(erros, palavra, acertos, erradas, dica)

    if all(c in acertos for c in palavra):
        print(bold(green("  🎉 Parabéns! Você descobriu a palavra!")))
        print(f"     A palavra era: {bold(cyan(palavra))}")
    else:
        print(bold(red(f"  💀 Que pena! A palavra era: {bold(yellow(palavra))}")))
    print()

    try:
        resp = input("  Jogar novamente? [s/N]: ").strip().lower()
    except (EOFError, KeyboardInterrupt):
        resp = ""
    print()

    if resp == "s":
        jogar()

# ─── Ajuda ─────────────────────────────────────────────────────────────────────

HELP = f"""
{bold(cyan("forca"))} — Jogo da Forca interativo

{bold("Uso:")}
  ./forca.py
  ./forca.py --help

{bold("Como jogar:")}
  Uma palavra secreta é escolhida aleatoriamente.
  Digite uma letra por vez e pressione Enter.
  Você tem {MAX_ERROS} erros antes de perder.
  A dica abaixo da forca ajuda a descobrir a palavra.

{bold("Atalhos:")}
  Ctrl+C   Encerra o jogo

{bold("Requisito:")}
  Terminal interativo — use a tecla {bold(yellow("i"))} no stg para rodar interativamente.
"""

# ─── Entry point ───────────────────────────────────────────────────────────────

def main():
    if "--help" in sys.argv or "-h" in sys.argv:
        print(HELP)
        sys.exit(0)

    if not sys.stdin.isatty():
        print("forca: requer terminal interativo.")
        print("  No stg, pressione 'i' para rodar de forma interativa.")
        sys.exit(1)

    jogar()


if __name__ == "__main__":
    main()
