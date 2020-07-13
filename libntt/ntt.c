#include "ntt.h"
#include "libntt.h"
#include <stdlib.h>
#include <stdio.h>

void path_error() {
       printf("error path is null\n");
       abort();
}



char *  ntt_list_testcases(const char * path) {
    if (path == NULL) {
        path_error();
    }

    return NttListTests(path);
}

char * ntt_list_imports(const char * path) {
    if (path == NULL) {
        path_error();
    }

    return NttListImports(path);
}

char * ntt_load_suite(const char * path) {
    if(path == NULL) {
        path_error();
    }

    return NttLoadSuite(path);
}
