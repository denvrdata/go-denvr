#!/usr/bin/env python3

import json
import os

from urllib.request import Request, urlopen

SCRIPT_PATH = os.path.dirname(os.path.abspath(__file__))
API_PATH = os.path.join(os.path.dirname(SCRIPT_PATH), "api", "v1")
API_SPEC_LOCATION = "https://api.cloud.denvrdata.dev/swagger/v1/swagger.json"

if __name__ == "__main__":
    request = Request(API_SPEC_LOCATION, headers={"User-Agent": "Mozilla"})

    with urlopen(request) as robj:
        d = json.load(robj)
        for path, ops in d["paths"].items():
            assert len(ops) == 1
            op = next(iter(ops.keys()))
            d["paths"][path][op]["operationId"] = os.path.split(path)[-1]

        with open(os.path.join(API_PATH, "api.json"), "w") as wobj:
            json.dump(d, wobj, indent=2)

