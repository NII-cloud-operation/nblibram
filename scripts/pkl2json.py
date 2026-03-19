#!/usr/bin/env python3
"""Convert a pickle file (.pkl) to JSON on stdout.

Usage: pkl2json.py <file.pkl>
       pkl2json.py < file.pkl
"""
import json
import pickle
import sys


def default_serializer(obj):
    return repr(obj)


def main():
    if len(sys.argv) > 1 and sys.argv[1] != "-":
        with open(sys.argv[1], "rb") as f:
            obj = pickle.load(f)
    else:
        obj = pickle.load(sys.stdin.buffer)

    json.dump(obj, sys.stdout, ensure_ascii=False, default=default_serializer)


if __name__ == "__main__":
    main()
