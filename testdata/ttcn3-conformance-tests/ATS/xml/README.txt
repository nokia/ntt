TTCN-3 Part 9 (XML) conformance test suite preliminary (unofficial) instructions
=======================================================

These are preliminary instructions on how to use the TTCN-3 conformance test suite.
Unlike ATSs in the usual context, this ATS does _not_ provide any means for test
automation, but the TTCN-3 files provide the test inputs for the TTCN-3 tool - in this case
the TTCN-3 tool is the IUT. This means that test automation has to be somehow scripted, however,
from the provided ATS it should be clear how the output should be interpreted.

1) ATS organization

   The ATS is organized according to the clauses of the TTCN-3 standard part 1. There are
   either positive syntactic tests, negative syntactic tests, positive semantic tests, or
   negative semantic tests.
   
   Except for the negative syntactic tests, all TTCN-3 files from the other three categories
   are always syntactically correct. In addition, the positive syntactic tests are designed
   to be semantically correct as well. The negative semantic tests are designed to violate the
   semantics only in the one property that is subject of the test - at least to the degree that
   it is possible. Because all test cases are syntactically correct (except for the negative 
   syntactic tests), the semantic and negative semantic tests are syntactically correct as well.
   
   Every TTCN-3 module corresponds to one TTCN-3 conformance test. Where more than one module is
   needed, the TTCN-3 file contains multiple modules.
   
2) Document tags

   Every module is annotated with document tags. Of relevance for the test automation tooling
   is the @verdict tag on top of each module (tag usage adaption in the TTCN-3 maintenance STF
   is pending). It is composed of three parts: the tag, the conformance verdict, and keywords 
   regarding the expected output.
   
   @verdict pass accept, ttcn3verdict:pass
   
   Hence the format of this tag is a three column entry with "@verdict", "pass", and 
   "accept, ttcn3verdict:pass, manual:'Freetext validation'".
   In order to reach the "pass" verdict, the TTCN-3 tool under test, the IUT, must accept the 
   TTCN-3 module as test input. The result of its execution must be the TTCN-3 verdict pass.
   When needed, a manual inspection of the execution results is required.
   Therefore, if the TTCN-3 verdict is pass after the execution of the module, the IUT passes
   the conformance test. If the tool output is anything else, or it does not comply
   with the manual inspection, it fails the conformance test.
   
   The keywords in use to describe the third column, i.e. the expected output by the IUT, are (currently) as 
   follows:
   
   reject
   accept, noexecution
   accept, ttcn3verdict:none
   accept, ttcn3verdict:pass
   accept, ttcn3verdict:inconc
   accept, ttcn3verdict:fail
   accept, ttcn3verdict:error
   accept, ttcn3verdict:xxx, manual:"Validation inspection, free text" (see Note)
   
   Note: 'xxx' for ttcn3verdict is a placeholder for verdict value.  The actual test cases must
   have one of the allowed values 'none', 'pass', 'inconc', 'fail', or 'error' in place of 'xxx'.  

   "reject" implies that the TTCN-3 module is either rejected at compile-time or at execution time.
   In the conformance test, we do not differentiate between these two cases as the standard does not
   make any statement where semantic checks have to be performed.
   
   "accept, noexecution" implies that the TTCN-3 module should be accepted by the TTCN-3 tool after
   the syntactic check. For passing the conformance test, the module simply has to be accepted and
   an execution is not necessary.
    
   "accept, ttcn3vercit:xxx" implies that the TTCN-3 module should be accepted by the TTCN-3 tool
   after the syntactic check and that an execution should take place. The result of the execution
   should be a TTCN-3 verdict and the "xxx" denotes what verdict is the exepcted verdict. If the
   verdict differs from the specified "xxx", the conformance test fails. Otherwise, it passes.
   
   In the usual case, each TTCN-3 file contains only one test case. In these cases the verdict
   determination is clear. In a few cases, the TTCN-3 file contains more than one test case.
   In that case, the overall conformance verdict is determined according to the TTCN-3 verdict
   overwrite rules applied to the results of each test case. Let's say we have two test cases. 
   The first test case ends with the verdict "fail" and the second one ends with the verdict "pass". 
   Then the overall verdict is "fail".
   
   The manual results inspection is a free text describing how a valid out should look like.  The text
   is informal since different tools have different logging formats and facilities.  The instruction 
   is surrounded by single or double quotas on a single line:
   
   @verdict pass, testverdict:pass, manual:'The following elements are logged: charstring "Extra", record { 1, "HELLO"}, integer template "?".  Make sure the elements are logged at setverdict, at MTC end, and at test case end.'
   
   All notations are designed to be easily machine readable. Therefore, the test automation using
   some sort of scripting will only take a small amount of time.
   
3) Other prerequisites

   In order to test communication and matching behavior, we expect that the test cases are executed
   using a loopback adapter.

