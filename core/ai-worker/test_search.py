#!/usr/bin/env python3
"""
Test script to verify DuckDuckGo search functionality
"""

import time
import requests
from duckduckgo_search import DDGS

def test_duckduckgo_direct():
    """Test DuckDuckGo search directly"""
    print("Testing DuckDuckGo search directly...")
    
    try:
        with DDGS() as ddgs:
            results = list(ddgs.text("python programming", max_results=3))
            
        print(f"Found {len(results)} results:")
        for i, result in enumerate(results, 1):
            print(f"{i}. {result.get('title', 'No title')}")
            print(f"   {result.get('body', 'No description')[:100]}...")
            print(f"   URL: {result.get('href', 'No URL')}")
            print()
            
    except Exception as e:
        print(f"Error: {e}")

def test_simple_web_request():
    """Test simple web request to verify internet connectivity"""
    print("Testing internet connectivity...")
    
    try:
        response = requests.get("https://httpbin.org/json", timeout=10)
        print(f"Status: {response.status_code}")
        print(f"Response: {response.json()}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_simple_web_request()
    print("-" * 50)
    test_duckduckgo_direct()
