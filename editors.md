---
nav_order: 2
---

# IDE Support for TTCN-3
ntt gives your editor advanced language features like auto complete, go to
definition, find all references etc.
{: .fs-6 .fw-300 }


**ntt** implements
the [language server protocol](https://microsoft.github.io/language-server-protocol).
This makes ntt a universal TTCN-3 language plugin for virtually any editor, like:

* [Visual Studio Code](#visual-studio-code)
* [Vim](#vim-8)
* ...

_If you are using ntt with an editor not listed here please share your configuration with other users and create a pull
request in GitHub against [this
markdown](https://github.com/nokia/ntt/blob/gh-pages/editors.md) document._


For most editors or IDEs you have to install ntt manually at a place where your
editor can find it. Check out the ntt [install
section](https://github.com/nokia/ntt#install) for details.

<!-- This table is disabled until we have some features finished and something to show:

## Features

| Name                        | Method                            |                    |
| --------------------------- | --------------------------------- | ------------------ |
| Workspace Symbols           | `workspace/symbol`                | :x:                |
| Execute Command             | `workspace/executeCommand`        | :x:                |
| Diagnostics                 | `textDocument/publishDiagnostics` | :x:                |
| Completion                  | `textDocument/completion`         | :x:                |
| Hover                       | `textDocument/hover`              | :x:                |
| Signature Help              | `textDocument/signatureHelp`      | :x:                |
| Goto Definition             | `textDocument/definition`         | :heavy_check_mark: |
| Goto Type Definition        | `textDocument/typeDefinition`     | :x:                |
| Goto Implementation         | `textDocument/implementation`     | :x:                |
| Find References             | `textDocument/references`         | :x:                |
| Document Highlights         | `textDocument/documentHighlight`  | :x:                |
| Document Symbols            | `textDocument/documentSymbol`     | :x:                |
| Code Action                 | `textDocument/codeAction`         | :x:                |
| Code Lens                   | `textDocument/codeLens`           | :x:                |
| Document Formatting         | `textDocument/formatting`         | :x:                |
| Document Range Formatting   | `textDocument/rangeFormatting`    | :x:                |
| Document on Type Formatting | `textDocument/onTypeFormatting`   | :x:                |
| Rename                      | `textDocument/rename`             | :x:                |
-->


**Troubleshooting: Go to Definition does not work**


Go to Definition works only for known TTCN-3 modules. Yet, there is no standard
way of telling the language server where to look for TTCN-3 modules.

You should always open whole folders (`Open > Folder`) and not just
single files (`Open > File`). The language server automatically recognizes all
TTCN-3 from that opened folder.
If your TTCN-3 test suite is organized across multiple folders, a [test suite manifest
file](getting-started#the-test-suite-manifest)
should be in the test suite's root folder.
When you open multiple folders (workspace), the first folder is considered the
test suite root folder.

The command "TTCN-3: Show language server status" shows a list of all known
TTCN-3 modules and will help to verify that all relevant modules are known by
the language server

If you do not open the right folders, very little will work. This is the most
common issue that we see. If you have suggestions on how to improve this, we'd
love to hear from you.


The language server is still work in progress. Not all TTCN-3 semantic is
implemented already, therefore some TTCN-3 structure might not work as
expected:

* _Component Variables_
* _Signatures_
* _Dot Expressions_ (such as `foo.bar`)
* _Ambiguous Module Definitions_ (only the first hit is returned)



## Visual Studio Code

Install the [TTCN-3 extension](https://marketplace.visualstudio.com/items?itemName=nokia.ttcn3)
for VS Code from the Visual Studio Marketplace. For additional details on
installing extensions, see [Extension Marketplace](https://code.visualstudio.com/docs/editor/extension-gallery).
The TTCN-3 extension is named TTCN-3 and it's published by Nokia:

![vscode TTCN-3 extension preview](/static/vscode-ttcn3-extension-preview.png)

**ntt** is still in beta and therefore disabled by default. Enable it by
opening [vscode settings](https://code.visualstudio.com/docs/getstarted/settings) and set
`ttcn3.useLanguageServer` to `true`.

If you use an older version of this extension, ntt won't be installed
automatically. Either you [install ntt manually](https://github.com/nokia/ntt#install)
or you update the TTCN-3 extension.

## Vim 8

Example configuration using [vim-plug](https://github.com/junegunn/vim-plug) and [vim-lsp](https://github.com/prabirshrestha/vim-lsp):
```vim
call plug#begin('~/.vim/plugged')

" Language Server Protocol Support
Plug 'prabirshrestha/async.vim' " Required for vim-lsp
Plug 'prabirshrestha/vim-lsp'   " Generic Language Protocol client
Plug 'mattn/vim-lsp-settings'   " Automatically install and configure language servers"

" For TTCN-3 syntax highlighting and to trigger vim-lsp-settings
Plug 'gustafj/vim-ttcn'

call plug#end()
```

Execute `:PlugInstall` to download and install all plugins. When you open a `.ttcn3` source file syntax highlighting should work already and you will be prompted to execute `:LspInstallServer` for installing ntt.

Further [vim-lsp](https://github.com/prabirshrestha/vim-lsp) recommends to map keys for your convenience:
```vim
function! s:on_lsp_buffer_enabled() abort
    setlocal omnifunc=lsp#complete
    setlocal signcolumn=yes
    if exists('+tagfunc') | setlocal tagfunc=lsp#tagfunc | endif
    nmap <buffer> gd <plug>(lsp-definition)
    nmap <buffer> gs <plug>(lsp-document-symbol-search)
    nmap <buffer> gS <plug>(lsp-workspace-symbol-search)
    nmap <buffer> gr <plug>(lsp-references)
    nmap <buffer> gi <plug>(lsp-implementation)
    nmap <buffer> gt <plug>(lsp-type-definition)
    nmap <buffer> <leader>rn <plug>(lsp-rename)
    nmap <buffer> [g <Plug>(lsp-previous-diagnostic)
    nmap <buffer> ]g <Plug>(lsp-next-diagnostic)
    nmap <buffer> K <plug>(lsp-hover)

    let g:lsp_format_sync_timeout = 1000
    autocmd! BufWritePre *.rs,*.go call execute('LspDocumentFormatSync')

    " refer to doc to add more commands
endfunction

augroup lsp_install
    au!
    " call s:on_lsp_buffer_enabled only for languages that has the server registered.
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

## Emacs

First of all make sure ntt is [manually installed](https://github.com/nokia/ntt#install).

Install [lsp-mode](https://github.com/emacs-lsp/lsp-mode) from MELPA/ELPA and configure TTCN-3 support:

```elisp
(use-package lsp-mode
  :hook ((ttcn-3-mode . lsp-deferred))
  :config
  (add-to-list 'lsp-language-id-configuration
    '(".*\\.ttcn3$" . "ttcn3"))
  (lsp-register-client
    (make-lsp-client :new-connection (lsp-stdio-connection '("ntt" "langserver"))
                     :activation-fn (lsp-activate-on "ttcn3")
                     :server-id 'ntt)))
```
