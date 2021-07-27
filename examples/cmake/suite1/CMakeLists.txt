find_package(NTT REQUIRED)

add_ttcn3_suite(suite1 SOURCES test.ttcn3 DEPENDS ${CMAKE_SOURCE_DIR}/Example)
