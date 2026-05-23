# vdl-pygments

Pygments lexer for VDL (Varavel Definition Language).

## Install

```bash
pip install integrations/pygments
# or
pip install vdl-pygments
```

## Usage

After installation, Pygments automatically discovers the lexer:

```python
from pygments import highlight
from pygments.formatters import HtmlFormatter

code = '''
type Person {
  name string
  age? int
}
'''
print(highlight(code, 'vdl', HtmlFormatter()))
```

Or use in code blocks with the `vdl` language:

````md
```vdl
type Foo {
  bar string
}
```
````
