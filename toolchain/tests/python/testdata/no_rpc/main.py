import sys
import os

sys.path.append(os.getcwd())

from gen import *

def main():
    s = Something(field="value")
    if s.field != "value":
        raise Exception("field mismatch")

    # Verify catalog.py does not exist or does not contain RPC definitions
    catalog_path = os.path.join(os.getcwd(), 'gen', 'catalog.py')
    if os.path.exists(catalog_path):
        raise Exception("catalog.py should not exist")

    print("Success")

if __name__ == "__main__":
    main()
