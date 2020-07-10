from distutils.core import setup, Extension

ntt = Extension('ntt',
                         sources = ['pyntt/ntt_python.c'],
                         include_dirs = ['./libntt/'],
                         libraries=['ntt'],
                         library_dirs = ['./libntt/'],
                         )

setup(  name='ntt',
        ext_modules=[ntt]
     )

