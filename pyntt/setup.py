from distutils.core import setup, Extension

ntt_impl = Extension('ntt_impl',
                         sources = ['ntt_python.c'],
                         include_dirs = ['../libntt/'],
                         libraries=['ntt'],
                         library_dirs = ['../libntt/'],
                         )

setup(  name='ntt',
        ext_modules=[ntt_impl],
        packages=['ntt'],
     )

