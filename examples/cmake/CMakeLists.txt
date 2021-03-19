cmake_minimum_required(VERSION 3.0)
project(example)

# Tell CMake we to look for FindNTT.cmake
list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/../../cmake")

find_package(NTT REQUIRED)
add_subdirectory(suite1)
add_subdirectory(suite2)
