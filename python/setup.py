from distutils.core import setup, Extension

ntt = Extension('ntt',
                         sources = ['python/ntt_python.c'],
                         include_dirs = ['./libntt/'],
                         libraries=['ntt'],
                         library_dirs = ['./libntt/'],
                         )

setup(ext_modules=[ntt])

