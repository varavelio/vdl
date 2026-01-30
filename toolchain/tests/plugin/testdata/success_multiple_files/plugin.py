#!/usr/bin/env python3
"""
Plugin that generates multiple files to test multi-file output handling.
"""
import json
import sys

def main():
    input_data = json.load(sys.stdin)
    ir = input_data.get("ir", {})
    
    # Generate multiple files in different directories
    output = {
        "files": [
            {
                "path": "types/product.py",
                "content": '''# Generated Product type
class Product:
    def __init__(self, id: str, name: str, price: float):
        self.id = id
        self.name = name
        self.price = price
'''
            },
            {
                "path": "services/product_service.py",
                "content": '''# Generated ProductService
from types.product import Product

class ProductService:
    def get_product(self, id: str) -> Product:
        raise NotImplementedError()
'''
            },
            {
                "path": "index.py",
                "content": '''# Generated index
from types.product import Product
from services.product_service import ProductService

__all__ = ["Product", "ProductService"]
'''
            },
            {
                "path": "metadata.json",
                "content": json.dumps({
                    "generated": True,
                    "types_count": len(ir.get("types", [])),
                    "rpcs_count": len(ir.get("rpcs", []))
                }, indent=2)
            }
        ]
    }
    
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
