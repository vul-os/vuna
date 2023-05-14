from pygments.lexers import PythonLexer
from pygments import lex

text = """
Here is some text. The following is Python code:

def factorial(n):
\tif n == 0:
\t\treturn 1
\telse:
\t\treturn n * factorial(n-1)

And here is some more text.
"""

tokens = list(lex(text, PythonLexer()))

python_code = ""
in_python_block = False
for ttype, value in tokens:
    if "Comment" in str(ttype):
        continue
    if "Keyword" in str(ttype) or "Indent" in str(ttype) or "Text" in str(ttype):
        in_python_block = True
    if in_python_block:
        python_code += value
    if value == "\n":
        in_python_block = False

print(python_code)
