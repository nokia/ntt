This document describes overall architecture and design goals of the ntt
project. It's purpose is to improve collaboration by defining common goals and
designs to provide orientation for new contributors.

Please review this file once in every while; and also keep it concise and
short as possible.


# Mission Statement and Design Guidelines

Our mission statement is:

> _to provide modern, trustworthy, free and open TTCN-3 tools, that enable
> users to write, run, analyze and automate tests conveniently and
> efficiently._

* Make ntt work out of the box, by selecting good defaults and automatic
configuration.
* Make ntt seem fast by preferring low latency over high throughput.
* Make ntt sustainable by preferring readability over performance.
* Make ntt easy to use and integrate by providing [good command line
interfaces](https://clig.dev/).
* Write _good_ tests, human lives might depend on your code.
* Improve reuse in the TTCN-3 community by creating interoperability with other
tools and vendors.
* Be excellent to each other.

# Architecture

![ntt-architecture](https://user-images.githubusercontent.com/12729086/192311761-6122eaec-8aa0-49cd-8509-fe645df1a551.svg)
