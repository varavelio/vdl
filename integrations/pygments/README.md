# VDL Pygments lexer

Pygments lexer for VDL (Varavel Definition Language).

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
