import os
import sys
from pygments.lexers import LEXERS

sys.path.insert(0, os.path.join(os.getcwd(), "integrations", "pygments"))

LEXERS["VdlLexer"] = ("vdl_lexer", "VDL", ("vdl",), ("*.vdl",), ())

def define_env(env):
    pass
