# GPTerm

GPTerm is a CLI tool for querying [OpenAI][openai]'s GPT-3 and Codex models.

## Features

- Completions
  - All GPT-3 models
  - All Codex models
  - Temperature control
  - Max tokens control
- Interactive mode
- Non-Interactive mode
- Prompt composition

### Prompt composition

If you provide gpterm a prompt through more than one mean, it will concatenate them like this:
- Prompt = [Arguments ...] + [Stdin]

```console
$ echo awesome | gpterm one synonim for the word

Fantastic

```
