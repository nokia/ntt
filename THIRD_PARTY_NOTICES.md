# Third-party notices

ntt builds on, or is inspired by, the following third-party projects.
Each entry retains the upstream copyright notice and license text in
keeping with the relevant license terms.

## Vanadium

ntt's pure-Go ASN.1 frontend (under `internal/asn1/`), its Wadler-style
TTCN-3 formatter combinator layer, its schema-driven TTCN-3 AST
generator (under `ttcn3/v2/syntax/nodes/`), and the `Rule` / `Context`
lifecycle of the new lint engine were inspired by the design of
[Vanadium](https://github.com/makekryl/vanadium) by Mikhail Krylov.

The directly-attributable architectural concepts include, but are not
limited to:

- `Asn1ModuleBasket` (Vanadium) -> `resolver.Basket` (ntt)
- `ClassObjectParser` / `ClassSetResolver` (Vanadium) ->
  `class.WithSyntaxParser` / `class.ObjectSetResolver` (ntt)
- `Asn1AstTransformer` (Vanadium) -> `transform.LowerModule` (ntt)
- `src/ast/nodes.yml` schema-driven AST (Vanadium) ->
  `ttcn3/v2/syntax/nodes` generator (ntt)
- `format::PrintDirective` combinator vocabulary (Vanadium) ->
  Wadler-style document combinator layer in ntt's formatter
- Lint `Rule` with `Register` / `Check` / `Exit` lifecycle (Vanadium) ->
  ntt's rule scaffolding

No verbatim Vanadium source code was copied into ntt - the ports above
are reimplementations in Go - but the design lineage is direct and is
acknowledged here.

```
BSD 3-Clause License

Copyright (c) 2025, Mikhail Krylov
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of the copyright holder nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
