from distutils.core import setup, Extension

ntt = Extension('ntt',
                         sources = ['ntt_python.c'],
                         include_dirs = ['.'],
                         libraries=['ntt'],
                         library_dirs = ['.'],
                         )

setup(ext_modules=[ntt])

