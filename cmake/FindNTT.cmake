# BSD 3-Clause License
#
# Copyright (c) 2020, Nokia
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
#
# * Redistributions of source code must retain the above copyright notice, this
#   list of conditions and the following disclaimer.
#
# * Redistributions in binary form must reproduce the above copyright notice,
#   this list of conditions and the following disclaimer in the documentation
#   and/or other materials provided with the distribution.
#
# * Neither the name of the copyright holder nor the names of its
#   contributors may be used to endorse or promote products derived from
#   this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
# FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
# DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
# SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
# CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
# OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#

#[====================================================================[.rst:
NTT Tests
--------

This module provides functions to help use the NTT/K3 Test infrastructure. It
provides function :command:`add_ttcn3_suite` for generating a test suite
manifest.

.. command:: add_ttnc3_suite

  Add a TTCN-3 test suite manifest for use with NTT and CTest::

  add_ttcn3_suite(TGT
    SOURCES src1...
    [DEPENDS ...]
    [NAME name]
    [TIMEOUT secs]
    [TEST_HOOK executable]
    [PARAMETER_FILE file]
    [WORKING_DIRECTORY dir]
    [TARGETS target1...]
  )

  ``add_ttcn3_suite`` creates a target TGT and test suite manifest
  ``package.yml`` as ``BYPRODUCT``.

  The options are:

  ``SOURCES src1...``
    The list of .ttcn3 source files. Files specified in SOURCES usually are
    testcases.

  ``DEPENDS target1...|directory...
    Specifies additional TTCN-3 packages required by the test suite.
    If the argument specifies a target it will be replaced by the location of
    the target directory (``$<TARGET_FILE_DIR>``). Additionally a target-level
    dependency will be added so that the depending target will be built before
    this test suite.

  ``NAME``
    Specfies the name of the test suite. If this option is not provided, ntt
    will assign one.

  ``TIMEOUT seconds``
    Specifies a timeout after which a test case will be aborted.

  ``TEST_HOOK executable``
    Specifies an executable to be used as test hook.
    If TEST_HOOK specifies an executable target (created by ADD_EXECUTABLE) it
    will automatically be replaced by the location of the executable created at
    build time. Additionally a target-level dependency will be added so that
    the executable target will be built before this hook is used.

  ``PARAMETER_FILE file``
    Specifies a file containing TOML formatted test configuration.

  ``WORKING_DIRECTORY dir``
    Specifies the directory in which to run the tests. If this option is not
    provided, the current binary directory is used.

  ``TARGETS target1...``
    Add additional target-level dependencies. This is used to assure SUTs are
    built before a test is executed.

]====================================================================]

if (NOT NTT_ROOT)
  set(NTT_ROOT ${CMAKE_SOURCE_DIR}/lib/ntt)
endif()

find_program(NTT_EXECUTABLE NAMES ntt k3 PATHS ${NTT_ROOT}/bin DOC "Path to NTT")

if(NTT_EXECUTABLE)
  execute_process(
    COMMAND ${NTT_EXECUTABLE} version
    OUTPUT_VARIABLE NTT_version_output
    ERROR_VARIABLE NTT_version_error
    RESULT_VARIABLE NTT_version_result
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )
  string(REGEX REPLACE "[^ ]+ ([0-9\.]+).*$" "\\1" NTT_VERSION "${NTT_version_output}")

  if(NOT TARGET ntt:ntt)
    add_executable(ntt::ntt IMPORTED)
    if(EXISTS "$NTT_EXECUTABLE")
      set_property(TARGET ntt:ntt PROPERTIES IMPORTED_LOCATION "${NTT_EXECUTABLE}")
    endif()
  endif()
endif()

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(NTT
  FOUND_VAR     NTT_FOUND
  REQUIRED_VARS NTT_EXECUTABLE
  VERSION_VAR   NTT_VERSION
)

mark_as_advanced(NTT_EXECUTABLE)

function(add_ttcn3_suite TGT)
  set("ARGS_PREFIX" "")
  set("ARGS_OPTIONS" "")
  set("ARGS_ONE_VALUE" "NAME;TIMEOUT;TEST_HOOK;PARAMETERS_FILE;WORKING_DIRECTORY")
  set("ARGS_MULTI_VALUE" "SOURCES;DEPENDS;TARGETS")
  cmake_parse_arguments("${ARGS_PREFIX}" "${ARGS_OPTIONS}" "${ARGS_ONE_VALUE}" "${ARGS_MULTI_VALUE}" ${ARGN})

  if (NOT _WORKING_DIRECTORY)
    set(_WORKING_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}")
  endif()
  file(MAKE_DIRECTORY "${_WORKING_DIRECTORY}")

  set(MANIFEST "${_WORKING_DIRECTORY}/package.yml")

  add_custom_target("${TGT}"
    BYPRODUCTS "${MANIFEST}"
    COMMENT "Generate manifest ${MANIFEST}"
  )

  # Write package.yml header
  manifest_write("# DO NOT MODIFY.")
  manifest_append("# This file was generated by ${CMAKE_CURRENT_LIST_FILE}:${CMAKE_CURRENT_LIST_LINE}")

  if (_NAME)
    manifest_append("name: ${_NAME}")
  endif()

  if (_TIMEOUT)
    manifest_append("timeout: ${_TIMEOUT}")
  endif()

  if (_TEST_HOOK)
    if(TARGET "${_TEST_HOOK}")
      manifest_append("test_hook: $<TARGET_FILE:${_TEST_HOOK}>")
      add_dependencies("${TGT}" "${_TEST_HOOK}")
    else()
      manifest_append("test_hook: ${_TEST_HOOK}")
    endif()
  endif()

  if (_PARAMETERS_FILE)
    manifest_append("parameters_file: ${_PARAMETERS_FILE}")
  endif()

  # Append sources
  manifest_append("sources:")
  foreach(src ${_SOURCES})
    get_filename_component(abs ${src} ABSOLUTE)
    manifest_append("  - ${abs}")
  endforeach()

  manifest_append("imports:")
  foreach(x ${_DEPENDS})
    if (TARGET "${x}")
      manifest_append("  - $<TARGET_FILE_DIR:${x}>")
      add_dependencies("${TGT}" "${x}")
    else()
      get_filename_component(abs ${x} ABSOLUTE)
      if(IS_DIRECTORY "${abs}")
        manifest_append("  - ${abs}")
      elseif(COMMAND "${x}")
        message(FATAL_ERROR "Command \"${x}\" does not define an existing CMake target or directory")
      else()
        message(FATAL_ERROR "\"${x}\" does not define an existing CMake target or directory")
      endif()
    endif()
  endforeach()

  foreach(x ${_TARGETS})
      add_dependencies("${TGT}" "${x}")
  endforeach()
endfunction()

macro(manifest_write LINE)
  file(WRITE "${MANIFEST}" "${LINE}\n")
endmacro()

macro(manifest_append LINE)
  file(APPEND "${MANIFEST}" "${LINE}\n")
endmacro()

