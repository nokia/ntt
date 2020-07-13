import json
from ntt_impl import load_suite


class Test:

    def __init__(self, data):
        self._tags = data["tags"]
        self._name = data["name"]
        self._module = data["mod"]

    def name(self):
        return self._name

    def tags(self):
        return self._tags

    def module(self):
        return self._module


class Suite:

    def __init__(self, path):
        self.path = path
        self._tests = list()

        self._load()

    def _load(self):
        data = json.loads(load_suite(self.path))
        for e in data:
            self._tests.append(Test(e))

    def tests(self):
        return self._tests
