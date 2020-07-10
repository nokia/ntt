#define PY_SSIZE_T_CLEAN
#include <Python.h>
#include <ntt.h>
#include <string.h>

static PyObject *
list_testcases(PyObject *self, PyObject *args)
{
    const char *path;
    PyObject * list = PyList_New(0);

    if (!PyArg_ParseTuple(args, "s", &path))
        return NULL;

    char * tests = ntt_list_testcases(path);

    // empty list case
    if (tests == NULL) {
        return list;
    }

    char * token = strtok(tests, "\n");

    while(token != NULL) {
        PyList_Append(list, Py_BuildValue("s", token));
        token = strtok(NULL, "\n");
    }

    if(tests) {
        free(tests);
    }
    
    return list;
}


struct ModuleImportPair {
    PyObject * module;
    PyObject * import;
};


struct ModuleImportPair split_module_import_string(char * str) {
    struct ModuleImportPair mip;
    mip.module = Py_BuildValue("s", strtok_r(str, "\t", &str));
    mip.import = Py_BuildValue("s", strtok_r(str, "\t", &str));
    return mip;
}


void insert_import_value(char * str, PyObject * dict) {
    char * buffer = strdup(str);
 
    struct ModuleImportPair mip = split_module_import_string(buffer);
    //check if module is a key in dict
    
    int rc = PyDict_Contains(dict, mip.module);
    
    if( rc == 0) {
        PyObject * list = PyList_New(0);

        PyList_Append(list, mip.import); 
        PyDict_SetItem(dict, mip.module, list);
    } else if( rc == 1)  {

        PyObject * list = PyDict_GetItem(dict, mip.module);
        PyList_Append(list, mip.import); 
    } else {
        PySys_WriteStdout("error \n");
    }

    if(buffer) {
       free(buffer);     
    }
}

static PyObject * list_imports(PyObject * self, PyObject * args)
{
    const char * path;
    PyObject * dict = PyDict_New();

    //every token gets an empty list first 
    // then check if dict contains the key -> get list -> append import

    if (!PyArg_ParseTuple(args, "s", &path))
        return NULL;

    char * imports = ntt_list_imports(path);
    // empty dictionary 
    if (imports == NULL) {
        return dict;
    }

    char * token = strtok(imports, "\n");
    while(token != NULL) {
        insert_import_value(token, dict);
        token = strtok(NULL, "\n"); 
    }

    if(imports) {
        free(imports);
    }

    return dict;
}



 /*  define functions in module */
 static PyMethodDef NttMethods[] =
 {
      {"list_testcases", list_testcases, METH_VARARGS, "list ttcn3 testcases"},
      {"list_imports", list_imports, METH_VARARGS, "list ttcn3 imports in testcase"},
      {NULL, NULL, 0, NULL}
 };


static struct PyModuleDef cModPyDem =
 {
     PyModuleDef_HEAD_INIT,
     "ntt", "Some documentation",
     -1,
     NttMethods
};
 
PyMODINIT_FUNC PyInit_ntt(void)
{
     return PyModule_Create(&cModPyDem);
}

