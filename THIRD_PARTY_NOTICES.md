# Third-party components

Primary pinned components:

| Component | Version | License |
| --- | --- | --- |
| [MyGo](https://github.com/egoist/mygo) | v0.4.3 | MIT |
| [anacrolix/torrent](https://github.com/anacrolix/torrent) | v1.61.0 | MPL-2.0 |

The complete resolved Go dependency graph is recorded in `go.mod` and `go.sum`. This document is a primary-component inventory, not a complete distribution notice bundle. Preserve upstream notices and prepare the transitive license inventory before binary distribution.

MyGo MIT notice:

```text
MIT License

Copyright (c) 2026 MyGo contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```

The engine's MPL-2.0 text is retained in [docs/licenses/anacrolix-torrent.txt](docs/licenses/anacrolix-torrent.txt). Engine code is consumed as an unmodified Go module.
