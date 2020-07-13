from ntt.suite import Suite

path = "/home/peter/src/ccsstubtest/sourcedir"

suite = Suite(path)

tests = suite.tests()

for test in tests:
    print(test.name())
