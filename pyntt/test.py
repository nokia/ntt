import ntt

path = "/home/peter/ccsstubtest"

testcases = ntt.list_testcases(path)
print(testcases)
for tc in testcases:
    print(tc)

imports = ntt.list_imports(path)
print(imports["test"])

for imp in imports["test"]:
    print(imp)


print(type(imports))
#print(ntt.list_imports(path))
#print(ntt.list_testcases(path + "/logs"))
