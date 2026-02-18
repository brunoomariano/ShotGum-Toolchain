# PRD 04 — Registry: Merge e Resolução

## Visão Geral

`internal/registry` combina config global/local e resolve caminho, help flag e executável efetivos.

## Estruturas

```go
type CategoryEntry struct {
    config.Category
    Source string // "local" | "user"
}

type ScriptEntry struct {
    config.Script
    Source string // "local" | "user"
}
```

## Merge

### Categorias (`GetCategories`)

- adiciona locais primeiro
- depois globais não duplicadas por `name`

### Scripts (`GetScripts(category)`)

- filtra por categoria
- adiciona locais primeiro
- depois globais não duplicados por `name` dentro da categoria

## Lookup

`FindScript(category, name)` varre resultado de `GetScripts(category)` e retorna erro se não encontrar.

## Resolução de Path (`ResolveScriptPath`)

Ordem:

1. `entry.Path` absoluto
2. `category.scripts_path + entry.Path`
3. `scripts_home + category + entry.Path`

## Resolução de Help Flag (`ResolveHelpFlag`)

Ordem:

1. `script.help_flag`
2. `category.help_flag`
3. `config.help_flag`
4. `--help`

## Resolução de Executável (`ResolveExecutable`)

Ordem:

1. `script.executable`
2. `config.default_executable` do source
3. `global.default_executable`
4. `/bin/sh`

## Regras de Negócio

| ID | Regra |
|---|---|
| R-REG-01 | Local sobrepõe global em categorias e scripts |
| R-REG-02 | Deduplicação: categorias por `name`; scripts por `name` dentro da categoria solicitada |
| R-REG-03 | `findCategory` e `configFor` usam o `Source` do entry |
| R-REG-04 | `scriptsHome` tenta source atual e depois global |
| R-REG-05 | Métodos retornam slices vazios quando não há dados |
| R-REG-06 | `FindScript` retorna erro descritivo para não-encontrado |
| R-REG-07 | `ResolveHelpFlag` nunca retorna vazio |
| R-REG-08 | `ResolveExecutable` aplica fallback até `/bin/sh` |
