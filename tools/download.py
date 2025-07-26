#!/usr/bin/env python3

import json
import os
import re

from urllib.request import Request, urlopen
from copy import deepcopy

SCRIPT_PATH = os.path.dirname(os.path.abspath(__file__))
API_PATH = os.path.join(os.path.dirname(SCRIPT_PATH), "api", "v1")
API_SPEC_LOCATION = "https://api.cloud.denvrdata.dev/swagger/v1/swagger.json"

# A bit of a hack to work around https://github.com/oapi-codegen/oapi-codegen/issues/1856
# Basically, it seems quicker to just filter out components not found in the INCLUDED_PATHS
# as part of our download process.
INCLUDED_PATHS = [
    "/api/v1/clusters/GetAll",
    "/api/v1/servers/images/GetOperatingSystemImages",
    "/api/v1/servers/applications/GetApplications",
    "/api/v1/servers/applications/GetApplicationDetails",
    "/api/v1/servers/applications/GetApplicationRuntimeLogs",
    "/api/v1/servers/applications/GetConfigurations",
    "/api/v1/servers/applications/GetAvailability",
    "/api/v1/servers/applications/GetApplicationCatalogItems",
    "/api/v1/servers/applications/CreateCatalogApplication",
    "/api/v1/servers/applications/CreateCustomApplication",
    "/api/v1/servers/applications/StartApplication",
    "/api/v1/servers/applications/StopApplication",
    "/api/v1/servers/applications/DestroyApplication",
    "/api/v1/servers/metal/GetHosts",
    "/api/v1/servers/metal/GetHost",
    "/api/v1/servers/metal/RebootHost",
    "/api/v1/servers/metal/ReprovisionHost",
    "/api/v1/servers/virtual/GetServers",
    "/api/v1/servers/virtual/GetServer",
    "/api/v1/servers/virtual/CreateServer",
    "/api/v1/servers/virtual/StartServer",
    "/api/v1/servers/virtual/StopServer",
    "/api/v1/servers/virtual/DestroyServer",
    "/api/v1/servers/virtual/GetConfigurations",
    "/api/v1/servers/virtual/GetAvailability",
    "/api/v1/servers/virtual/GetVirtualMachineBootLogs",
]

def paths(spec):
    """Returns a dict of the filtered paths with an operationId set."""
    results = {}
    for path, ops in spec["paths"].items():
        if path in INCLUDED_PATHS:
            results[path] = ops

            assert len(ops) == 1
            op = next(iter(ops.keys()))
            results[path][op]["operationId"] = os.path.split(path)[-1]

    return results

def schemas(spec, init):
    """Returns a dict of all schemas recursively referenced by the init objects"""
    found = set()
    queued = set()
    results = {}

    def add(string):
        """Adds a potential reference to the set of all and queued refs if not already present"""
        match = re.match(r"#/components/schemas/([^/]+)", string)
        if match and match.group(1) not in found:
            found.add(match.group(1))
            queued.add(match.group(1))

    def populate(item):
        """Add $ref values or recurse on lists or dicts"""
        if isinstance(item, dict) and "$ref" in item:
            add(item["$ref"])
        elif isinstance(item, (list, dict)):
            for value in item if isinstance(item, list) else item.values():
                populate(value)

    populate(init)

    while queued:
        ref = queued.pop()
        if ref in spec.get("components", {}).get("schemas", {}):
            results[ref] = spec["components"]["schemas"][ref]
            populate(results[ref])

    return results

if __name__ == "__main__":
    request = Request(API_SPEC_LOCATION, headers={"User-Agent": "Mozilla"})

    with urlopen(request) as robj:
        original = json.load(robj)
        results = deepcopy(original)
        results["paths"] = paths(original)
        results["components"]["schemas"] = schemas(original, results["paths"])

        with open(os.path.join(API_PATH, "api.json"), "w") as wobj:
            json.dump(results, wobj, indent=2)
