import sys
import os

# Add gen parent directory to path to allow importing gen as a package
sys.path.append(os.path.dirname(__file__))

import gen

if gen.OVERRIDE_WORKS is not True:
    print("OverrideWorks constant is false or missing")
    sys.exit(1)
